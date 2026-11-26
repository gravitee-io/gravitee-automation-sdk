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

// domainChild is an AM resource that lives in a domain, such as a reporter, applied from its testdata file into
// giops-domain.
type domainChild struct {
	resource string // its name on the command line
	file     string // its testdata file
	key      string // the key in the file
	name     string // the name in the file
	header   string // regexp of its table header
	row      string // regexp of its table row
}

// testDomainChild runs what every domain child does: apply, dry run, get, list and delete in its domain, given in
// the file or with --domain.
func testDomainChild(t *testing.T, child domainChild) {
	inDomain := []string{"--domain", "giops-domain"}
	setup := func(t *testing.T) cli {
		giops := newAMCLI(t)
		giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")
		return giops
	}
	apply := func(giops cli, t *testing.T) string {
		return giops.mustRun("am", "apply", child.resource, "-f", testdata(t, child.file), "--yes")
	}

	t.Run("apply in its domain", func(t *testing.T) {
		giops := setup(t)

		out := apply(giops, t)

		assert.Contains(t, out, child.key+" applied successfully")
		assert.Equal(t, child.name, giops.getJSON("am", child.resource, child.key, inDomain...)["name"])
	})

	t.Run("dry run saves nothing", func(t *testing.T) {
		giops := setup(t)

		out := giops.mustRun("am", "apply", child.resource, "-f", testdata(t, child.file), "--dry-run")

		assert.Contains(t, out, "Dry run result: successful")
		giops.requireAbsent("am", child.resource, child.key, inDomain...)
	})

	t.Run("get as a table", func(t *testing.T) {
		giops := setup(t)
		apply(giops, t)

		out := giops.mustRun(append([]string{"am", "get", child.resource, child.key}, inDomain...)...)

		assert.Regexp(t, child.header, out)
		assert.Regexp(t, child.row, out)
	})

	t.Run("list the domain's", func(t *testing.T) {
		giops := setup(t)
		apply(giops, t)

		out := giops.mustRun(append([]string{"am", "get", child.resource}, inDomain...)...)

		assert.Regexp(t, child.row, out)
	})

	t.Run("delete by id", func(t *testing.T) {
		giops := setup(t)
		apply(giops, t)

		out := giops.mustRun(append([]string{"am", "delete", child.resource, "-y", child.key}, inDomain...)...)

		assert.Contains(t, out, child.key+" deleted")
		giops.requireAbsent("am", child.resource, child.key, inDomain...)
	})

	t.Run("delete from its file", func(t *testing.T) {
		giops := setup(t)
		apply(giops, t)

		out := giops.mustRun("am", "delete", child.resource, "-y", "-f", testdata(t, child.file))

		assert.Contains(t, out, child.key+" deleted")
		giops.requireAbsent("am", child.resource, child.key, inDomain...)
	})

	t.Run("needs its domain", func(t *testing.T) {
		for _, args := range [][]string{
			{"am", "get", child.resource, child.key},
			{"am", "get", child.resource},
			{"am", "delete", child.resource, "-y", child.key},
		} {
			r := newAMCLI(t).run(args...)

			assert.ErrorContains(t, r.err, "domain identifier required", args)
		}
	})

	t.Run("unknown domain", func(t *testing.T) {
		r := newAMCLI(t).run("am", "get", child.resource, "--domain", "unknown")

		assert.ErrorContains(t, r.err, "404")
	})
}
