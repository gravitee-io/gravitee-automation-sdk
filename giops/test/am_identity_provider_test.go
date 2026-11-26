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
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests read the mock through the SDK, so a broken CLI get cannot hide a broken apply or delete.

func TestAMApplyIdentityProviderInItsDomain(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")

	out := giops.mustRun("am", "apply", "identityproviders", "-f", testdata(t, "am/identity-provider.yaml"), "--yes")

	assert.Contains(t, out, "giops-idp applied successfully")
	idp := giops.identityProvider("giops-domain", "giops-idp")
	require.NotNil(t, idp)
	assert.Equal(t, "Giops users", *idp.Name)
	assert.Equal(t, "inline-am-idp", *idp.Type)
	assert.Contains(t, idp.Configuration.Object, "users")
}

func TestAMDryRunIdentityProvider(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")

	out := giops.mustRun("am", "apply", "identityproviders", "-f", testdata(t, "am/identity-provider.yaml"), "--dry-run")

	assert.Contains(t, out, "Dry run result: successful")
	assert.Nil(t, giops.identityProvider("giops-domain", "giops-idp"))
}

func TestAMApplyIdentityProviderInAMissingDomain(t *testing.T) {
	r := newAMCLI(t).run("am", "apply", "identityproviders", "-f", testdata(t, "am/identity-provider.yaml"), "--yes")

	assert.ErrorContains(t, r.err, "Domain key [giops-domain] not found")
}

func TestAMDeleteIdentityProviderFromItsFile(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")
	giops.mustRun("am", "apply", "identityproviders", "-f", testdata(t, "am/identity-provider.yaml"), "--yes")

	out := giops.mustRun("am", "delete", "identityproviders", "-y", "-f", testdata(t, "am/identity-provider.yaml"))

	assert.Contains(t, out, "giops-idp deleted")
	assert.Nil(t, giops.identityProvider("giops-domain", "giops-idp"))
}

func TestAMIdentityProviderNeedsItsDomain(t *testing.T) {
	cases := map[string][]string{
		"get":          {"am", "get", "identityproviders", "giops-idp"},
		"delete by id": {"am", "delete", "identityproviders", "-y", "giops-idp"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			r := newAMCLI(t).run(args...)

			assert.ErrorContains(t, r.err, "domain identifier required")
		})
	}
}

func TestAMGetIdentityProviderWithDomainFlag(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")
	giops.mustRun("am", "apply", "identityproviders", "-f", testdata(t, "am/identity-provider.yaml"), "--yes")

	out := giops.mustRun("am", "get", "identityproviders", "giops-idp", "--domain", "giops-domain")

	assert.Regexp(t, `KEY\s+TYPE`, out)
	assert.Regexp(t, `giops-idp\s+inline-am-idp`, out)
}

func TestAMGetUnknownIdentityProvider(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")

	r := giops.run("am", "get", "identityproviders", "unknown", "--domain", "giops-domain")

	assert.ErrorContains(t, r.err, "404")
}

func TestAMDeleteIdentityProviderWithDomainFlag(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")
	giops.mustRun("am", "apply", "identityproviders", "-f", testdata(t, "am/identity-provider.yaml"), "--yes")

	out := giops.mustRun("am", "delete", "identityproviders", "-y", "giops-idp", "--domain", "giops-domain")

	assert.Contains(t, out, "giops-idp deleted")
	assert.Nil(t, giops.identityProvider("giops-domain", "giops-idp"))
}

func TestAMDeleteDomainRemovesItsIdentityProviders(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")
	giops.mustRun("am", "apply", "identityproviders", "-f", testdata(t, "am/identity-provider.yaml"), "--yes")

	giops.mustRun("am", "delete", "domains", "-y", "giops-domain")

	assert.Nil(t, giops.identityProvider("giops-domain", "giops-idp"))
}

// identityProvider reads one provider from the mock server, or returns nil when it does not exist.
func (c cli) identityProvider(domain, key string) *sdk.IdentityProvider {
	c.t.Helper()
	client, err := sdk.NewAMClient(c.api)
	require.NoError(c.t, err)
	resp, err := client.GetIdentityProviderWithResponse(c.t.Context(), domain, key)
	require.NoError(c.t, err)
	if resp.StatusCode() == http.StatusNotFound {
		return nil
	}
	require.Equal(c.t, http.StatusOK, resp.StatusCode(), string(resp.Body))
	return resp.JSON200
}

func TestAMListIdentityProvidersOfADomain(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")
	giops.mustRun("am", "apply", "identityproviders", "-f", testdata(t, "am/identity-provider.yaml"), "--yes")
	giops.mustRun("am", "apply", "identityproviders", "-f", testdata(t, "am/other-identity-provider.yaml"), "--yes")

	out := giops.mustRun("am", "get", "identityproviders", "--domain", "giops-domain")

	assert.Regexp(t, `KEY\s+TYPE`, out)
	assert.Regexp(t, `giops-idp\s+inline-am-idp`, out)
	assert.Regexp(t, `other-idp\s+inline-am-idp`, out)
}

func TestAMListIdentityProvidersAsJSON(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")
	giops.mustRun("am", "apply", "identityproviders", "-f", testdata(t, "am/identity-provider.yaml"), "--yes")
	giops.mustRun("am", "apply", "identityproviders", "-f", testdata(t, "am/other-identity-provider.yaml"), "--yes")

	out := giops.mustRun("am", "get", "identityproviders", "--domain", "giops-domain", "-o", "json")

	var providers []map[string]any
	require.NoErrorf(t, json.Unmarshal([]byte(out), &providers), "not a JSON list:\n%s", out)
	keys := make([]any, len(providers))
	for i, p := range providers {
		keys[i] = p["key"]
	}
	assert.ElementsMatch(t, []any{"giops-idp", "other-idp"}, keys)
}

func TestAMListNoIdentityProvider(t *testing.T) {
	giops := newAMCLI(t)
	giops.mustRun("am", "apply", "domains", "-f", testdata(t, "am/domain.yaml"), "--yes")

	out := giops.mustRun("am", "get", "identityproviders", "--domain", "giops-domain")

	assert.Contains(t, out, "No resources found")
}

func TestAMListIdentityProvidersNeedsADomain(t *testing.T) {
	r := newAMCLI(t).run("am", "get", "identityproviders")

	assert.ErrorContains(t, r.err, "domain identifier required")
}

func TestAMListIdentityProvidersOfAnUnknownDomain(t *testing.T) {
	r := newAMCLI(t).run("am", "get", "identityproviders", "--domain", "unknown")

	assert.ErrorContains(t, r.err, "404")
}
