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

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// The whole chain without APIM: an Automation API stub answers the two list operations with
// Console-style states (no hrid, two APIs named "Petstore"); gops goes through the generated
// SDK client, settles the names and writes GKO manifests.
func stubAutomation(t *testing.T) string {
	t.Helper()
	fixtures := map[string]string{
		"/automation/organizations/DEFAULT/environments/DEFAULT/apis":            "apis.json",
		"/automation/organizations/DEFAULT/environments/DEFAULT/aim/mcp-proxies": "mcp-proxies.json",
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer t" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		f, ok := fixtures[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, filepath.Join("testdata", f))
	}))
	t.Cleanup(s.Close)
	cfg := filepath.Join(t.TempDir(), "gops.yaml")
	require.NoError(t, os.WriteFile(cfg, []byte("url: "+s.URL+"/automation\nauth:\n  token: t\n"), 0o600))
	return cfg
}

func TestExport_EndToEnd(t *testing.T) {
	cfg := stubAutomation(t)
	out := t.TempDir()

	// Agent path, first attempt: the two "Petstore" APIs collide; the collisions come back as data.
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"gops", "export", "--config", cfg, "--format", "crd", "--resources", "apis,mcp-proxies", "--output", out, "--no-input", "--json"}, bytes.NewReader(nil), &stdout, &stderr)
	assert.Equal(t, ExitFailure, code)
	var collisions struct {
		Collisions []struct {
			Slug      string
			Resources []struct{ ID, Name string }
		}
	}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &collisions), stdout.String())
	require.Len(t, collisions.Collisions, 1)
	assert.Equal(t, "petstore", collisions.Collisions[0].Slug)
	assert.Equal(t, "7f2c", collisions.Collisions[0].Resources[0].ID)
	assert.Equal(t, "8a1d", collisions.Collisions[0].Resources[1].ID)

	// Second attempt: the agent names the second one.
	stdout.Reset()
	stderr.Reset()
	code = Run(context.Background(), []string{"gops", "export", "--config", cfg, "--format", "crd", "--resources", "apis,mcp-proxies", "--output", out, "--no-input", "--json", "--rename", "8a1d=petstore-v2"}, bytes.NewReader(nil), &stdout, &stderr)
	require.Equal(t, ExitOK, code, stderr.String())
	var report struct {
		Files []struct{ Kind, ID, Name, Path string }
		Notes []string
	}
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &report))
	require.Len(t, report.Files, 3)
	assert.Equal(t, "petstore", report.Files[0].Name)
	assert.Equal(t, "petstore-v2", report.Files[1].Name)
	assert.Equal(t, "github", report.Files[2].Name)
	assert.Equal(t, filepath.Join(out, "mcp-proxies", "github.yaml"), report.Files[2].Path)
	assert.Len(t, report.Notes, 1, "primaryOwner dropped once (the second API has none)")

	raw, err := os.ReadFile(filepath.Join(out, "apis", "petstore.yaml"))
	require.NoError(t, err)
	t.Logf("\n%s", raw)
	var m struct {
		APIVersion string         `yaml:"apiVersion"`
		Kind       string         `yaml:"kind"`
		Metadata   map[string]any `yaml:"metadata"`
		Spec       map[string]any `yaml:"spec"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &m))
	assert.Equal(t, "gravitee.io/v1alpha1", m.APIVersion)
	assert.Equal(t, "ApiV4Definition", m.Kind)
	assert.Equal(t, "petstore", m.Metadata["name"])
	for _, status := range []string{"id", "crossId", "environmentId", "organizationId", "errors", "hrid", "primaryOwner"} {
		assert.NotContains(t, m.Spec, status)
	}
	listeners := m.Spec["listeners"].([]any)
	assert.Equal(t, "HTTP", listeners[0].(map[string]any)["type"], "the Listener union survives the typed SDK round trip")
	assert.Contains(t, m.Spec["plans"].(map[string]any), "Keyless")
	assert.Contains(t, m.Spec["pages"].(map[string]any), "Home")
	flows := m.Spec["flows"].([]any)
	step := flows[0].(map[string]any)["request"].([]any)[0].(map[string]any)
	assert.Equal(t, "rate-limit", step["policy"])
	assert.Equal(t, map[string]any{"rate": map[string]any{"limit": 10}}, step["configuration"], "policy configuration passes through")

	// Third attempt: nothing may be overwritten without --force.
	stdout.Reset()
	stderr.Reset()
	code = Run(context.Background(), []string{"gops", "export", "--config", cfg, "--format", "crd", "--resources", "apis", "--output", out, "--no-input", "--on-collision", "suffix"}, bytes.NewReader(nil), &stdout, &stderr)
	assert.Equal(t, ExitFailure, code)
	assert.Contains(t, stderr.String(), "pass --force")
	assert.Contains(t, stderr.String(), filepath.Join(out, "apis", "petstore.yaml"))

	// The human-readable report of a raw export.
	stdout.Reset()
	code = Run(context.Background(), []string{"gops", "export", "--config", cfg, "--format", "yaml", "--resources", "mcp-proxies", "--output", out, "--no-input", "--force"}, bytes.NewReader(nil), &stdout, &stderr)
	require.Equal(t, ExitOK, code, stderr.String())
	assert.Contains(t, stdout.String(), "wrote "+filepath.Join(out, "mcp-proxies", "github.yaml"))
	assert.Contains(t, stdout.String(), "1 file(s) written")
}

func TestExport_UnauthorizedIsReported(t *testing.T) {
	cfg := stubAutomation(t)
	require.NoError(t, os.WriteFile(cfg, append(mustRead(t, cfg)[:0], []byte("url: "+serverURLFrom(t, cfg)+"\nauth:\n  token: wrong\n")...), 0o600))
	_, msg := execute(t, "export", "--config", cfg, "--format", "json", "--resources", "apis", "--output", t.TempDir())
	assert.Contains(t, msg, "list apis: HTTP 401")
}

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	return b
}

func serverURLFrom(t *testing.T, cfg string) string {
	t.Helper()
	var c struct {
		URL string `yaml:"url"`
	}
	require.NoError(t, yaml.Unmarshal(mustRead(t, cfg), &c))
	return c.URL
}
