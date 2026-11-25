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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAMCheck(t *testing.T) {
	out := newAMCLI(t).mustRun("am", "check")

	assert.Contains(t, out, "module am is accessible")
}

func TestAMApplyDomain(t *testing.T) {
	giops := newAMCLI(t)

	out := giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")

	assert.Contains(t, out, "giops-domain applied successfully")
	domain := giops.getJSON("am", "domains", "giops-domain")
	assert.Equal(t, "Giops domain", domain["name"])
	assert.Equal(t, "/giops", domain["path"])
	assert.Equal(t, "default", domain["dataPlaneId"])
	assert.Equal(t, true, domain["enabled"])
}

func TestAMApplyDomainFromJSON(t *testing.T) {
	giops := newAMCLI(t)

	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.json"), "--yes")

	assert.Equal(t, "JSON domain", giops.getJSON("am", "domains", "giops-json-domain")["name"])
}

func TestAMApplyDomainAsksFirst(t *testing.T) {
	t.Run("after showing the dry run result", func(t *testing.T) {
		r := newAMCLI(t).runWithInput("n\n", "am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"))

		require.NoError(t, r.err, r.out)
		assert.Regexp(t, `(?s)Dry run result: successful.*Apply\? \[y/N\]`, r.out)
	})

	t.Run("no keeps the domain out", func(t *testing.T) {
		giops := newAMCLI(t)

		r := giops.runWithInput("n\n", "am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"))

		assert.NoError(t, r.err, r.out)
		assert.Contains(t, r.out, "Apply? [y/N]")
		giops.requireAbsent("am", "domains", "giops-domain")
	})

	t.Run("yes applies it", func(t *testing.T) {
		giops := newAMCLI(t)

		r := giops.runWithInput("y\n", "am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"))

		require.NoError(t, r.err, r.out)
		assert.Contains(t, r.out, "giops-domain applied successfully")
		assert.Equal(t, "Giops domain", giops.getJSON("am", "domains", "giops-domain")["name"])
	})
}

func TestAMDryRunDomain(t *testing.T) {
	giops := newAMCLI(t)

	out := giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--dry-run")

	assert.Contains(t, out, "Dry run result: successful")
	giops.requireAbsent("am", "domains", "giops-domain")
}

func TestAMDryRunErrorsStopTheApply(t *testing.T) {
	t.Run("dry run only", func(t *testing.T) {
		giops := newRejectingAMCLI(t)

		r := giops.run("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--dry-run")

		assert.ErrorContains(t, r.err, "all PUT request are rejected on purpose")
	})

	t.Run("before asking", func(t *testing.T) {
		giops := newRejectingAMCLI(t)

		r := giops.runWithInput("y\n", "am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"))

		assert.ErrorContains(t, r.err, "all PUT request are rejected on purpose")
		assert.NotContains(t, r.out, "Apply?")
		giops.requireAbsent("am", "domains", "giops-domain")
	})
}

func TestAMApplyDomainWithYesSkipsTheDryRun(t *testing.T) {
	giops := newRejectingAMCLI(t)

	out := giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")

	assert.NotContains(t, out, "Dry run")
	assert.Equal(t, "Giops domain", giops.getJSON("am", "domains", "giops-domain")["name"])
}

func TestAMApplyDomainRefusesDryRunWithYes(t *testing.T) {
	r := newAMCLI(t).run("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--dry-run", "--yes")

	assert.ErrorContains(t, r.err, "none of the others can be")
}

func TestAMApplyDomainNeedsAFile(t *testing.T) {
	r := newAMCLI(t).run("am", "apply", "domains", "--yes")

	assert.ErrorContains(t, r.err, "no file specified")
}

func TestAMGetDomainAsTable(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")

	out := giops.mustRun("am", "get", "domains", "giops-domain")

	assert.Regexp(t, `KEY\s+NAME\s+PATH\s+ENABLED`, out)
	assert.Regexp(t, `giops-domain\s+Giops domain\s+/giops\s+true`, out)
}

func TestAMGetDomainAsYAMLCanBeAppliedBack(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")
	exported := giops.mustRun("am", "get", "domains", "giops-domain", "-o", "yaml")
	assert.Contains(t, exported, "dataPlaneId: default")

	giops.mustRun("am", "delete", "domains", "-y", "giops-domain")
	file := filepath.Join(t.TempDir(), "exported.yaml")
	require.NoError(t, os.WriteFile(file, []byte(exported), 0o600))
	giops.mustRun("am", "apply", "domains", "-f", file, "--yes")

	assert.Equal(t, "default", giops.getJSON("am", "domains", "giops-domain")["dataPlaneId"])
}

func TestAMListDomains(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/other-domain.yaml"), "--yes")

	out := giops.mustRun("am", "get", "domains")

	assert.Regexp(t, `KEY\s+NAME`, out)
	assert.Regexp(t, `giops-domain\s+Giops domain\s+/giops\s+true`, out)
	assert.Regexp(t, `giops-other-domain\s+Other domain\s+/giops-other\s+false`, out)
}

func TestAMListNoDomain(t *testing.T) {
	requireMock(t, "a live server may have domains")
	out := newAMCLI(t).mustRun("am", "get", "domains")

	assert.Contains(t, out, "No resources found")
}

func TestAMGetUnknownDomain(t *testing.T) {
	r := newAMCLI(t).run("am", "get", "domains", "unknown")

	assert.ErrorContains(t, r.err, "unknown")
}

func TestAMDeleteDomain(t *testing.T) {
	cases := map[string]func(t *testing.T) []string{
		"by argument": func(*testing.T) []string { return []string{"giops-domain"} },
		"by --id":     func(*testing.T) []string { return []string{"--id", "giops-domain"} },
		"by file":     func(t *testing.T) []string { return []string{"-f", testdata(t, "am/domain.yaml")} },
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			giops := newAMCLI(t)
			giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")

			out := giops.mustRun(append([]string{"am", "delete", "domains", "-y"}, target(t)...)...)

			assert.Contains(t, out, "giops-domain deleted")
			giops.requireAbsent("am", "domains", "giops-domain")
		})
	}
}

func TestAMDeleteDomainNeedsATarget(t *testing.T) {
	r := newAMCLI(t).run("am", "delete", "domains", "-y")

	assert.ErrorContains(t, r.err, "provide id")
}

func TestAMApplyDomainHasNoGroupOptions(t *testing.T) {
	r := newAMCLI(t).run("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes", "--ignore-members")

	assert.ErrorContains(t, r.err, "unknown flag: --ignore-members")
}
