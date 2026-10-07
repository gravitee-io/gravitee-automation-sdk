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

package transform

import (
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func apiState() map[string]any {
	return map[string]any{
		"id":             "7f2c",
		"crossId":        "x-1",
		"environmentId":  "DEFAULT",
		"organizationId": "DEFAULT",
		"errors":         map[string]any{"severe": []any{}},
		"hrid":           "",
		"name":           "Petstore",
		"version":        "1.0",
		"type":           "PROXY",
		"description":    "pets",
		"primaryOwner":   map[string]any{"id": "u1", "displayName": "admin"},
		"listeners": []any{map[string]any{
			"type":        "HTTP",
			"paths":       []any{map[string]any{"path": "/pets"}},
			"entrypoints": []any{map[string]any{"type": "http-proxy"}},
		}},
		"endpointGroups": []any{map[string]any{
			"name":      "default",
			"type":      "http-proxy",
			"endpoints": []any{map[string]any{"name": "default", "type": "http-proxy", "configuration": map[string]any{"target": "https://petstore.example"}}},
		}},
		"plans": []any{
			map[string]any{"id": "p1", "hrid": "", "name": "Keyless", "security": map[string]any{"type": "KEY_LESS"}, "status": "PUBLISHED"},
			map[string]any{"id": "p2", "hrid": "", "name": "Keyless", "security": map[string]any{"type": "API_KEY"}, "status": "PUBLISHED"},
		},
		"pages": []any{map[string]any{"id": "pg1", "name": "Home", "type": "MARKDOWN", "content": "# hi"}},
	}
}

func transformAPI(t *testing.T, state map[string]any) (map[string]any, []string) {
	t.Helper()
	return transformAPIWith(t, state, Options{})
}

func transformAPIWith(t *testing.T, state map[string]any, opts Options) (map[string]any, []string) {
	t.Helper()
	tr, err := Lookup("crd", opts)
	require.NoError(t, err)
	kind, _ := resource.Lookup("apis")
	out, notes, err := tr.Transform(resource.Resource{Kind: kind, ID: "7f2c", Name: "Petstore", State: state}, "petstore")
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, yaml.Unmarshal(out, &m), string(out))
	return m, notes
}

func TestCRD_ApiV4Definition(t *testing.T) {
	m, notes := transformAPI(t, apiState())

	assert.Equal(t, "gravitee.io/v1alpha1", m["apiVersion"])
	assert.Equal(t, "ApiV4Definition", m["kind"])
	assert.Equal(t, map[string]any{"name": "petstore"}, m["metadata"])

	spec := m["spec"].(map[string]any)
	for _, status := range []string{"id", "crossId", "environmentId", "organizationId", "errors"} {
		assert.NotContains(t, spec, status, "readOnly status field")
	}
	assert.NotContains(t, spec, "hrid", "empty hrid is dropped: GKO derives it from the object name")
	assert.NotContains(t, spec, "primaryOwner", "not in the CRD")
	assert.Equal(t, "Petstore", spec["name"])
	assert.Equal(t, "PROXY", spec["type"])

	plans := spec["plans"].(map[string]any)
	require.Len(t, plans, 2, "the API lists plans, the CRD keys them by name")
	keyless := plans["Keyless"].(map[string]any)
	assert.Equal(t, map[string]any{"type": "KEY_LESS"}, keyless["security"])
	assert.NotContains(t, keyless, "hrid")
	assert.Contains(t, plans, "Keyless-2", "a second plan with the same name gets a numbered key")

	pages := spec["pages"].(map[string]any)
	assert.Equal(t, "# hi", pages["Home"].(map[string]any)["content"])

	listeners := spec["listeners"].([]any)
	assert.Equal(t, "HTTP", listeners[0].(map[string]any)["type"], "untyped in the CRD: passed through")

	require.Len(t, notes, 1)
	assert.Contains(t, notes[0], "spec.primaryOwner")
}

func TestCRD_KeepsAStoredHrid(t *testing.T) {
	state := apiState()
	state["hrid"] = "petstore"
	m, _ := transformAPI(t, state)
	assert.Equal(t, "petstore", m["spec"].(map[string]any)["hrid"])
}

func TestCRD_StripIDs(t *testing.T) {
	m, _ := transformAPI(t, apiState())
	plans := m["spec"].(map[string]any)["plans"].(map[string]any)
	assert.Equal(t, "p1", plans["Keyless"].(map[string]any)["id"], "kept by default: GKO adopts the plan on apply")

	m, _ = transformAPIWith(t, apiState(), Options{StripIDs: true})
	spec := m["spec"].(map[string]any)
	assert.NotContains(t, spec["plans"].(map[string]any)["Keyless"], "id")
	assert.NotContains(t, spec["pages"].(map[string]any)["Home"], "id")
	assert.Equal(t, "Petstore", spec["name"], "only identifiers go")
}

func TestCRD_McpProxy(t *testing.T) {
	tr, err := Lookup("crd", Options{})
	require.NoError(t, err)
	kind, _ := resource.Lookup("mcp-proxies")
	state := map[string]any{
		"id": "m1", "environmentId": "DEFAULT", "organizationId": "DEFAULT",
		"entityId":    "mcp-proxy.github",
		"name":        "GitHub",
		"contextPath": "/mcp/github",
		"mode":        "PROXY",
		"proxy":       map[string]any{"serverUrl": "https://api.githubcopilot.com/mcp/", "upstreamAuth": map[string]any{"type": "BEARER"}},
		"state":       "STARTED",
		"plans":       []any{map[string]any{"name": "gold", "security": map[string]any{"type": "API_KEY", "source": "HEADER"}}},
	}
	out, _, err := tr.Transform(resource.Resource{Kind: kind, ID: "m1", Name: "GitHub", State: state}, "github")
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, yaml.Unmarshal(out, &m), string(out))
	assert.Equal(t, "McpProxy", m["kind"])
	spec := m["spec"].(map[string]any)
	assert.NotContains(t, spec, "id")
	assert.NotContains(t, spec, "environmentId")
	assert.Equal(t, "mcp-proxy.github", spec["entityId"])
	assert.Equal(t, "https://api.githubcopilot.com/mcp/", spec["proxy"].(map[string]any)["serverUrl"])
	assert.Contains(t, spec, "plans")
}

func TestLookup_TerraformIsASlot(t *testing.T) {
	tr, err := Lookup("tf", Options{})
	require.NoError(t, err)
	kind, _ := resource.Lookup("apis")
	_, _, err = tr.Transform(resource.Resource{Kind: kind}, "x")
	assert.ErrorIs(t, err, ErrNotImplemented)

	_, err = Lookup("hcl", Options{})
	assert.ErrorContains(t, err, `unknown format "hcl"`)
}
