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

package sdk

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/apicontext"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_NoAuth(t *testing.T) {
	var ac apicontext.APIContext
	_, err := NewAPIMClient(ac)
	assert.Error(t, err, errors.NoAuthProvided)
}

func TestNewClient_InvalidAuth(t *testing.T) {
	ac := apicontext.APIContext{
		Auth: apicontext.Auth{
			BasicAuth:   &apicontext.BasicAuth{},
			BearerToken: new(""),
		},
	}
	_, err := NewAPIMClient(ac)
	assert.Error(t, err, errors.ManyAuthProvided)
}

func TestNewClient_InvalidURL(t *testing.T) {
	ac := apicontext.APIContext{
		BaseURL: "::",
	}
	_, err := NewAPIMClient(ac)
	assert.Error(t, err)
}

func TestNewClient_Valid(t *testing.T) {
	ac := apicontext.APIContext{
		BaseURL: "http://localhost/automation/",
		Auth:    apicontext.Auth{BearerToken: new("123")},
	}
	client, err := NewAPIMClient(ac)
	assert.NoError(t, err)
	assert.NotNil(t, client.ClientWithResponsesInterface)
}

func TestNewClient_Call(t *testing.T) {

	tests := []struct {
		name               string
		apiContextSupplier func(string) apicontext.APIContext
		expectedAuth       string
		expectedUri        string
	}{
		{
			name: "default with token",
			apiContextSupplier: func(serverURL string) apicontext.APIContext {
				return apicontext.APIContext{
					BaseURL: serverURL + "/automation",
					Auth: apicontext.Auth{
						BearerToken: new("123"),
					},
				}
			},
			expectedAuth: "Bearer 123",
			expectedUri:  "/automation/organizations/DEFAULT/environments/DEFAULT/apis",
		}, {
			name: "default with trailing slash with token",
			apiContextSupplier: func(serverURL string) apicontext.APIContext {
				return apicontext.APIContext{
					BaseURL: serverURL + "/automation/",
					Auth: apicontext.Auth{
						BearerToken: new("123"),
					},
				}
			},
			expectedAuth: "Bearer 123",
			expectedUri:  "/automation/organizations/DEFAULT/environments/DEFAULT/apis",
		}, {
			name: "given org and env with basic",
			apiContextSupplier: func(serverURL string) apicontext.APIContext {
				return apicontext.APIContext{
					BaseURL: serverURL + "/automation",
					OrgID:   "foo",
					EnvID:   "bar",
					Auth: apicontext.Auth{
						BasicAuth: &apicontext.BasicAuth{
							Username: "admin",
							Password: "admin",
						},
					},
				}
			},
			expectedAuth: "Basic " + base64.URLEncoding.EncodeToString([]byte("admin:admin")),
			expectedUri:  "/automation/organizations/foo/environments/bar/apis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var givenAuth string
			var givenUri string
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				givenUri = r.RequestURI
				givenAuth = r.Header.Get("Authorization")
				_, err := w.Write([]byte("[]"))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
			}))
			defer s.Close()
			client, err := NewAPIMClient(tt.apiContextSupplier(s.URL))
			assert.NoError(t, err)
			r, err := client.ListApisWithResponse(context.Background())
			assert.NoError(t, err)
			assert.NotNil(t, r)
			assert.Equal(t, r.StatusCode(), http.StatusOK)
			assert.Equal(t, tt.expectedAuth, givenAuth)
			assert.Equal(t, tt.expectedUri, givenUri)
		})
	}
}

func TestNewClient_WithoutTimeout(t *testing.T) {
	s := slowServer(t, 50*time.Millisecond)
	client, err := NewAPIMClient(bearerContext(s.URL))
	require.NoError(t, err)

	_, err = client.ListApisWithResponse(context.Background())
	assert.NoError(t, err)
}

func TestNewClient_WithTimeout(t *testing.T) {
	s := slowServer(t, 200*time.Millisecond)
	client, err := NewAPIMClient(bearerContext(s.URL))
	require.NoError(t, err)
	client, err = client.WithHTTPClient(&http.Client{Timeout: 10 * time.Millisecond})
	require.NoError(t, err)

	_, err = client.ListApisWithResponse(context.Background())
	assert.ErrorContains(t, err, "Client.Timeout exceeded")
}

func TestNewClient_WithTimeoutAndInvalidAuth(t *testing.T) {
	_, err := NewAPIMClient(apicontext.APIContext{BaseURL: "http://localhost"})
	assert.ErrorIs(t, err, errors.NoAuthProvided)
}

func TestWithHTTPClient_ReturnsNewClientUsingGivenHTTPClient(t *testing.T) {
	var givenAuth string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		givenAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte("[]"))
	}))
	defer s.Close()
	client, err := NewAPIMClient(bearerContext(s.URL))
	require.NoError(t, err)
	transport := &countingTransport{}

	withHTTPClient, err := client.WithHTTPClient(&http.Client{Transport: transport})
	require.NoError(t, err)
	_, err = withHTTPClient.ListApisWithResponse(context.Background())
	require.NoError(t, err)

	assert.NotSame(t, client, withHTTPClient)
	assert.Equal(t, 1, transport.calls)
	assert.Equal(t, "Bearer 123", givenAuth)

	_, err = client.ListApisWithResponse(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, transport.calls, "original client must keep its own HTTP client")
}

type countingTransport struct {
	calls int
}

func (c *countingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.calls++
	return http.DefaultTransport.RoundTrip(r)
}

func bearerContext(serverURL string) apicontext.APIContext {
	return apicontext.APIContext{
		BaseURL: serverURL + "/automation",
		Auth:    apicontext.Auth{BearerToken: new("123")},
	}
}

func slowServer(t *testing.T, delay time.Duration) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(delay)
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(s.Close)
	return s
}
