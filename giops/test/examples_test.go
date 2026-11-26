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

package test

import (
	"path/filepath"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func example(t *testing.T, name string) string {
	path, err := filepath.Abs(filepath.Join("..", "examples", name))
	require.NoError(t, err)
	return path
}

func TestExampleConfig(t *testing.T) {
	c, err := config.Load(example(t, "config/config.yaml"))
	require.NoError(t, err)

	am, err := c.AMContext()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8093/automation", am.BaseURL)
	assert.NotNil(t, am.Auth.BearerToken)

	apim, err := c.APIMContext()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8083/automation", apim.BaseURL)
	assert.NotNil(t, apim.Auth.BasicAuth)
}

func TestExampleAMResources(t *testing.T) {
	requireMock(t, "the examples are not giops-* resources the live tests clean up")
	giops := newAMCLI(t)

	giops.mustRun("am", "apply", "domains", "-f", example(t, "am/domain.yaml"), "--yes")
	giops.mustRun("am", "apply", "identityproviders", "-f", example(t, "am/identity-provider.yaml"), "--yes")
	for _, apply := range [][]string{
		{"reporters", "am/reporter.yaml"},
		{"certificates", "am/certificate.yaml"},
		{"applications", "am/application.yaml"},
		{"secrets", "am/secret.yaml"},
		{"dataplanes", "am/dataplane.yaml"},
	} {
		giops.mustRun("am", "apply", apply[0], "-f", example(t, apply[1]), "--yes")
	}

	domain := giops.getJSON("am", "domains", "example-domain")
	assert.Equal(t, "Example domain", domain["name"])
	out := giops.mustRun("am", "get", "identityproviders", "--domain", "example-domain")
	assert.Regexp(t, `example-users\s+inline-am-idp`, out)
	inDomain := []string{"--domain", "example-domain"}
	assert.Equal(t, "Example audit file", giops.getJSON("am", "reporters", "example-reporter", inDomain...)["name"])
	assert.Equal(t, "Example signing certificate", giops.getJSON("am", "certificates", "example-certificate", inDomain...)["name"])
	assert.Equal(t, "Example service", giops.getJSON("am", "applications", "example-app", inDomain...)["name"])
	assert.Equal(t, "example-secret", giops.getJSON("am", "secrets", "example-secret", "--domain", "example-domain", "--application", "example-app")["key"])
	assert.Equal(t, "Example data plane", giops.getJSON("am", "dataplanes", "example-dataplane")["name"])
}

func TestExampleAPIMResources(t *testing.T) {
	requireMock(t, "the examples are not giops-* resources the live tests clean up")
	giops := newAPIMCLI(t)

	for _, apply := range [][]string{
		{"groups", "apim/group.yaml"},
		{"apis", "apim/api.yaml"},
		{"documentations", "apim/api-documentation.yaml"},
		{"links", "apim/api-link.yaml"},
		{"applications", "apim/application.yaml"},
		{"subscriptions", "apim/subscription.yaml"},
		{"dictionaries", "apim/dictionary.yaml"},
		{"shared-policy-groups", "apim/shared-policy-group.yaml"},
	} {
		giops.mustRun("apim", "apply", apply[0], "-f", example(t, apply[1]), "--yes")
	}

	group := giops.getJSON("apim", "groups", "example-developers")
	assert.Equal(t, "Example developers", group["name"])
	onApi := []string{"--api", "example-api"}
	assert.Equal(t, "Example API", giops.getJSON("apim", "apis", "example-api")["name"])
	assert.Equal(t, "Getting started", giops.getJSON("apim", "documentations", "example-api-doc", onApi...)["name"])
	assert.Equal(t, "Status page", giops.getJSON("apim", "links", "example-api-link", onApi...)["name"])
	assert.Equal(t, "Example app", giops.getJSON("apim", "applications", "example-app")["name"])
	assert.Equal(t, "example-plan", giops.getJSON("apim", "subscriptions", "example-subscription", onApi...)["planHrid"])
	assert.Equal(t, "Example dictionary", giops.getJSON("apim", "dictionaries", "example-dictionary")["name"])
	assert.Equal(t, "Example headers", giops.getJSON("apim", "shared-policy-groups", "example-spg")["name"])
}

func TestExamplePortalResources(t *testing.T) {
	requireMock(t, "the examples are not giops-* resources the live tests clean up")
	giops := newGammaCLI(t, "portal")
	giops.mustRun("apim", "apply", "apis", "-f", example(t, "apim/api.yaml"), "--yes")

	for _, apply := range [][]string{
		{"themes", "portal/theme.yaml"},
		{"portals", "portal/portal.yaml"},
		{"documentations", "portal/documentation.yaml"},
		{"links", "portal/link.yaml"},
		{"listings", "portal/listing.yaml"},
	} {
		giops.mustRun("portal", "apply", apply[0], "-f", example(t, apply[1]), "--yes")
	}

	inPortal := []string{"--portal", "example-portal"}
	assert.Equal(t, "Example theme", giops.getJSON("portal", "themes", "example-theme")["name"])
	assert.Equal(t, "Example portal", giops.getJSON("portal", "portals", "example-portal")["name"])
	assert.Equal(t, "Getting started", giops.getJSON("portal", "documentations", "example-guide", inPortal...)["name"])
	assert.Equal(t, "Status page", giops.getJSON("portal", "links", "example-status", inPortal...)["name"])
	assert.Equal(t, "example-listing", giops.getJSON("portal", "listings", "example-listing", inPortal...)["hrid"])
}

func TestExampleAIMResources(t *testing.T) {
	requireMock(t, "the examples are not giops-* resources the live tests clean up")
	giops := newGammaCLI(t, "aim")

	giops.mustRun("aim", "apply", "mcp-servers", "-f", example(t, "aim/mcp-server.yaml"), "--yes")
	giops.mustRun("aim", "apply", "mcp-proxies", "-f", example(t, "aim/mcp-proxy.yaml"), "--yes")

	assert.Equal(t, "mcp-server.example", giops.getJSON("aim", "mcp-servers", "example-mcp-server")["entityId"])
	assert.Equal(t, "Example MCP proxy", giops.getJSON("aim", "mcp-proxies", "example-mcp-proxy")["name"])
}
