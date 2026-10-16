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

	"github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/apicontext"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Bodies are raw JSON so the tests prove what goes over the wire, not what a Go type decodes to.
type jsonObject = map[string]any

const (
	apiHrid         = "api"
	planHrid        = "keyless"
	applicationHrid = "app"
	portalHrid      = "portal"
)

func createAPIMServer(t *testing.T) (*MockAPIM, *httptest.Server) {
	t.Helper()
	return createAPIMServerWithDryRunReject(t, false)
}

func createAPIMServerWithDryRunReject(t *testing.T, dryRunReject bool) (*MockAPIM, *httptest.Server) {
	t.Helper()
	mock := NewMockAPIM(dryRunReject)
	srv := httptest.NewServer(New(mock))
	t.Cleanup(srv.Close)
	return mock, srv
}

// createAPIMServerWithParents also creates the API (with plan planHrid), the application and the portal
// that sub-resources hang off.
func createAPIMServerWithParents(t *testing.T) (*MockAPIM, *httptest.Server) {
	t.Helper()
	mock, srv := createAPIMServer(t)
	putJSON(t, collectionURL(srv, "/apis"), apiBody(apiHrid))
	putJSON(t, collectionURL(srv, "/applications"), applicationBody(applicationHrid))
	putJSON(t, collectionURL(srv, "/portals"), portalBody(portalHrid))
	return mock, srv
}

func newTestClient(t *testing.T, srv *httptest.Server, orgID, envID, token string) *sdk.APIMClient {
	t.Helper()
	client, err := sdk.NewAPIMClient(sdk.APIContext{APIContext: apicontext.APIContext{
		BaseURL: srv.URL + BasePath,
		OrgID:   orgID,
		EnvID:   envID,
		Auth:    apicontext.Auth{BearerToken: new(token)},
	}})
	require.NoError(t, err)
	return client
}

func newAPIMClient(t *testing.T, srv *httptest.Server) *sdk.APIMClient {
	t.Helper()
	return newTestClient(t, srv, "", "", "test")
}

func collectionURL(srv *httptest.Server, resourcePath string) string {
	return orgEnvURL(srv, "DEFAULT", "DEFAULT", resourcePath)
}

func orgEnvURL(srv *httptest.Server, org, env, resourcePath string) string {
	return srv.URL + BasePath + "/organizations/" + org + "/environments/" + env + resourcePath
}

// Smallest valid body of each kind.

func apiBody(hrid string) jsonObject {
	return jsonObject{
		"hrid": hrid, "name": "API", "type": "PROXY", "version": "1.0", "lifecycleState": "CREATED",
		"listeners": []any{jsonObject{
			"type":        "HTTP",
			"paths":       []any{jsonObject{"path": "/" + hrid}},
			"entrypoints": []any{jsonObject{"type": "http-proxy"}},
		}},
		"endpointGroups": []any{jsonObject{
			"type":      "http-proxy",
			"endpoints": []any{jsonObject{"name": "default", "type": "http-proxy"}},
		}},
		"plans": []any{jsonObject{
			"hrid": planHrid, "name": "Keyless", "status": "PUBLISHED", "type": "API", "mode": "STANDARD",
			"validation": "AUTO", "security": jsonObject{"type": "KEY_LESS"},
		}},
	}
}

func applicationBody(hrid string) jsonObject {
	return jsonObject{"hrid": hrid, "name": "Application", "description": "An application"}
}

func groupBody(hrid string) jsonObject {
	return jsonObject{"hrid": hrid, "name": "Group"}
}

func dictionaryBody(hrid string) jsonObject {
	return jsonObject{
		"hrid": hrid, "name": "Dictionary", "type": "MANUAL", "deployed": false,
		"manual": jsonObject{"properties": jsonObject{"key": "value"}},
	}
}

func sharedPolicyGroupBody(hrid string) jsonObject {
	return jsonObject{"hrid": hrid, "name": "Shared policy group", "apiType": "PROXY", "phase": "REQUEST"}
}

func portalBody(hrid string) jsonObject {
	return jsonObject{"hrid": hrid, "name": "Portal"}
}

func portalThemeBody(hrid string) jsonObject {
	return jsonObject{"hrid": hrid, "name": "Theme", "definition": jsonObject{}}
}

func aimCatalogMcpServerBody(hrid string) jsonObject {
	return jsonObject{
		"hrid": hrid, "entityId": "mcp-server." + hrid,
		"connection": jsonObject{"endpoint": "https://mcp.example.com/mcp", "transport": "HTTP"},
	}
}

func aimMcpProxyBody(hrid string) jsonObject {
	return jsonObject{
		"hrid": hrid, "entityId": "mcp-proxy." + hrid, "name": "MCP proxy", "contextPath": "/" + hrid, "mode": "PROXY",
		"plans": []any{jsonObject{"name": "Keyless", "security": jsonObject{"type": "KEY_LESS"}}},
	}
}

func subscriptionBody(hrid string) jsonObject {
	return jsonObject{"hrid": hrid, "applicationHrid": applicationHrid, "planHrid": planHrid}
}

func documentationBody(hrid string) jsonObject {
	return jsonObject{"hrid": hrid, "name": "Documentation", "type": "GRAVITEE_MARKDOWN", "content": "# Hello"}
}

func linkBody(hrid string) jsonObject {
	return jsonObject{"hrid": hrid, "name": "Link", "href": "https://www.gravitee.io"}
}

func listingBody(hrid string) jsonObject {
	return jsonObject{"hrid": hrid, "apis": []any{jsonObject{"apiHrid": apiHrid, "location": "/apis"}}}
}

// as decodes a raw body into an SDK type.
func as[T any](t *testing.T, body jsonObject) T {
	t.Helper()
	var typed T
	require.NoError(t, json.Unmarshal(encode(t, body).Bytes(), &typed))
	return typed
}

func httpGet(t *testing.T, url, hrid string) *http.Response {
	t.Helper()
	resp, err := http.Get(url + "/" + hrid)
	require.NoError(t, err)
	return resp
}

func httpPut(t *testing.T, url string, body any) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodPut, url, encode(t, body))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	return resp
}

func httpDelete(t *testing.T, url string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	return resp
}

func encode(t *testing.T, object any) *bytes.Buffer {
	t.Helper()
	body := &bytes.Buffer{}
	require.NoError(t, json.NewEncoder(body).Encode(object))
	return body
}

func decodeTo[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	target := new(T)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(target))
	return *target
}

func failOnNotOK(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}
}

func putJSON(t *testing.T, url string, body jsonObject) jsonObject {
	t.Helper()
	resp := httpPut(t, url, body)
	defer resp.Body.Close()
	failOnNotOK(t, resp)
	return decodeTo[jsonObject](t, resp)
}

func getJSON(t *testing.T, url, hrid string) jsonObject {
	t.Helper()
	resp := httpGet(t, url, hrid)
	defer resp.Body.Close()
	failOnNotOK(t, resp)
	return decodeTo[jsonObject](t, resp)
}

func assertHridNotFound(t *testing.T, resp *http.Response, hrid string) {
	t.Helper()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "HRID ["+hrid+"] cannot be found.", decodeTo[jsonObject](t, resp)["message"])
}

func assertGet404(t *testing.T, url, hrid string) {
	t.Helper()
	assertHridNotFound(t, httpGet(t, url, hrid), hrid)
}

func assertDeleteGone(t *testing.T, url, hrid string) {
	t.Helper()
	resp := httpDelete(t, url+"/"+hrid)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assertGet404(t, url, hrid)
}

func assertSDKOK(t *testing.T, resp response.StatusCoder, err error) any {
	t.Helper()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	payload, ok := response.Payload[any](resp)
	require.True(t, ok)
	return payload
}

func assertSDK404(t *testing.T, resp response.StatusCoder, err error) {
	t.Helper()
	assert.True(t, response.IsNotFound(resp, err))
}

func assertSDKNoContent(t *testing.T, resp response.StatusCoder, err error) {
	t.Helper()
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode())
}
