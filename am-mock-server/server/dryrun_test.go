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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDryRunPut_NoRejectReturnsEmptyAndDoesNotPersist(t *testing.T) {
	am, srv := createAMServer(t)
	body := Domain{Key: "test", Name: "Test domain", Path: "/test"}

	resp := httpPut(t, domainsURL(srv)+"?dryRun=true", body)
	defer resp.Body.Close()
	failOnNotOK(t, resp)

	domain := decodeTo[Domain](t, resp)
	assert.Nil(t, domain.CreatedAt, "dry run should not persist")

	_, exists := defaultTenant(am).Domains.Get("test")
	assert.False(t, exists)
}

func TestDryRunPut_DoesNotPersist(t *testing.T) {
	am, srv := createAMServerWithDryRunReject(t, true)
	body := Domain{Key: "test", Name: "Test domain", Path: "/test"}

	resp := httpPut(t, domainsURL(srv)+"?dryRun=true", body)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	got := decodeTo[Domain](t, resp)
	assert.Equal(t, body.Name, got.Name)
	assert.Equal(t, &dryRunErrors, got.DryRunErrors)
	_, exists := defaultTenant(am).Domains.Get("test")
	assert.False(t, exists)
	assertGet404(t, domainsURL(srv), "test", "Domain")
}

func TestDryRunPut_FalseStillPersists(t *testing.T) {
	am, srv := createAMServerWithDryRunReject(t, true)
	body := Domain{Key: "test", Name: "Test domain", Path: "/test"}.WithDefaults()

	resp := httpPut(t, domainsURL(srv)+"?dryRun=false", body)
	defer resp.Body.Close()
	failOnNotOK(t, resp)

	got, exists := defaultTenant(am).Domains.Get("test")
	require.True(t, exists)
	assertEqualIgnoringTimestamps(t, body, got)
}

func TestDryRunPut_LeavesExistingUnchanged(t *testing.T) {
	am, srv := createAMServerWithDryRunReject(t, true)
	existing := Domain{Key: "test", Name: "Original", Path: "/test"}
	defaultTenant(am).Domains.Put(existing)

	resp := httpPut(t, domainsURL(srv)+"?dryRun=true", Domain{Key: "test", Name: "Changed", Path: "/test"})
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, &dryRunErrors, decodeTo[Domain](t, resp).DryRunErrors)

	got, exists := defaultTenant(am).Domains.Get("test")
	require.True(t, exists)
	assert.Equal(t, existing, got)
}

var nestedDryRunCases = []struct {
	name   string
	url    func(*httptest.Server) string
	body   any
	stored func(*MockAM) int
}{
	{"application", applicationsURL, testApplication(), func(am *MockAM) int { return len(defaultTenant(am).Applications.GetAll()) }},
	{"certificate", certificatesURL, testCertificate(), func(am *MockAM) int { return len(defaultTenant(am).Certificates.GetAll()) }},
	{"identity provider", identitiesURL, testIdentityProvider(), func(am *MockAM) int { return len(defaultTenant(am).IdentityProviders.GetAll()) }},
	{"reporter", reportersURL, testReporter(), func(am *MockAM) int { return len(defaultTenant(am).Reporters.GetAll()) }},
}

func TestDryRunPut_NestedNoRejectReturnsPayloadAndDoesNotPersist(t *testing.T) {
	for _, tc := range nestedDryRunCases {
		t.Run(tc.name, func(t *testing.T) {
			am, srv := createAMServerWithDomain(t)

			resp := httpPut(t, tc.url(srv)+"?dryRun=true", tc.body)
			defer resp.Body.Close()
			failOnNotOK(t, resp)

			got, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.JSONEq(t, encode(t, tc.body).String(), string(got))
			assert.Zero(t, tc.stored(am), "dry run should not persist")
		})
	}
}

func TestDryRunPut_NestedDoesNotPersist(t *testing.T) {
	for _, tc := range nestedDryRunCases {
		t.Run(tc.name, func(t *testing.T) {
			am, srv := createAMServerWithDryRunReject(t, true)
			defaultTenant(am).Domains.Put(Domain{Key: defaultDomainKey, Name: "Test"})

			resp := httpPut(t, tc.url(srv)+"?dryRun=true", tc.body)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, dryRunErrors, decodeTo[struct{ DryRunErrors []DryRunError }](t, resp).DryRunErrors)
			assert.Zero(t, tc.stored(am))
		})
	}
}

func TestDryRunPut_NestedFalseStillPersists(t *testing.T) {
	for _, tc := range nestedDryRunCases {
		t.Run(tc.name, func(t *testing.T) {
			am, srv := createAMServerWithDryRunReject(t, true)
			defaultTenant(am).Domains.Put(Domain{Key: defaultDomainKey, Name: "Test"})

			resp := httpPut(t, tc.url(srv)+"?dryRun=false", tc.body)
			defer resp.Body.Close()
			failOnNotOK(t, resp)

			assert.Equal(t, 1, tc.stored(am))
		})
	}
}

func TestDryRunPut_SDKNestedReturnsDryRunErrors(t *testing.T) {
	am, srv := createAMServerWithDryRunReject(t, true)
	defaultTenant(am).Domains.Put(Domain{Key: defaultDomainKey, Name: "Test"})
	client := newAMClient(t, srv)
	dryRun := new(true)

	app, err := client.UpsertApplicationWithResponse(t.Context(), defaultDomainKey, &sdk.UpsertApplicationParams{DryRun: dryRun}, testApplicationSDK())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, app.StatusCode())
	assert.Len(t, app.JSON200.DryRunErrors, 1)
	cert, err := client.UpsertCertificateWithResponse(t.Context(), defaultDomainKey, &sdk.UpsertCertificateParams{DryRun: dryRun}, testCertificateSDK())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, cert.StatusCode())
	assert.Len(t, cert.JSON200.DryRunErrors, 1)
	idp, err := client.UpsertIdentityProviderWithResponse(t.Context(), defaultDomainKey, &sdk.UpsertIdentityProviderParams{DryRun: dryRun}, testIdentityProviderSDK())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, idp.StatusCode())
	assert.Len(t, idp.JSON200.DryRunErrors, 1)
	rep, err := client.UpsertReporterWithResponse(t.Context(), defaultDomainKey, &sdk.UpsertReporterParams{DryRun: dryRun}, testReporterSDK())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rep.StatusCode())
	assert.Len(t, rep.JSON200.DryRunErrors, 1)

	for _, tc := range nestedDryRunCases {
		assert.Zero(t, tc.stored(am), tc.name)
	}
}
