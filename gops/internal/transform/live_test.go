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
	"encoding/json"
	"os"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/resource"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// States captured from a live Gamma (APIM 4.14 master + the PoC list operations): a Console-style
// API and an MCP proxy. They carry what the stubs did not: lifecycleState CREATED, which the CRD's
// enum rejects, and the flat plan security the CRD nests under apiKey.
func liveManifest(t *testing.T, kindName, fixture string) (map[string]any, []string) {
	t.Helper()
	raw, err := os.ReadFile("testdata/" + fixture)
	require.NoError(t, err)
	var state map[string]any
	require.NoError(t, json.Unmarshal(raw, &state))
	kind, _ := resource.Lookup(kindName)
	tr, err := Lookup("crd", Options{})
	require.NoError(t, err)
	out, notes, err := tr.Transform(resource.Resource{Kind: kind, State: state}, "x")
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, yaml.Unmarshal(out, &m), string(out))
	return m["spec"].(map[string]any), notes
}

func TestLive_ApiV4Definition_DropsValuesOutsideTheCRDEnum(t *testing.T) {
	spec, notes := liveManifest(t, "apis", "live-api.json")
	assert.NotContains(t, spec, "lifecycleState")
	assert.Contains(t, notes, `spec.lifecycleState: dropped "CREATED", the CRD accepts PUBLISHED, UNPUBLISHED, DEPRECATED, ARCHIVED`)
}

func TestLive_McpProxy_FoldsPlanSecurityUnderItsType(t *testing.T) {
	spec, notes := liveManifest(t, "mcp-proxies", "live-mcp-proxy.json")
	security := spec["plans"].([]any)[0].(map[string]any)["security"]
	assert.Equal(t, map[string]any{"type": "API_KEY", "apiKey": map[string]any{"source": "HEADER"}}, security)
	assert.Empty(t, notes, "nothing dropped once folded")
}

func TestFold_SuffixAndUntyped(t *testing.T) {
	crd := `{"properties":{"type":{"type":"string","enum":["OAUTH2_AUTH0","GRAVITEE_AM"]},"name":{"type":"string"},
	  "auth0":{"type":"object","properties":{"domain":{"type":"string"}}}}}`
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(crd), &m))
	s := schemaFrom(m)
	assert.Equal(t, map[string]any{"type": "OAUTH2_AUTH0", "name": "a0", "auth0": map[string]any{"domain": "x"}},
		fold(map[string]any{"type": "OAUTH2_AUTH0", "name": "a0", "domain": "x"}, s))
	assert.Equal(t, map[string]any{"type": "GRAVITEE_AM", "name": "am", "domain": "x"},
		fold(map[string]any{"type": "GRAVITEE_AM", "name": "am", "domain": "x"}, s), "no branch for the value: left as is")
}

func schemaFrom(m map[string]any) *schema.Schema { return schema.FromStructural(m) }
