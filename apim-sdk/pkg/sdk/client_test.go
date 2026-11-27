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

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/apicontext"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_NoAuth(t *testing.T) {
	var ac APIContext
	_, err := NewAPIMClient(ac)
	assert.ErrorIs(t, err, errors.NoAuthProvided)
}

func TestNewClient_InvalidAuth(t *testing.T) {
	ac := APIContext{APIContext: apicontext.APIContext{
		Auth: apicontext.Auth{
			BasicAuth:   &apicontext.BasicAuth{},
			BearerToken: new(""),
		},
	}}
	_, err := NewAPIMClient(ac)
	assert.ErrorIs(t, err, errors.ManyAuthProvided)
}

func TestNewClient_InvalidURL(t *testing.T) {
	ac := APIContext{APIContext: apicontext.APIContext{
		BaseURL: "::",
	}}
	_, err := NewAPIMClient(ac)
	assert.Error(t, err)
}

func TestNewClient_Valid(t *testing.T) {
	ac := APIContext{APIContext: apicontext.APIContext{
		BaseURL: "http://localhost/automation/",
		Auth:    apicontext.Auth{BearerToken: new("123")},
	}}
	client, err := NewAPIMClient(ac)
	assert.NoError(t, err)
	assert.NotNil(t, client.ClientWithResponsesInterface)
}

func TestNewClient_Call(t *testing.T) {
	tests := []struct {
		name               string
		apiContextSupplier func(string) APIContext
		expectedAuth       string
		expectedUri        string
	}{
		{
			name: "default with token",
			apiContextSupplier: func(serverURL string) APIContext {
				return bearerContext(serverURL)
			},
			expectedAuth: "Bearer 123",
			expectedUri:  "/automation/organizations/DEFAULT/environments/DEFAULT/apis/x",
		}, {
			name: "default with trailing slash with token",
			apiContextSupplier: func(serverURL string) APIContext {
				ac := bearerContext(serverURL)
				ac.BaseURL += "/"
				return ac
			},
			expectedAuth: "Bearer 123",
			expectedUri:  "/automation/organizations/DEFAULT/environments/DEFAULT/apis/x",
		}, {
			name: "given org and env with basic",
			apiContextSupplier: func(serverURL string) APIContext {
				return APIContext{APIContext: apicontext.APIContext{
					BaseURL: serverURL + "/automation",
					OrgID:   "O",
					EnvID:   "E",
					Auth: apicontext.Auth{
						BasicAuth: &apicontext.BasicAuth{
							Username: "admin",
							Password: "admin",
						},
					},
				}}
			},
			expectedAuth: "Basic " + base64.URLEncoding.EncodeToString([]byte("admin:admin")),
			expectedUri:  "/automation/organizations/O/environments/E/apis/x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var givenAuth string
			var givenUri string
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				givenUri = r.RequestURI
				givenAuth = r.Header.Get("Authorization")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"hrid":"x"}`))
			}))
			defer s.Close()
			client, err := NewAPIMClient(tt.apiContextSupplier(s.URL))
			require.NoError(t, err)
			r, err := client.GetApiWithResponse(context.Background(), "x", nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, r.StatusCode())
			assert.Equal(t, tt.expectedAuth, givenAuth)
			assert.Equal(t, tt.expectedUri, givenUri)
		})
	}
}

func TestWithHTTPClient_ReturnsNewClientUsingGivenHTTPClient(t *testing.T) {
	var givenAuth string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		givenAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte("{}"))
	}))
	defer s.Close()
	client, err := NewAPIMClient(bearerContext(s.URL))
	require.NoError(t, err)
	transport := &countingTransport{}

	withHTTPClient, err := client.WithHTTPClient(&http.Client{Transport: transport})
	require.NoError(t, err)
	_, err = withHTTPClient.GetApiWithResponse(context.Background(), "x", nil)
	require.NoError(t, err)

	assert.NotSame(t, client, withHTTPClient)
	assert.Equal(t, 1, transport.calls)
	assert.Equal(t, "Bearer 123", givenAuth)

	_, err = client.GetApiWithResponse(context.Background(), "x", nil)
	require.NoError(t, err)
	assert.Equal(t, 1, transport.calls, "original client must keep its own HTTP client")
}

func TestUpsert_DecodesErrorDetailsWhateverTheirInvalidValue(t *testing.T) {
	tests := []struct {
		name         string
		invalidValue string
		want         interface{}
	}{
		{"raw value, as the spec example", `"foo"`, "foo"},
		{"JsonNullable wrapper, as APIM sends today", `{"present":true,"undefined":false}`,
			map[string]interface{}{"present": true, "undefined": false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"httpStatus":400,"message":"Validation error",` +
					`"details":[{"message":"must not be blank","invalidValue":` + tt.invalidValue + `}]}`))
			}))
			defer s.Close()
			client, err := NewAPIMClient(bearerContext(s.URL))
			require.NoError(t, err)

			r, err := client.UpsertGroupWithResponse(context.Background(), nil, Group{Hrid: "developers"})

			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, r.StatusCode())
			require.NotNil(t, r.JSON400)
			assert.Equal(t, new("Validation error"), r.JSON400.Message)
			assert.Equal(t, tt.want, r.JSON400.Details[0].InvalidValue)
		})
	}
}

// GKO sends hridContainsUUID=true for shared policy groups created before the Automation API, whose hrid is their id.
func TestSharedPolicyGroup_SendsHridContainsUUIDOnlyWhenSet(t *testing.T) {
	calls := map[string]func(c *APIMClient, flag *bool) error{
		"upsert": func(c *APIMClient, flag *bool) error {
			_, err := c.UpsertSharedPolicyGroupWithResponse(context.Background(),
				&UpsertSharedPolicyGroupParams{HridContainsUUID: flag}, SharedPolicyGroup{Hrid: "x"})
			return err
		},
		"get": func(c *APIMClient, flag *bool) error {
			_, err := c.GetSharedPolicyGroupWithResponse(context.Background(), "x",
				&GetSharedPolicyGroupParams{HridContainsUUID: flag})
			return err
		},
		"delete": func(c *APIMClient, flag *bool) error {
			_, err := c.DeleteSharedPolicyGroupWithResponse(context.Background(), "x",
				&DeleteSharedPolicyGroupParams{HridContainsUUID: flag})
			return err
		},
	}
	for name, call := range calls {
		for _, tt := range []struct {
			name string
			flag *bool
			want []string
		}{
			{"set", new(true), []string{"true"}},
			{"nil", nil, nil},
		} {
			t.Run(name+" "+tt.name, func(t *testing.T) {
				var query map[string][]string
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					query = r.URL.Query()
					w.WriteHeader(http.StatusNoContent)
				}))
				defer s.Close()
				client, err := NewAPIMClient(bearerContext(s.URL))
				require.NoError(t, err)

				require.NoError(t, call(client, tt.flag))

				assert.Equal(t, tt.want, query["hridContainsUUID"])
			})
		}
	}
}

type countingTransport struct {
	calls int
}

func (c *countingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.calls++
	return http.DefaultTransport.RoundTrip(r)
}

func bearerContext(serverURL string) APIContext {
	return APIContext{APIContext: apicontext.APIContext{
		BaseURL: serverURL + "/automation",
		Auth:    apicontext.Auth{BearerToken: new("123")},
	}}
}

func TestProbe(t *testing.T) {
	groupPath := "/automation/organizations/DEFAULT/environments/DEFAULT/groups/" + probeHrid
	probeNotFound := `{"httpStatus":404,"message":"HRID [` + probeHrid + `] cannot be found."}`

	tests := []struct {
		name          string
		openAPIStatus int
		groupStatus   int
		groupBody     string
		wantStatus    int // 0: Probe succeeds
	}{
		{name: "probe group not found", openAPIStatus: 200, groupStatus: 404, groupBody: probeNotFound},
		{name: "probe group found", openAPIStatus: 200, groupStatus: 200, groupBody: `{"hrid":"` + probeHrid + `"}`},
		{name: "not an APIM Automation API", openAPIStatus: 404, wantStatus: 404},
		{name: "OpenAPI document refused by a gateway", openAPIStatus: 403, wantStatus: 403},
		{name: "bad credentials", openAPIStatus: 200, groupStatus: 401, groupBody: `{"httpStatus":401}`, wantStatus: 401},
		{name: "no access to the environment", openAPIStatus: 200, groupStatus: 403, groupBody: `{"httpStatus":403}`, wantStatus: 403},
		{name: "unknown environment", openAPIStatus: 200, groupStatus: 404, groupBody: `{"httpStatus":404,"message":"Environment [DEFAULT] cannot be found."}`, wantStatus: 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/automation/open-api.yaml":
					assert.Equal(t, "Bearer 123", r.Header.Get("Authorization"), "a gateway in front of APIM may need the credentials")
					w.WriteHeader(tt.openAPIStatus)
				case groupPath:
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.groupStatus)
					_, _ = w.Write([]byte(tt.groupBody))
				default:
					t.Errorf("unexpected request to %s", r.URL.Path)
				}
			}))
			defer s.Close()
			client, err := NewAPIMClient(bearerContext(s.URL))
			require.NoError(t, err)

			err = client.Probe(context.Background())

			if tt.wantStatus == 0 {
				assert.NoError(t, err)
				return
			}
			var httpErr errors.HttpError
			require.ErrorAs(t, err, &httpErr)
			assert.Equal(t, tt.wantStatus, httpErr.Status)
		})
	}
}

func TestProbe_ServerUnreachable(t *testing.T) {
	s := httptest.NewServer(http.NotFoundHandler())
	s.Close()
	client, err := NewAPIMClient(bearerContext(s.URL))
	require.NoError(t, err)

	err = client.Probe(context.Background())

	assert.ErrorAs(t, err, new(errors.ClientError))
}
