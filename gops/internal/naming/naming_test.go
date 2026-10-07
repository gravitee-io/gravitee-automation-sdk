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

package naming

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Petstore":              "petstore",
		"Petstore API (v2)":     "petstore-api-v2",
		"  GitHub MCP proxy ":   "github-mcp-proxy",
		"déjà vu":               "d-j-vu",
		"":                      "unnamed",
		"---":                   "unnamed",
		strings.Repeat("a", 70): strings.Repeat("a", 63),
	}
	for in, want := range cases {
		assert.Equal(t, want, Slug(in), in)
	}
	assert.True(t, IsSlug("petstore-2"))
	assert.False(t, IsSlug("Petstore"))
}

func TestPlan_NamesAndCollisions(t *testing.T) {
	names, collisions, err := Plan([]Named{
		{ID: "b", Name: "Petstore"},
		{ID: "a", Name: "Petstore"},
		{ID: "c", Name: "Orders"},
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"c": "orders"}, names)
	require.Len(t, collisions, 1)
	assert.Equal(t, "petstore", collisions[0].Slug)
	assert.Equal(t, []Named{{ID: "a", Name: "Petstore"}, {ID: "b", Name: "Petstore"}}, collisions[0].Resources, "sorted by id")
}

func TestPlan_OverridesWinAndCanCollide(t *testing.T) {
	names, collisions, err := Plan([]Named{
		{ID: "a", Name: "Petstore"},
		{ID: "b", Name: "Orders"},
	}, map[string]string{"a": "orders"})
	require.NoError(t, err)
	assert.Empty(t, names)
	require.Len(t, collisions, 1)
	assert.Equal(t, "orders", collisions[0].Slug)

	_, _, err = Plan(nil, map[string]string{"a": "Not A Slug"})
	assert.ErrorContains(t, err, `"not-a-slug"`)
}

func TestSuffixResolver(t *testing.T) {
	collisions := []Collision{{Slug: "petstore", Resources: []Named{{ID: "a"}, {ID: "b"}, {ID: "c"}}}}
	taken := map[string]bool{"petstore-2": true}
	names, err := SuffixResolver{}.Resolve(collisions, taken)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "petstore", "b": "petstore-3", "c": "petstore-4"}, names)
}

func TestFailResolver(t *testing.T) {
	_, err := FailResolver{}.Resolve([]Collision{{Slug: "petstore", Resources: []Named{{ID: "a", Name: "Petstore"}}}}, nil)
	var ce *CollisionError
	require.ErrorAs(t, err, &ce)
	assert.Contains(t, err.Error(), "--rename")
	assert.Contains(t, err.Error(), "(id a)")
}
