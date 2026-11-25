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
	"net/http"
	"testing"

	apimmock "github.com/gravitee-io/gravitee-automation-sdk/apim-mock-server/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIMCheck(t *testing.T) {
	newAPIMCLI(t).mustRun("apim", "check")
}

func TestAPIMApplyGroup(t *testing.T) {
	giops := newAPIMCLI(t)

	out := giops.mustRun("apim", "apply", "groups", "-f", testdata(t, "apim/group.yaml"), "--yes")

	assert.Contains(t, out, "giops-group applied successfully")
	group := giops.getJSON("apim", "groups", "giops-group")
	assert.Equal(t, "Giops group", group["name"])
	assert.Equal(t, "giops-group", group["hrid"])
}

func TestAPIMDryRunGroup(t *testing.T) {
	giops := newAPIMCLI(t)

	out := giops.mustRun("apim", "apply", "groups", "-f", testdata(t, "apim/group.yaml"), "--dry-run")

	assert.Contains(t, out, "Dry run result: successful")
	giops.requireAbsent("apim", "groups", "giops-group")
}

func TestAPIMDryRunGroupWithoutBody(t *testing.T) {
	requireMock(t, "a server answering the dry run with no body")
	noContent := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	giops := newCLI(t, "apim", noContent, apimmock.BasePath, mockAPI())

	r := giops.run("apim", "apply", "groups", "-f", testdata(t, "apim/group.yaml"), "--dry-run")

	assert.ErrorContains(t, r.err, "the dry run response has no JSON body")
}

func TestAPIMApplyGroupAfterSayingYes(t *testing.T) {
	giops := newAPIMCLI(t)

	r := giops.runWithInput("y\n", "apim", "apply", "groups", "-f", testdata(t, "apim/group.yaml"))

	assert.NoError(t, r.err, r.out)
	assert.Equal(t, "Giops group", giops.getJSON("apim", "groups", "giops-group")["name"])
}

func TestAPIMGetGroupAsTable(t *testing.T) {
	giops := newAPIMCLI(t)
	giops.mustRun("apim", "apply", "groups", "-f", testdata(t, "apim/group.yaml"), "--yes")

	out := giops.mustRun("apim", "get", "groups", "giops-group")

	assert.Regexp(t, `HRID\s+NAME`, out)
	assert.Regexp(t, `giops-group\s+Giops group`, out)
}

func TestAPIMGetUnknownGroup(t *testing.T) {
	r := newAPIMCLI(t).run("apim", "get", "groups", "unknown")

	assert.ErrorContains(t, r.err, "unknown")
}

func TestAPIMDeleteGroup(t *testing.T) {
	cases := map[string]func(t *testing.T) []string{
		"by argument": func(*testing.T) []string { return []string{"giops-group"} },
		"by file":     func(t *testing.T) []string { return []string{"-f", testdata(t, "apim/group.yaml")} },
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			giops := newAPIMCLI(t)
			giops.mustRun("apim", "apply", "groups", "-f", testdata(t, "apim/group.yaml"), "--yes")

			out := giops.mustRun(append([]string{"apim", "delete", "groups", "-y"}, target(t)...)...)

			assert.Contains(t, out, "giops-group deleted")
			giops.requireAbsent("apim", "groups", "giops-group")
		})
	}
}

func TestAPIMDeleteUnknownGroup(t *testing.T) {
	r := newAPIMCLI(t).run("apim", "delete", "groups", "-y", "unknown")

	assert.Error(t, r.err)
}

func TestAPIMApplyGroupIgnoringMembers(t *testing.T) {
	cases := map[string]struct {
		args []string
		want string
	}{
		"--ignore-members": {[]string{"--ignore-members"}, "true"},
		"-m":               {[]string{"-m"}, "true"},
		"without the flag": {nil, "false"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			giops := newAPIMCLI(t)
			args := append([]string{"apim", "apply", "groups", "-f", testdata(t, "apim/group.yaml"), "--yes"}, c.args...)

			giops.mustRun(args...)

			puts := giops.queries(http.MethodPut)
			require.Len(t, puts, 1)
			assert.Equal(t, c.want, puts[0].Get("ignoreMembers"))
		})
	}
}

func TestAPIMDryRunGroupIgnoringMembers(t *testing.T) {
	giops := newAPIMCLI(t)

	r := giops.runWithInput("y\n", "apim", "apply", "groups", "-f", testdata(t, "apim/group.yaml"), "--ignore-members")

	require.NoError(t, r.err, r.out)
	puts := giops.queries(http.MethodPut)
	require.Len(t, puts, 2, "a dry run, then the apply")
	assert.Equal(t, "true", puts[0].Get("dryRun"))
	assert.Equal(t, "false", puts[1].Get("dryRun"))
	for _, query := range puts {
		assert.Equal(t, "true", query.Get("ignoreMembers"))
	}
}
