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

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// load writes content to a config file and loads it, away from any .env file and GIOPS_AM_* or GIOPS_GAMMA_* variable.
func load(t *testing.T, content string) (CLIConfig, error) {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	clearEnv(t)
	file := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(file, []byte(content), 0o600))
	return Load(file)
}

// clearEnv unsets the GIOPS_AM_* and GIOPS_GAMMA_* variables for the test.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "GIOPS_AM_") || strings.HasPrefix(name, "GIOPS_GAMMA_") {
			t.Setenv(name, "")
		}
	}
}

func TestLoad_EnvVarsAloneDefineASection(t *testing.T) {
	t.Chdir(t.TempDir())
	clearEnv(t)
	t.Setenv("GIOPS_AM_BASE_URL", "http://localhost:8080/automation")
	t.Setenv("GIOPS_AM_AUTH_BEARER_TOKEN", "env-token")

	c, err := Load(filepath.Join(t.TempDir(), "absent.yaml"))
	require.NoError(t, err)

	ac, err := c.AMContext()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/automation", ac.BaseURL)
	assert.Equal(t, "env-token", *ac.Auth.BearerToken)
	_, err = c.APIMContext()
	assert.ErrorContains(t, err, "no 'gamma' section", "no GIOPS_GAMMA_* variable, no APIM section")
}

func TestLoad_APIMBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		section string
		wantErr string
	}{
		{name: "cloud token without base url", section: "cloud_token: true"},
		{name: "cloud token with base url", section: "cloud_token: true\n  base_url: http://localhost:8083/automation"},
		{name: "no cloud token, no base url", section: "org_id: DEFAULT", wantErr: "required_unless"},
		{name: "not a url", section: "base_url: not-a-url", wantErr: "http_url"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := load(t, "gamma:\n  "+tt.section+"\n  auth:\n    bearer_token: token\n")

			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestLoad_KeepsOrgAndEnvFromTheFile(t *testing.T) {
	c, err := load(t, `
am:
  base_url: http://localhost:8080/automation
  org_id: ACME
  env_id: PROD
  auth:
    bearer_token: token
`)
	require.NoError(t, err)

	ac, err := c.AMContext()
	require.NoError(t, err)
	assert.Equal(t, "ACME", ac.OrgID)
	assert.Equal(t, "PROD", ac.EnvID)
}

func TestLoad_LeavesAbsentOrgAndEnvToTheSDK(t *testing.T) {
	c, err := load(t, `
am:
  base_url: http://localhost:8080/automation
  auth:
    bearer_token: token
`)
	require.NoError(t, err)

	ac, err := c.AMContext()
	require.NoError(t, err)
	assert.Empty(t, ac.OrgID)
	assert.Empty(t, ac.EnvID)
}

func TestLoad_EnvVarsOverrideTheFile(t *testing.T) {
	content := `
am:
  base_url: http://localhost:8080/automation
  org_id: ACME
  auth:
    bearer_token: token
`
	dir := t.TempDir()
	t.Chdir(dir)
	file := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(file, []byte(content), 0o600))
	t.Setenv("GIOPS_AM_ORG_ID", "OTHER")
	t.Setenv("GIOPS_AM_AUTH_BEARER_TOKEN", "env-token")

	c, err := Load(file)
	require.NoError(t, err)

	ac, err := c.AMContext()
	require.NoError(t, err)
	assert.Equal(t, "OTHER", ac.OrgID)
	assert.Equal(t, "env-token", *ac.Auth.BearerToken)
}

func TestLoad_BasicAuthHasNoLengthLimit(t *testing.T) {
	c, err := load(t, `
am:
  base_url: http://localhost:8080/automation
  auth:
    basic_auth:
      username: `+strings.Repeat("u", 100)+`
      password: `+strings.Repeat("p", 200)+`
`)
	require.NoError(t, err)

	ac, err := c.AMContext()
	require.NoError(t, err)
	assert.Len(t, ac.Auth.BasicAuth.Password, 200)
	assert.Nil(t, ac.Auth.BearerToken)
}

func TestAPIMContext_KeepsTheCloudToken(t *testing.T) {
	c := CLIConfig{Gamma: &APIM{BaseURL: "http://apim", CloudToken: true}}

	ac, err := c.APIMContext()

	require.NoError(t, err)
	assert.True(t, ac.CloudToken)
	assert.Equal(t, "http://apim", ac.BaseURL)
}

func TestModuleContext_MissingSection(t *testing.T) {
	_, err := CLIConfig{}.AMContext()
	assert.ErrorContains(t, err, "no 'am' section")

	_, err = CLIConfig{}.APIMContext()
	assert.ErrorContains(t, err, "no 'gamma' section")
}
