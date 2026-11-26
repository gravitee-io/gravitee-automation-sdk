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

func TestAMApplyDataPlane(t *testing.T) {
	giops := newAMCLI(t)

	out := giops.mustRun("am", "apply", "dataplanes", "-f", testdata(t, "am/dataplane.yaml"), "--yes")

	assert.Contains(t, out, "giops-dataplane applied successfully")
	assert.Equal(t, "Giops data plane", giops.getJSON("am", "dataplanes", "giops-dataplane")["name"])
}

func TestAMDryRunDataPlaneSavesNothing(t *testing.T) {
	giops := newAMCLI(t)

	out := giops.mustRun("am", "apply", "dataplanes", "-f", testdata(t, "am/dataplane.yaml"), "--dry-run")

	assert.Contains(t, out, "Dry run is not configured")
	giops.requireAbsent("am", "dataplanes", "giops-dataplane")
}

func TestAMApplyDataPlaneAfterSayingYes(t *testing.T) {
	giops := newAMCLI(t)

	r := giops.runWithInput("y\n", "am", "apply", "dataplanes", "-f", testdata(t, "am/dataplane.yaml"))

	assert.NoError(t, r.err, r.out)
	assert.Contains(t, r.out, "Dry run is not configured")
	assert.Contains(t, r.out, "Apply?")
	assert.Equal(t, "Giops data plane", giops.getJSON("am", "dataplanes", "giops-dataplane")["name"])
}

func TestAMGetAndListDataPlanes(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "dataplanes", "-f", testdata(t, "am/dataplane.yaml"), "--yes")

	for _, args := range [][]string{
		{"am", "get", "dataplanes", "giops-dataplane"},
		{"am", "get", "dataplanes"},
	} {
		out := giops.mustRun(args...)

		assert.Regexp(t, `ID\s+NAME\s+TYPE`, out)
		assert.Regexp(t, `giops-dataplane\s+Giops data plane\s+mongodb`, out)
	}
}

func TestAMDeleteDataPlane(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "dataplanes", "-f", testdata(t, "am/dataplane.yaml"), "--yes")

	out := giops.mustRun("am", "delete", "dataplanes", "-y", "giops-dataplane")

	assert.Contains(t, out, "giops-dataplane deleted")
	giops.requireAbsent("am", "dataplanes", "giops-dataplane")
}
