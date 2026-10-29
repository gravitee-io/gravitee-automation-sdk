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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/apicontext"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultDomainKey = "test"

func defaultTenant(am *MockAM) *tenant {
	return am.getTenant(orgEnv{"DEFAULT", "DEFAULT"})
}

func createAMServer(t *testing.T) (*MockAM, *httptest.Server) {
	t.Helper()
	return createAMServerWithDryRunReject(t, false)
}

func createAMServerWithDryRunReject(t *testing.T, dryRunReject bool) (*MockAM, *httptest.Server) {
	t.Helper()
	am := NewMockAM()
	srv := httptest.NewServer(NewWithPath(am, BasePath, nil, dryRunReject))
	t.Cleanup(srv.Close)
	return am, srv
}

func createAMServerWithDomain(t *testing.T) (*MockAM, *httptest.Server) {
	t.Helper()
	am, srv := createAMServer(t)
	defaultTenant(am).Domains.Put(Domain{Key: defaultDomainKey, Name: "Test"})
	return am, srv
}

func newTestClient(t *testing.T, srv *httptest.Server, orgID, envID, token string) *sdk.AMClient {
	t.Helper()
	client, err := sdk.NewAMClient(apicontext.APIContext{
		BaseURL: srv.URL + BasePath,
		OrgID:   orgID,
		EnvID:   envID,
		Auth:    apicontext.Auth{BearerToken: new(token)},
	})
	require.NoError(t, err)
	return client
}

func newAMClient(t *testing.T, srv *httptest.Server) *sdk.AMClient {
	t.Helper()
	return newTestClient(t, srv, "", "", "test")
}

func collectionURL(srv *httptest.Server, resourcePath string) string {
	return srv.URL + BasePath + "/organizations/DEFAULT/environments/DEFAULT" + resourcePath
}

func domainsURL(srv *httptest.Server) string {
	return collectionURL(srv, "/domains")
}

func dataPlanesURL(srv *httptest.Server) string {
	return collectionURL(srv, "/dataplanes")
}

func applicationsURL(srv *httptest.Server) string {
	return collectionURL(srv, "/domains/"+defaultDomainKey+"/applications")
}

func certificatesURL(srv *httptest.Server) string {
	return collectionURL(srv, "/domains/"+defaultDomainKey+"/certificates")
}

func identitiesURL(srv *httptest.Server) string {
	return collectionURL(srv, "/domains/"+defaultDomainKey+"/identities")
}

func reportersURL(srv *httptest.Server) string {
	return collectionURL(srv, "/domains/"+defaultDomainKey+"/reporters")
}

func httpGet(t *testing.T, url, key string) *http.Response {
	t.Helper()
	return httpGetURL(t, url+"/"+key)
}

func httpGetURL(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func httpPut(t *testing.T, url string, body any) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodPut, url, encode(t, body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func httpDelete(t *testing.T, url, key string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodDelete, url+"/"+key, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func encode(t *testing.T, object any) *bytes.Buffer {
	t.Helper()
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(object); err != nil {
		t.Fatal(err)
	}
	return body
}

func decodeToSliceOf[T any](t *testing.T, resp *http.Response) []T {
	t.Helper()
	target := make([]T, 0)
	if err := json.NewDecoder(resp.Body).Decode(&target); err != nil {
		t.Fatal(err)
	}
	return target
}

func decodeTo[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	target := new(T)
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
	return *target
}

func failOnNotOK(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}
}

func assertListEqual[T any](t *testing.T, url string, expected []T) {
	t.Helper()
	res := httpGetURL(t, url)
	defer res.Body.Close()
	failOnNotOK(t, res)
	assertEqualIgnoringTimestamps(t, expected, decodeToSliceOf[T](t, res))
}

func assertGetEqual[T any](t *testing.T, url, key string, expected T) {
	t.Helper()
	resp := httpGet(t, url, key)
	defer resp.Body.Close()
	failOnNotOK(t, resp)
	assertEqualIgnoringTimestamps(t, expected, decodeTo[T](t, resp))
}

func assertGet404(t *testing.T, url, key, kind string) {
	t.Helper()
	resp := httpGet(t, url, key)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, notFoundError(kind, key), decodeTo[Error](t, resp))
}

func assertPutEqual[T any](t *testing.T, url string, body T) {
	t.Helper()
	resp := httpPut(t, url, body)
	defer resp.Body.Close()
	failOnNotOK(t, resp)
	assertEqualIgnoringTimestamps(t, body, decodeTo[T](t, resp))
}

func assertDeleteGone(t *testing.T, url, key, kind string) {
	t.Helper()
	res := httpDelete(t, url, key)
	defer res.Body.Close()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
	assertGet404(t, url, key, kind)
}

func assertSDKOK[T any](t *testing.T, resp response.StatusCoder, err error, expected T) {
	t.Helper()
	assert.False(t, response.IsNetworkError(resp, err))
	got, ok := response.Payload[T](resp)
	require.True(t, ok)
	assertEqualIgnoringTimestamps(t, expected, got)
}

// assertEqualIgnoringTimestamps compares expected and actual as JSON, without the server-set createdAt/updatedAt.
func assertEqualIgnoringTimestamps(t *testing.T, expected, actual any) {
	t.Helper()
	assert.Equal(t, withoutTimestamps(t, expected), withoutTimestamps(t, actual))
}

func withoutTimestamps(t *testing.T, v any) any {
	t.Helper()
	var generic any
	require.NoError(t, json.Unmarshal(encode(t, v).Bytes(), &generic))
	return stripTimestamps(generic)
}

func stripTimestamps(v any) any {
	switch v := v.(type) {
	case map[string]any:
		delete(v, "createdAt")
		delete(v, "updatedAt")
		for k, child := range v {
			v[k] = stripTimestamps(child)
		}
	case []any:
		for i, child := range v {
			v[i] = stripTimestamps(child)
		}
	}
	return v
}

func assertSDK404(t *testing.T, resp response.StatusCoder, err error) {
	t.Helper()
	assert.True(t, response.IsNotFound(resp, err))
}

func assertSDKNoContent(t *testing.T, resp response.StatusCoder, err error) {
	t.Helper()
	assert.False(t, response.IsNetworkError(resp, err))
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode())
}
