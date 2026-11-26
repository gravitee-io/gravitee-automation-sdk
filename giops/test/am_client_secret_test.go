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

var inGiopsApp = []string{"--domain", "giops-domain", "--application", "giops-app"}

// newAMCLIWithApplication applies giops-domain and its application giops-app, the parents of giops-secret.
func newAMCLIWithApplication(t *testing.T) cli {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")
	giops.mustRun("am", "apply", "applications", "-f", testdata(t, "am/application.yaml"), "--yes")
	return giops
}

func TestAMApplySecretWithParentsFromTheFile(t *testing.T) {
	giops := newAMCLIWithApplication(t)

	out := giops.mustRun("am", "apply", "secrets", "-f", testdata(t, "am/secret.yaml"), "--yes")

	assert.Contains(t, out, "giops-secret applied successfully")
	secret := giops.getJSON("am", "secrets", "giops-secret", inGiopsApp...)
	assert.Equal(t, "giops-secret", secret["key"])
	assert.NotContains(t, secret, "value", "the API never returns the value")
}

func TestAMDryRunSecret(t *testing.T) {
	giops := newAMCLIWithApplication(t)

	out := giops.mustRun("am", "apply", "secrets", "-f", testdata(t, "am/secret.yaml"), "--dry-run")

	assert.Contains(t, out, "Dry run result: successful")
	giops.requireAbsent("am", "secrets", "giops-secret", inGiopsApp...)
}

func TestAMReapplySameSecret(t *testing.T) {
	giops := newAMCLIWithApplication(t)
	giops.mustRun("am", "apply", "secrets", "-f", testdata(t, "am/secret.yaml"), "--yes")

	out := giops.mustRun("am", "apply", "secrets", "-f", testdata(t, "am/secret.yaml"), "--yes")

	assert.Contains(t, out, "giops-secret applied successfully")
}

func TestAMApplicationFlagWinsOverTheSecretFile(t *testing.T) {
	giops := newAMCLIWithApplication(t)
	giops.mustRun("am", "apply", "applications", "-f", testdata(t, "am/other-application.yaml"), "--yes")

	giops.mustRun("am", "apply", "secrets", "-f", testdata(t, "am/secret.yaml"), "--application", "giops-other-app", "--yes")

	giops.getJSON("am", "secrets", "giops-secret", "--domain", "giops-domain", "--application", "giops-other-app")
	giops.requireAbsent("am", "secrets", "giops-secret", inGiopsApp...)
}

func TestAMGetSecretAsTable(t *testing.T) {
	giops := newAMCLIWithApplication(t)
	giops.mustRun("am", "apply", "secrets", "-f", testdata(t, "am/secret.yaml"), "--yes")

	out := giops.mustRun(append([]string{"am", "get", "secrets", "giops-secret"}, inGiopsApp...)...)

	assert.Regexp(t, `KEY\s+EXPIRESAT`, out)
	assert.Contains(t, out, "giops-secret")
}

func TestAMListSecretsOfAnApplication(t *testing.T) {
	giops := newAMCLIWithApplication(t)
	giops.mustRun("am", "apply", "secrets", "-f", testdata(t, "am/secret.yaml"), "--yes")

	out := giops.mustRun(append([]string{"am", "get", "secrets"}, inGiopsApp...)...)

	assert.Contains(t, out, "giops-secret")
}

func TestAMDeleteSecret(t *testing.T) {
	cases := map[string]func(t *testing.T) []string{
		"by id with both flags": func(*testing.T) []string { return append([]string{"giops-secret"}, inGiopsApp...) },
		"from its file":         func(t *testing.T) []string { return []string{"-f", testdata(t, "am/secret.yaml")} },
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			giops := newAMCLIWithApplication(t)
			giops.mustRun("am", "apply", "secrets", "-f", testdata(t, "am/secret.yaml"), "--yes")

			out := giops.mustRun(append([]string{"am", "delete", "secrets", "-y"}, target(t)...)...)

			assert.Contains(t, out, "giops-secret deleted")
			giops.requireAbsent("am", "secrets", "giops-secret", inGiopsApp...)
		})
	}
}

func TestAMSecretNeedsBothParents(t *testing.T) {
	cases := map[string]struct {
		args []string
		want string
	}{
		"no parent":        {[]string{"am", "get", "secrets", "giops-secret"}, "application identifier required"},
		"no domain":        {[]string{"am", "get", "secrets", "giops-secret", "--application", "giops-app"}, "domain identifier required"},
		"no application":   {[]string{"am", "get", "secrets", "giops-secret", "--domain", "giops-domain"}, "application identifier required"},
		"delete by id":     {[]string{"am", "delete", "secrets", "-y", "giops-secret", "--application", "giops-app"}, "domain identifier required"},
		"list, no parents": {[]string{"am", "get", "secrets"}, "application identifier required"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			r := newAMCLI(t).run(c.args...)

			assert.ErrorContains(t, r.err, c.want)
		})
	}
}

func TestAMSecretOfAnUnknownApplication(t *testing.T) {
	giops := newAMCLIWithApplication(t)

	r := giops.run("am", "get", "secrets", "--domain", "giops-domain", "--application", "unknown")

	assert.ErrorContains(t, r.err, "404")
}

func TestAMDeleteApplicationRemovesItsSecrets(t *testing.T) {
	giops := newAMCLIWithApplication(t)
	giops.mustRun("am", "apply", "secrets", "-f", testdata(t, "am/secret.yaml"), "--yes")

	giops.mustRun("am", "delete", "applications", "-y", "giops-app", "--domain", "giops-domain")

	giops.requireAbsent("am", "secrets", "giops-secret", inGiopsApp...)
}
