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

package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSecretValue = strings.Repeat("v", minClientSecretLength)

func testClientSecret() ClientSecret {
	return ClientSecret{Key: "ci", Value: new(testSecretValue)}
}

func createAMServerWithApplication(t *testing.T) (*MockAM, *httptest.Server) {
	t.Helper()
	am, srv := createAMServerWithDomain(t)
	defaultTenant(am).Applications.Put(newChild(testApplication(), defaultDomainKey))
	return am, srv
}

func storeClientSecret(am *MockAM, secret ClientSecret) {
	defaultTenant(am).ClientSecrets.Put(newChild(secret, applicationIdentity(defaultDomainKey, testApplication().Key)))
}

func TestListClientSecretsSortedByKeyWithoutValue(t *testing.T) {
	am, srv := createAMServerWithApplication(t)
	storeClientSecret(am, ClientSecret{Key: "b", Value: new(testSecretValue)})
	storeClientSecret(am, ClientSecret{Key: "a", Value: new(testSecretValue)})

	assertListEqual(t, clientSecretsURL(srv), []ClientSecret{{Key: "a"}, {Key: "b"}})
}

func TestListClientSecrets404WhenApplicationMissing(t *testing.T) {
	_, srv := createAMServerWithDomain(t)

	resp := httpGetURL(t, clientSecretsURL(srv))
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, notFoundError("Application", "test"), decodeTo[Error](t, resp))
}

func TestGetClientSecretWithoutValue(t *testing.T) {
	am, srv := createAMServerWithApplication(t)
	storeClientSecret(am, testClientSecret())

	assertGetEqual(t, clientSecretsURL(srv), "ci", ClientSecret{Key: "ci"})
}

func TestGetClientSecret404(t *testing.T) {
	_, srv := createAMServerWithApplication(t)

	assertGet404(t, clientSecretsURL(srv), "ci", "ClientSecret")
}

func TestPutGetClientSecretWithoutValue(t *testing.T) {
	am, srv := createAMServerWithApplication(t)

	got := putJSON(t, clientSecretsURL(srv), jsonObject{"key": "ci", "value": testSecretValue})

	assert.Equal(t, jsonObject{"key": "ci"}, got)
	assert.Equal(t, jsonObject{"key": "ci"}, getJSON(t, clientSecretsURL(srv), "ci"))
	stored, ok := isChildOf(defaultTenant(am).ClientSecrets, applicationIdentity(defaultDomainKey, "test"), "ci")
	require.True(t, ok)
	assert.Equal(t, testSecretValue, *stored.Self.Value)
}

func TestDeleteClientSecret(t *testing.T) {
	am, srv := createAMServerWithApplication(t)
	storeClientSecret(am, testClientSecret())

	assertDeleteGone(t, clientSecretsURL(srv), "ci", "ClientSecret")
}

func TestDeleteClientSecretMissing204(t *testing.T) {
	_, srv := createAMServerWithDomain(t)

	res := httpDelete(t, clientSecretsURL(srv), "ci")
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

type clientSecretUpsertCase struct {
	name          string
	setup         func(*MockAM)
	domainKey     string
	secret        ClientSecret
	status        int
	errorContains string
}

func clientSecretUpsertCases() []clientSecretUpsertCase {
	withApplication := func(am *MockAM) {
		defaultTenant(am).Domains.Put(Domain{Key: defaultDomainKey, Name: "Test"})
		defaultTenant(am).Applications.Put(newChild(testApplication(), defaultDomainKey))
	}
	withSecret := func(am *MockAM) {
		withApplication(am)
		storeClientSecret(am, testClientSecret())
	}
	withMaxSecrets := func(am *MockAM) {
		withApplication(am)
		for i := range maxClientSecrets {
			storeClientSecret(am, ClientSecret{Key: fmt.Sprintf("s%d", i), Value: new(testSecretValue)})
		}
	}
	return []clientSecretUpsertCase{
		{"creates", withApplication, defaultDomainKey, testClientSecret(), http.StatusOK, ""},
		{"same value is unchanged", withSecret, defaultDomainKey, testClientSecret(), http.StatusOK, ""},
		{"same value at the limit is unchanged", withMaxSecrets, defaultDomainKey, ClientSecret{Key: "s0", Value: new(testSecretValue)}, http.StatusOK, ""},
		{"rejects a different value", withSecret, defaultDomainKey, ClientSecret{Key: "ci", Value: new(testSecretValue + "x")}, http.StatusBadRequest, "client secret is immutable"},
		{"rejects a short value", withApplication, defaultDomainKey, ClientSecret{Key: "ci", Value: new(testSecretValue[1:])}, http.StatusBadRequest, "at least 43 characters"},
		{"rejects more than the limit", withMaxSecrets, defaultDomainKey, testClientSecret(), http.StatusBadRequest, "more than 10 client secrets"},
		{"404 when the application is missing", func(am *MockAM) { defaultTenant(am).Domains.Put(Domain{Key: defaultDomainKey, Name: "Test"}) }, defaultDomainKey, testClientSecret(), http.StatusNotFound, "Application [test] not found"},
		{"404 when the domain is missing", func(*MockAM) {}, "missing", testClientSecret(), http.StatusNotFound, "Domain key [missing] not found"},
	}
}

func TestUpsertClientSecretRules(t *testing.T) {
	for _, tc := range clientSecretUpsertCases() {
		t.Run(tc.name, func(t *testing.T) {
			am, srv := createAMServer(t)
			tc.setup(am)

			resp := httpPut(t, collectionURL(srv, "/domains/"+tc.domainKey+"/applications/test/secrets"), tc.secret)
			defer resp.Body.Close()

			assert.Equal(t, tc.status, resp.StatusCode)
			if tc.status == http.StatusOK {
				assert.Equal(t, ClientSecret{Key: tc.secret.Key}, decodeTo[ClientSecret](t, resp))
			} else {
				assert.Contains(t, *decodeTo[Error](t, resp).Message, tc.errorContains)
			}
		})
	}
}

func TestUpsertClientSecretRulesSDK(t *testing.T) {
	for _, tc := range clientSecretUpsertCases() {
		t.Run(tc.name, func(t *testing.T) {
			am, srv := createAMServer(t)
			tc.setup(am)
			client := newAMClient(t, srv)

			res, err := client.UpsertClientSecretWithResponse(t.Context(), tc.domainKey, "test", nil, sdk.ClientSecret{Key: tc.secret.Key, Value: tc.secret.Value})
			require.NoError(t, err)

			assert.Equal(t, tc.status, res.StatusCode())
			if tc.status == http.StatusOK {
				assert.Equal(t, &sdk.ClientSecret{Key: tc.secret.Key}, res.JSON200)
			} else {
				assert.Contains(t, string(res.Body), tc.errorContains)
			}
		})
	}
}

func TestDeleteApplicationCascadesClientSecrets(t *testing.T) {
	am, srv := createAMServerWithApplication(t)
	defaultTenant(am).Applications.Put(newChild(Application{Key: "other", Name: "Other", Type: ApplicationTypeWeb, Certificate: "cert", ClientId: new("other")}, defaultDomainKey))
	storeClientSecret(am, testClientSecret())
	defaultTenant(am).ClientSecrets.Put(newChild(testClientSecret(), applicationIdentity(defaultDomainKey, "other")))

	res := httpDelete(t, applicationsURL(srv), "test")
	defer res.Body.Close()
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	assert.Equal(t, []Child[ClientSecret]{newChild(testClientSecret(), applicationIdentity(defaultDomainKey, "other"))}, defaultTenant(am).ClientSecrets.GetAll())
}

func TestDeleteDomainCascadesClientSecrets(t *testing.T) {
	am, srv := createAMServerWithApplication(t)
	defaultTenant(am).Domains.Put(Domain{Key: "other", Name: "Other"})
	defaultTenant(am).Applications.Put(newChild(testApplication(), "other"))
	storeClientSecret(am, testClientSecret())
	defaultTenant(am).ClientSecrets.Put(newChild(testClientSecret(), applicationIdentity("other", "test")))

	res := httpDelete(t, domainsURL(srv), defaultDomainKey)
	defer res.Body.Close()
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	assert.Equal(t, []Child[ClientSecret]{newChild(testClientSecret(), applicationIdentity("other", "test"))}, defaultTenant(am).ClientSecrets.GetAll())
}

func TestDryRunClientSecretNeverReturnsValue(t *testing.T) {
	for _, reject := range []bool{false, true} {
		t.Run(fmt.Sprintf("reject=%t", reject), func(t *testing.T) {
			am, srv := createAMServerWithDryRunReject(t, reject)
			defaultTenant(am).Domains.Put(Domain{Key: defaultDomainKey, Name: "Test"})
			defaultTenant(am).Applications.Put(newChild(testApplication(), defaultDomainKey))

			got := putJSON(t, clientSecretsURL(srv)+"?dryRun=true", jsonObject{"key": "ci", "value": testSecretValue})

			assert.Equal(t, "ci", got["key"])
			assert.NotContains(t, got, "value")
			assert.Equal(t, reject, got["dryRunErrors"] != nil)
			assert.Empty(t, defaultTenant(am).ClientSecrets.GetAll(), "dry run should not persist")
		})
	}
}

func TestListClientSecretsSDK(t *testing.T) {
	am, srv := createAMServerWithApplication(t)
	storeClientSecret(am, testClientSecret())
	client := newAMClient(t, srv)

	res, err := client.ListClientSecretsWithResponse(t.Context(), defaultDomainKey, "test")
	assertSDKOK(t, res, err, []sdk.ClientSecret{{Key: "ci"}})
}

func TestGetClientSecretSDK(t *testing.T) {
	am, srv := createAMServerWithApplication(t)
	storeClientSecret(am, testClientSecret())
	client := newAMClient(t, srv)

	res, err := client.GetClientSecretWithResponse(t.Context(), defaultDomainKey, "test", "ci")
	assertSDKOK(t, res, err, sdk.ClientSecret{Key: "ci"})
}

func TestGetClientSecret404SDK(t *testing.T) {
	_, srv := createAMServerWithApplication(t)
	client := newAMClient(t, srv)

	res, err := client.GetClientSecretWithResponse(t.Context(), defaultDomainKey, "test", "ci")
	assertSDK404(t, res, err)
}

func TestDeleteClientSecretSDK(t *testing.T) {
	am, srv := createAMServerWithApplication(t)
	storeClientSecret(am, testClientSecret())
	client := newAMClient(t, srv)

	del, err := client.DeleteClientSecretWithResponse(t.Context(), defaultDomainKey, "test", "ci")
	assertSDKNoContent(t, del, err)

	get, err := client.GetClientSecretWithResponse(t.Context(), defaultDomainKey, "test", "ci")
	assertSDK404(t, get, err)
}

func TestDryRunClientSecretSDKReturnsDryRunErrorsWithoutValue(t *testing.T) {
	am, srv := createAMServerWithDryRunReject(t, true)
	defaultTenant(am).Domains.Put(Domain{Key: defaultDomainKey, Name: "Test"})
	defaultTenant(am).Applications.Put(newChild(testApplication(), defaultDomainKey))
	client := newAMClient(t, srv)

	res, err := client.UpsertClientSecretWithResponse(t.Context(), defaultDomainKey, "test", &sdk.UpsertClientSecretParams{DryRun: new(true)}, sdk.ClientSecret{Key: "ci", Value: new(testSecretValue)})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())

	assert.Len(t, res.JSON200.DryRunErrors, 1)
	assert.Nil(t, res.JSON200.Value)
	assert.Empty(t, defaultTenant(am).ClientSecrets.GetAll())
}
