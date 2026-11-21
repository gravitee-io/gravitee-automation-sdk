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

package sdk

import (
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// readOnlyInPutBodies lists the read-only properties a PUT body holds, as <schema>.<property>, and what Spec() does with
// each. The spec marks them read-only, but only the platform says whether a PUT refuses them.
var readOnlyInPutBodies = map[string]string{
	"AimMcpProxyStudioTool.entityId": "cleared by clearReadOnly: AIM refuses a studio tool's entityId",
	"ApplicationSpec.primaryOwner":   "kept: APIM ignores it",
	"PageV4.crossId":                 "kept: APIM keeps the crossId of a page sent with its legacy id",
	"Property.encrypted":             "kept: APIM refuses a stored secret sent back without it",
}

// A spec sync that adds a read-only property to a PUT body fails here. Clear it in a clearReadOnly method when the
// platform refuses it on PUT, and list it above either way.
func TestSpec_ReadOnlyFieldsInPutBodies(t *testing.T) {
	raw, err := os.ReadFile("../../openapi/openapi.yaml")
	require.NoError(t, err)
	var spec struct {
		Components struct {
			Schemas map[string]any `yaml:"schemas"`
		} `yaml:"components"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &spec))
	schemas := spec.Components.Schemas

	found := map[string]bool{}
	visited := map[string]bool{}
	var walk func(owner string, node any)
	visit := func(name string) {
		if !visited[name] {
			visited[name] = true
			walk(name, schemas[name])
		}
	}
	walk = func(owner string, node any) {
		schema, ok := node.(map[string]any)
		if !ok {
			return
		}
		if ref, ok := schema["$ref"].(string); ok {
			visit(strings.TrimPrefix(ref, "#/components/schemas/"))
		}
		properties, _ := schema["properties"].(map[string]any)
		for name, property := range properties {
			if p, ok := property.(map[string]any); ok && p["readOnly"] == true {
				found[owner+"."+name] = true
			}
			walk(owner, property)
		}
		walk(owner, schema["items"])
		walk(owner, schema["additionalProperties"])
		for _, key := range []string{"allOf", "oneOf", "anyOf"} {
			variants, _ := schema[key].([]any)
			for _, variant := range variants {
				walk(owner, variant)
			}
		}
	}
	for name := range schemas {
		if base, ok := strings.CutSuffix(name, "State"); ok && schemas[base+"Spec"] != nil {
			visit(base + "Spec")
		}
	}

	assert.Equal(t, slices.Sorted(maps.Keys(readOnlyInPutBodies)), slices.Sorted(maps.Keys(found)))
}
