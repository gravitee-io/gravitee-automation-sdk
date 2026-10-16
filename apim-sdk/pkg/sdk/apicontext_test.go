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
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/apicontext"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	oneEnv := cloudToken(t, jwt.MapClaims{"org": "acme", "envs": []string{"prod"}, "cpg": "eu"})
	twoEnvs := cloudToken(t, jwt.MapClaims{"org": "acme", "envs": []string{"dev", "prod"}, "cpg": "eu"})

	tests := []struct {
		name    string
		given   APIContext
		want    apicontext.APIContext
		wantErr string
	}{
		{
			name: "not cloud is unchanged",
			given: APIContext{APIContext: apicontext.APIContext{
				BaseURL: "http://localhost:8083/automation", OrgID: "O", EnvID: "E", Auth: bearer("123"),
			}},
			want: apicontext.APIContext{
				BaseURL: "http://localhost:8083/automation", OrgID: "O", EnvID: "E", Auth: bearer("123"),
			},
		},
		{
			name:  "one env is inferred and the URL built from cpg",
			given: cloud(apicontext.APIContext{Auth: bearer(oneEnv)}),
			want: apicontext.APIContext{
				BaseURL: "https://eu.cloudgate.gravitee.io/apim/automation", OrgID: "acme", EnvID: "prod", Auth: bearer(oneEnv),
			},
		},
		{
			name:  "base URL is kept",
			given: cloud(apicontext.APIContext{BaseURL: "https://apim.local/automation", Auth: bearer(oneEnv)}),
			want: apicontext.APIContext{
				BaseURL: "https://apim.local/automation", OrgID: "acme", EnvID: "prod", Auth: bearer(oneEnv),
			},
		},
		{
			name:  "org is overwritten",
			given: cloud(apicontext.APIContext{OrgID: "other", Auth: bearer(oneEnv)}),
			want: apicontext.APIContext{
				BaseURL: "https://eu.cloudgate.gravitee.io/apim/automation", OrgID: "acme", EnvID: "prod", Auth: bearer(oneEnv),
			},
		},
		{
			name:  "given env present in the token",
			given: cloud(apicontext.APIContext{EnvID: "dev", Auth: bearer(twoEnvs)}),
			want: apicontext.APIContext{
				BaseURL: "https://eu.cloudgate.gravitee.io/apim/automation", OrgID: "acme", EnvID: "dev", Auth: bearer(twoEnvs),
			},
		},
		{
			name:    "given env absent from the token",
			given:   cloud(apicontext.APIContext{EnvID: "staging", Auth: bearer(twoEnvs)}),
			wantErr: `does not grant env "staging"`,
		},
		{
			name:    "several envs and no env",
			given:   cloud(apicontext.APIContext{Auth: bearer(twoEnvs)}),
			wantErr: "grants several envs [dev prod]",
		},
		{
			name:    "missing claim",
			given:   cloud(apicontext.APIContext{Auth: bearer(cloudToken(t, jwt.MapClaims{"org": "acme", "envs": []string{"prod"}}))}),
			wantErr: "no cpg claim",
		},
		{
			name:    "malformed token",
			given:   cloud(apicontext.APIContext{Auth: bearer("not-a-jwt")}),
			wantErr: "invalid cloud token",
		},
		{
			name: "basic auth",
			given: cloud(apicontext.APIContext{Auth: apicontext.Auth{
				BearerToken: new(oneEnv), BasicAuth: &apicontext.BasicAuth{Username: "admin", Password: "admin"},
			}}),
			wantErr: "cloud token requires bearer auth only",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.given.Resolve()
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.ErrorAs(t, err, new(errors.ClientError))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewClient_CloudTokenServerURL(t *testing.T) {
	token := cloudToken(t, jwt.MapClaims{"org": "acme", "envs": []string{"prod"}, "cpg": "eu"})
	client, err := NewAPIMClient(cloud(apicontext.APIContext{Auth: bearer(token)}))
	require.NoError(t, err)
	assert.Equal(t, "https://eu.cloudgate.gravitee.io/apim/automation/organizations/acme/environments/prod", client.server)
}

func cloud(ac apicontext.APIContext) APIContext {
	return APIContext{APIContext: ac, CloudToken: true}
}

func bearer(token string) apicontext.Auth {
	return apicontext.Auth{BearerToken: new(token)}
}

func cloudToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("k"))
	require.NoError(t, err)
	return token
}
