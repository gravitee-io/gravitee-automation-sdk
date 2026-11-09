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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const inUse = "You can't delete a certificate with existing applications."

func certificatePath(key string) string {
	return requestPath("/domains/" + defaultDomainKey + "/certificates/" + key)
}

func refuseCertificateDelete(key string) Fault {
	return Fault{Method: http.MethodDelete, Path: certificatePath(key), Status: http.StatusBadRequest, Message: inUse}
}

func createAMServerWithCertificates(t *testing.T, keys ...string) (*MockAM, *httptest.Server, *sdk.AMClient) {
	t.Helper()
	am, srv := createAMServerWithDomain(t)
	for _, key := range keys {
		defaultTenant(am).Certificates.Put(newChild(Certificate{Key: key, Name: new(key)}.WithDefaults(), defaultDomainKey))
	}
	return am, srv, newAMClient(t, srv)
}

func assertDeleteRefused(t *testing.T, client *sdk.AMClient, key string) {
	t.Helper()
	res, err := client.DeleteCertificateWithResponse(t.Context(), defaultDomainKey, key)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode())
	require.NotNil(t, res.JSONDefault)
	assert.Equal(t, sdk.Error{HttpStatus: new(int32(http.StatusBadRequest)), Message: new(inUse)}, *res.JSONDefault)
}

func assertDeleted(t *testing.T, client *sdk.AMClient, key string) {
	t.Helper()
	res, err := client.DeleteCertificateWithResponse(t.Context(), defaultDomainKey, key)
	assertSDKNoContent(t, res, err)
}

func TestFault_AnswersOnlyItsMethodAndPath(t *testing.T) {
	am, _, client := createAMServerWithCertificates(t, "test", "other")
	am.Faults.Set(refuseCertificateDelete("test"))

	assertDeleteRefused(t, client, "test")

	get, err := client.GetCertificateWithResponse(t.Context(), defaultDomainKey, "test")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, get.StatusCode())
	assertDeleted(t, client, "other")
}

func TestFault_MethodIsCaseInsensitive(t *testing.T) {
	am, _, client := createAMServerWithCertificates(t, "test")
	fault := refuseCertificateDelete("test")
	fault.Method = "delete"
	am.Faults.Set(fault)

	assertDeleteRefused(t, client, "test")
}

func TestFault_ClearServesTheStoredResourceAgain(t *testing.T) {
	am, _, client := createAMServerWithCertificates(t, "test")
	am.Faults.Set(refuseCertificateDelete("test"))
	assertDeleteRefused(t, client, "test")

	am.Faults.Clear(http.MethodDelete, certificatePath("test"))

	assertDeleted(t, client, "test")
}

func TestFault_TimesAnswersThatManyRequests(t *testing.T) {
	am, _, client := createAMServerWithCertificates(t, "test")
	fault := refuseCertificateDelete("test")
	fault.Times = 1
	am.Faults.Set(fault)

	assertDeleteRefused(t, client, "test")

	assertDeleted(t, client, "test")
	assert.Empty(t, am.Faults.List())
}

func TestFault_ResetClearsEveryFault(t *testing.T) {
	am, _, client := createAMServerWithCertificates(t, "test", "other")
	am.Faults.Set(refuseCertificateDelete("test"))
	am.Faults.Set(refuseCertificateDelete("other"))

	am.Faults.Reset()

	assert.Empty(t, am.Faults.List())
	assertDeleted(t, client, "test")
	assertDeleted(t, client, "other")
}

func TestFault_SetPanicsOnAnInvalidFault(t *testing.T) {
	valid := refuseCertificateDelete("test")
	for name, mutate := range map[string]func(*Fault){
		"no method":  func(f *Fault) { f.Method = "" },
		"no path":    func(f *Fault) { f.Path = "" },
		"status 200": func(f *Fault) { f.Status = http.StatusOK },
		"status 600": func(f *Fault) { f.Status = 600 },
		"times -1":   func(f *Fault) { f.Times = -1 },
	} {
		t.Run(name, func(t *testing.T) {
			fault := valid
			mutate(&fault)
			assert.Panics(t, func() { (&Faults{}).Set(fault) })
		})
	}
}

func TestFault_AnswersAfterAuth(t *testing.T) {
	am, srv := createAMServerWithAuth(t)
	defaultTenant(am).Domains.Put(Domain{Key: "test", Name: "Test domain", Path: "/test"})
	am.Faults.Set(Fault{Method: http.MethodDelete, Path: requestPath("/domains/test"), Status: http.StatusConflict, Message: "refused"})
	domainURL := domainsURL(srv) + "/test"

	for token, status := range map[string]int{
		"":             http.StatusUnauthorized,
		"reader-token": http.StatusForbidden, // no DOMAIN_DELETE permission
		"admin-token":  http.StatusConflict,
	} {
		req, _ := http.NewRequest(http.MethodDelete, domainURL, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, status, resp.StatusCode, "token %q", token)
	}
}

func faultsURL(srv *httptest.Server) string {
	return srv.URL + FaultsPath
}

func deleteFaults(t *testing.T, srv *httptest.Server, query url.Values) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, faultsURL(srv)+"?"+query.Encode(), nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func putFault(t *testing.T, srv *httptest.Server, fault Fault) *http.Response {
	t.Helper()
	resp := httpPut(t, faultsURL(srv), fault)
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func listFaults(t *testing.T, srv *httptest.Server) []Fault {
	t.Helper()
	resp := httpGetURL(t, faultsURL(srv))
	defer resp.Body.Close()
	failOnNotOK(t, resp)
	return decodeToSliceOf[Fault](t, resp)
}

func TestFaultsEndpoint_PutThenGet(t *testing.T) {
	_, srv, client := createAMServerWithCertificates(t, "test")
	fault := refuseCertificateDelete("test")
	fault.Times = 2

	assert.Equal(t, http.StatusNoContent, putFault(t, srv, fault).StatusCode)

	assert.Equal(t, []Fault{fault}, listFaults(t, srv))
	assertDeleteRefused(t, client, "test")
}

func TestFaultsEndpoint_DeleteWithMethodAndPathClearsOne(t *testing.T) {
	_, srv, client := createAMServerWithCertificates(t, "test", "other")
	putFault(t, srv, refuseCertificateDelete("test"))
	putFault(t, srv, refuseCertificateDelete("other"))

	resp := deleteFaults(t, srv, url.Values{"method": {http.MethodDelete}, "path": {certificatePath("test")}})

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []Fault{refuseCertificateDelete("other")}, listFaults(t, srv))
	assertDeleted(t, client, "test")
}

func TestFaultsEndpoint_DeleteWithoutParametersClearsAll(t *testing.T) {
	_, srv := createAMServer(t)
	putFault(t, srv, refuseCertificateDelete("test"))
	putFault(t, srv, refuseCertificateDelete("other"))

	resp := deleteFaults(t, srv, nil)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Empty(t, listFaults(t, srv))
}

func TestFaultsEndpoint_DeleteWithOneParameterIsRefused(t *testing.T) {
	_, srv := createAMServer(t)
	putFault(t, srv, refuseCertificateDelete("test"))

	resp := deleteFaults(t, srv, url.Values{"method": {http.MethodDelete}})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Len(t, listFaults(t, srv), 1)
}

func TestFaultsEndpoint_PutRefusesAnInvalidFault(t *testing.T) {
	_, srv := createAMServer(t)
	fault := refuseCertificateDelete("test")
	fault.Status = http.StatusOK

	assert.Equal(t, http.StatusBadRequest, putFault(t, srv, fault).StatusCode)
	assert.Empty(t, listFaults(t, srv))
}

func TestFaultsEndpoint_NeedsNoToken(t *testing.T) {
	_, srv := createAMServerWithAuth(t)

	assert.Equal(t, http.StatusNoContent, putFault(t, srv, refuseCertificateDelete("test")).StatusCode)
	assert.Len(t, listFaults(t, srv), 1)
}
