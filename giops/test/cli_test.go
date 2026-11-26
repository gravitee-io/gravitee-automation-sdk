// Copyright (C) 2015 The Gravitee team (http://gravitee.io)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package test runs the giops CLI end to end against the AM and APIM mock servers.
//
// To run them against live servers instead, set GIOPS_LIVE_CONFIG to an giops config file:
//
//	GIOPS_LIVE_CONFIG=/path/to/config.yaml go test ./test/...
//
// Each module without a section in that file is skipped, as are the tests that need a mock.
// The tests create resources named giops-*, and delete them before and after each test.
package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	ammock "github.com/gravitee-io/gravitee-automation-sdk/am-mock-server/server"
	apimmock "github.com/gravitee-io/gravitee-automation-sdk/apim-mock-server/server"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/apicontext"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/cmd"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/config"
	"github.com/stretchr/testify/require"
)

// TestMain drops GIOPS_AM_* and GIOPS_GAMMA_* variables, which would override the config files of the tests.
func TestMain(m *testing.M) {
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "GIOPS_AM_") || strings.HasPrefix(name, "GIOPS_GAMMA_") {
			_ = os.Unsetenv(name)
		}
	}
	m.Run()
}

// liveConfig is the config file naming the live servers, or empty to run against mock servers.
var liveConfig = os.Getenv("GIOPS_LIVE_CONFIG")

// liveResources are what the tests create, as delete arguments. On a live server they are
// deleted before each test, for leftovers of a failed run, and after it.
var liveResources = map[string][][]string{
	"am": {
		{"secrets", "giops-secret", "--domain", "giops-domain", "--application", "giops-app"},
		{"secrets", "giops-secret", "--domain", "giops-domain", "--application", "giops-other-app"},
		{"applications", "giops-app", "--domain", "giops-domain"},
		{"applications", "giops-other-app", "--domain", "giops-domain"},
		{"reporters", "giops-reporter", "--domain", "giops-domain"},
		{"certificates", "giops-certificate", "--domain", "giops-domain"},
		{"identityproviders", "giops-idp", "--domain", "giops-domain"},
		{"domains", "giops-domain"},
		{"domains", "giops-other-domain"},
		{"domains", "giops-json-domain"},
		{"dataplanes", "giops-dataplane"},
	},
	"apim": {
		{"subscriptions", "giops-subscription", "--api", "giops-api"},
		{"subscriptions", "giops-subscription", "--api", "giops-other-api"},
		{"documentations", "giops-doc", "--api", "giops-api"},
		{"links", "giops-link", "--api", "giops-api"},
		{"apis", "giops-api"},
		{"apis", "giops-other-api"},
		{"applications", "giops-app"},
		{"dictionaries", "giops-dictionary"},
		{"shared-policy-groups", "giops-spg"},
		{"groups", "giops-group"},
	},
	"portal": {
		{"listings", "giops-listing", "--portal", "giops-portal"},
		{"documentations", "giops-guide", "--portal", "giops-portal"},
		{"documentations", "giops-guide", "--portal", "giops-other-portal"},
		{"links", "giops-link", "--portal", "giops-portal"},
		{"portals", "giops-portal"},
		{"portals", "giops-other-portal"},
		{"themes", "giops-theme"},
	},
	"aim": {
		{"mcp-proxies", "giops-mcp-proxy"},
		{"mcp-servers", "giops-mcp-server"},
	},
	"authz": {
		{"policies", "giops-policy"},
		{"entities", "giops-entity"},
		{"schemas", "giops-schema"},
		{"pdps", "giops-pdp"},
	},
}

// requireMock skips the test on live servers.
func requireMock(t *testing.T, why string) {
	t.Helper()
	if liveConfig != "" {
		t.Skip("needs a mock server: " + why)
	}
}

// result is what a user sees: the output, and the error main prints before exiting with 1.
type result struct {
	out string
	err error
}

// cli runs giops commands with a config file pointing at one server, mock or live.
type cli struct {
	t        *testing.T
	config   string
	api      apicontext.APIContext // the server the config file points at, and its auth
	requests *recorder
}

// recorder keeps the method and URL of every request the server receives.
type recorder struct {
	mu   sync.Mutex
	urls map[string][]*url.URL
}

func record(next http.Handler) (*recorder, http.Handler) {
	rec := &recorder{urls: map[string][]*url.URL{}}
	return rec, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.mu.Lock()
		rec.urls[r.Method] = append(rec.urls[r.Method], r.URL)
		rec.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

// queries returns the query of every request sent with method, in order.
func (c cli) queries(method string) []url.Values {
	c.requests.mu.Lock()
	defer c.requests.mu.Unlock()
	queries := make([]url.Values, len(c.requests.urls[method]))
	for i, u := range c.requests.urls[method] {
		queries[i] = u.Query()
	}
	return queries
}

func newAMCLI(t *testing.T) cli {
	return startAM(t, false)
}

// newRejectingAMCLI answers every dry run with an error, as AM does when a resource is invalid.
func newRejectingAMCLI(t *testing.T) cli {
	requireMock(t, "only the mock rejects every dry run")
	return startAM(t, true)
}

func startAM(t *testing.T, rejectDryRun bool) cli {
	if liveConfig != "" {
		return newLiveCLI(t, "am")
	}
	return newCLI(t, "am", ammock.NewWithPath(ammock.NewMockAM(), ammock.BasePath, nil, rejectDryRun), ammock.BasePath, mockAPI())
}

func newAPIMCLI(t *testing.T) cli {
	return newGammaCLI(t, "apim")
}

// newGammaCLI points a module that uses the APIM client (apim, portal, aim, authz) at the APIM server.
func newGammaCLI(t *testing.T, module string) cli {
	if liveConfig != "" {
		return newLiveCLI(t, module)
	}
	return newCLI(t, module, apimmock.New(apimmock.NewMockAPIM(false)), apimmock.BasePath, mockAPI())
}

func mockAPI() apicontext.APIContext {
	return apicontext.APIContext{Auth: apicontext.Auth{BearerToken: new("test-token")}}
}

// newLiveCLI points the module at its live server, through a proxy that records the requests.
func newLiveCLI(t *testing.T, module string) cli {
	live, err := config.Load(liveConfig)
	require.NoError(t, err, "cannot load GIOPS_LIVE_CONFIG")
	api, ok := liveAPI(live, module)
	if !ok {
		t.Skipf("no section for module %s in %s", module, liveConfig)
	}
	server, err := url.Parse(api.BaseURL)
	require.NoError(t, err)
	proxy := &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
		r.SetURL(&url.URL{Scheme: server.Scheme, Host: server.Host})
	}}

	c := newCLI(t, module, proxy, server.Path, api)
	c.deleteTestResources(module)
	t.Cleanup(func() { c.deleteTestResources(module) })
	return c
}

func liveAPI(live config.CLIConfig, module string) (apicontext.APIContext, bool) {
	if module == "am" {
		api, err := live.AMContext()
		return api, err == nil
	}
	api, err := live.APIMContext()
	return api.APIContext, err == nil
}

// deleteTestResources deletes what the tests create, ignoring what does not exist.
func (c cli) deleteTestResources(module string) {
	for _, resource := range liveResources[module] {
		c.run(append([]string{module, "delete", "-y"}, resource...)...)
	}
}

// newCLI serves the server through a recorder, and writes a config file that points the module at it.
func newCLI(t *testing.T, module string, server http.Handler, basePath string, api apicontext.APIContext) cli {
	requests, handler := record(server)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	api.BaseURL = srv.URL + basePath

	shared := config.Context{
		OrgID: api.OrgID,
		EnvID: api.EnvID,
		Auth:  config.Auth{BearerToken: api.Auth.BearerToken},
	}
	if basic := api.Auth.BasicAuth; basic != nil {
		shared.Auth.BasicAuth = &config.BasicAuth{Username: basic.Username, Password: basic.Password}
	}
	var content config.CLIConfig
	if module == "am" {
		content.AM = &config.AM{BaseURL: api.BaseURL, Context: shared}
	} else {
		content.Gamma = &config.APIM{BaseURL: api.BaseURL, Context: shared}
	}
	file := filepath.Join(t.TempDir(), "config.json")
	data, err := json.Marshal(content)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(file, data, 0o600))
	return cli{t: t, config: file, api: api, requests: requests}
}

func (c cli) run(args ...string) result {
	return c.runWithInput("", args...)
}

// runWithInput runs giops with input as stdin, to answer prompts.
func (c cli) runWithInput(input string, args ...string) result {
	var out bytes.Buffer
	giops := cmd.MainCommand()
	giops.SetArgs(append(args, "--config", c.config))
	giops.SetIn(strings.NewReader(input))
	giops.SetOut(&out)
	giops.SetErr(&out)
	err := giops.ExecuteContext(c.t.Context())
	return result{out: out.String(), err: err}
}

// mustRun fails the test unless the command succeeds.
func (c cli) mustRun(args ...string) string {
	c.t.Helper()
	r := c.run(args...)
	require.NoErrorf(c.t, r.err, "giops %v failed:\n%s", args, r.out)
	return r.out
}

// getJSON reads one resource with `get -o json` and decodes it. flags are added to the command, such as parents.
func (c cli) getJSON(module, subject, id string, flags ...string) map[string]any {
	c.t.Helper()
	out := c.mustRun(append([]string{module, "get", subject, id, "-o", "json"}, flags...)...)
	var resource map[string]any
	require.NoErrorf(c.t, json.Unmarshal([]byte(out), &resource), "not JSON:\n%s", out)
	return resource
}

// requireAbsent fails the test if the resource can be read. flags are added to the command, such as parents.
func (c cli) requireAbsent(module, subject, id string, flags ...string) {
	c.t.Helper()
	r := c.run(append([]string{module, "get", subject, id, "-o", "json"}, flags...)...)
	require.Errorf(c.t, r.err, "%s %s should not exist, got:\n%s", subject, id, r.out)
}

func TestMissingModuleSection(t *testing.T) {
	r := newAMCLI(t).run("apim", "check")

	require.ErrorContains(t, r.err, "no 'gamma' section in the configuration")
}

func TestMissingConfigFile(t *testing.T) {
	giops := newAMCLI(t)
	giops.config = filepath.Join(t.TempDir(), "absent.yaml")

	r := giops.run("am", "check")

	require.ErrorContains(t, r.err, "was not loaded")
}

func testdata(t *testing.T, name string) string {
	path, err := filepath.Abs(filepath.Join("testdata", name))
	require.NoError(t, err)
	return path
}
