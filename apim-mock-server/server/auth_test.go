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
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAuthConfig() auth.Config {
	return auth.Config{
		Users: map[string]auth.UserConfig{
			"admin": {
				Password:    "admin",
				Token:       "admin-token",
				Permissions: nil,
			},
			"reader": {
				Token:       "reader-token",
				Permissions: []string{"API_READ"},
			},
		},
		Permissions: []auth.RoutePermission{
			{
				Path: "/organizations/{orgId}/environments/{envId}/apis",
				Put:  "API_UPDATE",
			},
			{
				Path:   "/organizations/{orgId}/environments/{envId}/apis/{hrid}",
				Get:    "API_READ",
				Delete: "API_DELETE",
			},
		},
	}
}

// createAPIMServerWithAuth starts a mock with testAuthConfig and an API to read.
func createAPIMServerWithAuth(t *testing.T) *httptest.Server {
	t.Helper()
	mock := NewMockAPIM(false)
	srv := httptest.NewServer(NewWithPath(mock, BasePath, auth.NewRegistry(testAuthConfig(), BasePath)))
	t.Cleanup(srv.Close)
	resp := doWithAuth(t, http.MethodPut, collectionURL(srv, "/apis"), apiBody("test"), "Bearer admin-token")
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return srv
}

func doWithAuth(t *testing.T, method, url string, body any, authorization string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		reader = encode(t, body)
	}
	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func basic(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		body          any
		authorization string
		want          int
	}{
		{"bearer OK", http.MethodGet, nil, "Bearer admin-token", http.StatusOK},
		{"basic OK", http.MethodGet, nil, basic("admin", "admin"), http.StatusOK},
		{"no header", http.MethodGet, nil, "", http.StatusUnauthorized},
		{"bad token", http.MethodGet, nil, "Bearer wrong-token", http.StatusUnauthorized},
		{"wrong password", http.MethodGet, nil, basic("admin", "wrong"), http.StatusUnauthorized},
		{"reader can read", http.MethodGet, nil, "Bearer reader-token", http.StatusOK},
		{"reader cannot write", http.MethodPut, apiBody("test"), "Bearer reader-token", http.StatusForbidden},
		{"admin can write", http.MethodPut, apiBody("test"), "Bearer admin-token", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := createAPIMServerWithAuth(t)
			url := collectionURL(srv, "/apis")
			if tt.method == http.MethodGet {
				url += "/test"
			}

			resp := doWithAuth(t, tt.method, url, tt.body, tt.authorization)
			defer resp.Body.Close()
			assert.Equal(t, tt.want, resp.StatusCode)
		})
	}
}

func TestAuth_NoAuthConfig_OpenAccess(t *testing.T) {
	_, srv := createAPIMServer(t)
	putJSON(t, collectionURL(srv, "/apis"), apiBody("test"))

	resp := httpGet(t, collectionURL(srv, "/apis"), "test")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
