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
	"testing"

	"github.com/stretchr/testify/assert"
)

// gammaResource is a resource of a module that uses the APIM client (apim, portal, aim, authz), applied from its
// testdata file.
type gammaResource struct {
	module   string // its module
	resource string // its name on the command line
	file     string // its testdata file
	hrid     string // the hrid in the file
	field    string // a JSON field the file sets, read back after apply
	value    any    // the value of that field
	header   string // regexp of its table header
	row      string // regexp of its table row
	parent   string // the name of its parent flag, empty for a flat resource
	parentID string // the parent id in the file
	// setup returns the CLI with the parent applied; nil starts the module without any resource.
	setup func(t *testing.T) cli
}

// testGammaResource runs what every resource of these modules does: apply, dry run, get and delete; and for a
// child resource, its parent given in the file or as a flag.
func testGammaResource(t *testing.T, r gammaResource) {
	var inParent []string
	if r.parent != "" {
		inParent = []string{"--" + r.parent, r.parentID}
	}
	start := func(t *testing.T) cli {
		if r.setup != nil {
			return r.setup(t)
		}
		return newGammaCLI(t, r.module)
	}
	apply := func(giops cli, t *testing.T) string {
		return giops.mustRun(r.module, "apply", r.resource, "-f", testdata(t, r.file), "--yes")
	}

	t.Run("apply", func(t *testing.T) {
		giops := start(t)

		out := apply(giops, t)

		assert.Contains(t, out, r.hrid+" applied successfully")
		assert.Equal(t, r.value, giops.getJSON(r.module, r.resource, r.hrid, inParent...)[r.field])
	})

	t.Run("dry run saves nothing", func(t *testing.T) {
		giops := start(t)

		out := giops.mustRun(r.module, "apply", r.resource, "-f", testdata(t, r.file), "--dry-run")

		assert.Contains(t, out, "Dry run result: successful")
		giops.requireAbsent(r.module, r.resource, r.hrid, inParent...)
	})

	t.Run("get as a table", func(t *testing.T) {
		giops := start(t)
		apply(giops, t)

		out := giops.mustRun(append([]string{r.module, "get", r.resource, r.hrid}, inParent...)...)

		assert.Regexp(t, r.header, out)
		assert.Regexp(t, r.row, out)
	})

	t.Run("get unknown", func(t *testing.T) {
		r2 := start(t).run(append([]string{r.module, "get", r.resource, "unknown"}, inParent...)...)

		assert.ErrorContains(t, r2.err, "404")
	})

	t.Run("delete by id", func(t *testing.T) {
		giops := start(t)
		apply(giops, t)

		out := giops.mustRun(append([]string{r.module, "delete", r.resource, "-y", r.hrid}, inParent...)...)

		assert.Contains(t, out, r.hrid+" deleted")
		giops.requireAbsent(r.module, r.resource, r.hrid, inParent...)
	})

	t.Run("delete from its file", func(t *testing.T) {
		giops := start(t)
		apply(giops, t)

		out := giops.mustRun(r.module, "delete", r.resource, "-y", "-f", testdata(t, r.file))

		assert.Contains(t, out, r.hrid+" deleted")
		giops.requireAbsent(r.module, r.resource, r.hrid, inParent...)
	})

	if r.parent == "" {
		return
	}

	t.Run("needs its parent", func(t *testing.T) {
		for _, args := range [][]string{
			{r.module, "get", r.resource, r.hrid},
			{r.module, "delete", r.resource, "-y", r.hrid},
		} {
			r2 := newGammaCLI(t, r.module).run(args...)

			assert.ErrorContains(t, r2.err, r.parent+" identifier required", args)
		}
	})

	t.Run("unknown parent", func(t *testing.T) {
		r2 := start(t).run(r.module, "apply", r.resource, "-f", testdata(t, r.file), "--"+r.parent, "unknown", "--yes")

		assert.ErrorContains(t, r2.err, "404")
	})
}

// testIgnoreGroups checks that --ignore-groups and -g reach the dry run and the apply of an APIM resource.
func testIgnoreGroups(t *testing.T, resource, file string) {
	cases := map[string]struct {
		args []string
		want string
	}{
		"--ignore-groups":  {[]string{"--ignore-groups"}, "true"},
		"-g":               {[]string{"-g"}, "true"},
		"without the flag": {nil, "false"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			giops := newAPIMCLI(t)
			args := append([]string{"apim", "apply", resource, "-f", testdata(t, file)}, c.args...)

			r := giops.runWithInput("y\n", args...)

			assert.NoError(t, r.err, r.out)
			puts := giops.queries("PUT")
			if assert.Len(t, puts, 2, "a dry run, then the apply") {
				assert.Equal(t, "true", puts[0].Get("dryRun"))
				assert.Equal(t, "false", puts[1].Get("dryRun"))
			}
			for _, query := range puts {
				assert.Equal(t, c.want, query.Get("ignoreGroups"))
			}
		})
	}
}

// testDeleteDryRun checks the delete of a resource whose API has a delete dry run: delete asks the API first, then
// asks the user.
func testDeleteDryRun(t *testing.T, module, resource, file, hrid string) {
	t.Run("delete --dry-run keeps it", func(t *testing.T) {
		giops := newGammaCLI(t, module)
		giops.mustRun(module, "apply", resource, "-f", testdata(t, file), "--yes")

		out := giops.mustRun(module, "delete", resource, hrid, "--dry-run")

		assert.Contains(t, out, "Dry run result: successful")
		deletes := giops.queries("DELETE")
		if assert.Len(t, deletes, 1) {
			assert.Equal(t, "true", deletes[0].Get("dryRun"))
		}
		giops.getJSON(module, resource, hrid)
	})

	t.Run("delete after saying yes", func(t *testing.T) {
		giops := newGammaCLI(t, module)
		giops.mustRun(module, "apply", resource, "-f", testdata(t, file), "--yes")

		r := giops.runWithInput("y\n", module, "delete", resource, hrid)

		assert.NoError(t, r.err, r.out)
		assert.Contains(t, r.out, "Dry run result: successful")
		assert.Contains(t, r.out, hrid+" deleted")
		deletes := giops.queries("DELETE")
		if assert.Len(t, deletes, 2, "a dry run, then the delete") {
			assert.Equal(t, "true", deletes[0].Get("dryRun"))
			assert.NotEqual(t, "true", deletes[1].Get("dryRun"))
		}
		giops.requireAbsent(module, resource, hrid)
	})

	t.Run("delete unknown stops at the dry run", func(t *testing.T) {
		giops := newGammaCLI(t, module)

		r := giops.runWithInput("y\n", module, "delete", resource, "unknown")

		assert.ErrorContains(t, r.err, "dry run did not complete")
		assert.ErrorContains(t, r.err, "404")
		assert.Len(t, giops.queries("DELETE"), 1, "no delete after the dry run")
	})
}
