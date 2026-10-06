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

package export

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/naming"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeList(ctx context.Context, kind *resource.Kind) ([]resource.Resource, error) {
	mk := func(id, name string) resource.Resource {
		return resource.Resource{Kind: kind, ID: id, Name: name, State: map[string]any{"id": id, "name": name}}
	}
	return []resource.Resource{mk("1", "Petstore"), mk("2", "Petstore"), mk("3", "Orders")}, nil
}

func TestRun_CollisionsFailWithoutAResolver(t *testing.T) {
	_, err := Run(context.Background(), Options{Format: "json", Resources: []string{"apis"}, Output: t.TempDir()}, Deps{List: fakeList})
	ce, ok := IsCollision(err)
	require.True(t, ok, err)
	assert.Equal(t, "petstore", ce.Collisions[0].Slug)
}

func TestRun_SuffixResolverRenamesAndOverwriteNeedsForce(t *testing.T) {
	dir := t.TempDir()
	opts := Options{Format: "json", Resources: []string{"apis"}, Output: dir}
	deps := Deps{List: fakeList, Resolver: naming.SuffixResolver{}}

	report, err := Run(context.Background(), opts, deps)
	require.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join(dir, "apis", "orders.json"),
		filepath.Join(dir, "apis", "petstore-2.json"),
		filepath.Join(dir, "apis", "petstore.json"),
	}, SortedPaths(report.Files))
	content, err := os.ReadFile(filepath.Join(dir, "apis", "petstore-2.json"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `"id": "2"`)

	_, err = Run(context.Background(), opts, deps)
	var exists *ExistsError
	require.ErrorAs(t, err, &exists)
	assert.Len(t, exists.Paths, 3)

	declined := false
	deps.ConfirmOverwrite = func(paths []string) (bool, error) { declined = true; return false, nil }
	_, err = Run(context.Background(), opts, deps)
	require.ErrorAs(t, err, &exists)
	assert.True(t, declined)
	assert.True(t, exists.Declined)

	opts.Force = true
	_, err = Run(context.Background(), opts, deps)
	require.NoError(t, err)
}

func TestRun_RenameOverridesTheSlug(t *testing.T) {
	dir := t.TempDir()
	report, err := Run(context.Background(), Options{
		Format: "yaml", Resources: []string{"apis"}, Output: dir,
		Renames: map[string]string{"2": "petstore-legacy"},
	}, Deps{List: fakeList})
	require.NoError(t, err)
	assert.Contains(t, SortedPaths(report.Files), filepath.Join(dir, "apis", "petstore-legacy.yaml"))
}

func TestRun_UnknownKindAndFormat(t *testing.T) {
	_, err := Run(context.Background(), Options{Format: "json", Resources: []string{"nope"}}, Deps{List: fakeList})
	assert.ErrorContains(t, err, `unknown resource kind "nope"`)
	_, err = Run(context.Background(), Options{Format: "nope"}, Deps{})
	assert.ErrorContains(t, err, `unknown format "nope"`)
}
