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
	"errors"

	apimsdk "github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/apicontext"
)

// CLIConfig is the giops configuration, one section per module. The file sets it first, then env vars
// override it (GIOPS_AM_BASE_URL, GIOPS_GAMMA_AUTH_BEARER_TOKEN, ...). Empty org and env are left to the SDKs.
type CLIConfig struct {
	// AM is the am module's connection. Nil when the file has no am section.
	AM *AM `json:"am" envPrefix:"GIOPS_AM_" validate:"required_without=Gamma"`
	// Gamma is the APIM connection the apim module uses. Nil when the file has no gamma section.
	Gamma *APIM `json:"gamma" envPrefix:"GIOPS_GAMMA_" validate:"required_without=AM"`
}

// AM is the connection to AM.
type AM struct {
	// BaseURL is the AM Automation API URL, such as http://localhost:8093/automation.
	BaseURL string `json:"base_url" env:"BASE_URL" validate:"required,http_url"`
	Context
}

// APIM is the connection to APIM. With CloudToken, the bearer token is a Gravitee Cloud token,
// and an empty BaseURL becomes the token's cloud gate.
type APIM struct {
	// BaseURL is the APIM Automation API URL, such as http://localhost:8083/automation. Optional with CloudToken.
	BaseURL string `json:"base_url,omitempty" env:"BASE_URL" validate:"required_unless=CloudToken true,omitempty,http_url"`
	// CloudToken says the bearer token is a Gravitee Cloud token.
	CloudToken bool `json:"cloud_token,omitempty" env:"CLOUD_TOKEN"`
	Context
}

// Context is what AM and APIM connections share.
type Context struct {
	// OrgID is the organization id. Empty leaves the SDK's default.
	OrgID string `json:"org_id,omitempty" env:"ORG_ID"`
	// EnvID is the environment id. Empty leaves the SDK's default.
	EnvID string `json:"env_id,omitempty" env:"ENV_ID"`
	// Auth is how giops authenticates.
	Auth Auth `json:"auth" envPrefix:"AUTH_"`
}

// Auth is either a bearer token or basic auth.
type Auth struct {
	// BearerToken is sent as "Authorization: Bearer <token>".
	BearerToken *string `json:"bearer_token,omitempty" env:"BEARER_TOKEN" validate:"required_without=BasicAuth,excluded_with=BasicAuth"`
	// BasicAuth is sent as HTTP Basic.
	BasicAuth *BasicAuth `json:"basic_auth,omitempty" envPrefix:"BASIC_" validate:"required_without=BearerToken,excluded_with=BearerToken"`
}

// BasicAuth is a username and password sent as HTTP Basic.
type BasicAuth struct {
	// Username is the login.
	Username string `json:"username" env:"USERNAME" validate:"required"`
	// Password is the login's password.
	Password string `json:"password" env:"PASSWORD" validate:"required"`
}

// AMContext returns the context the AM SDK client is built from.
func (c CLIConfig) AMContext() (apicontext.APIContext, error) {
	if c.AM == nil {
		return apicontext.APIContext{}, errors.New("no 'am' section in the configuration")
	}
	return c.AM.apiContext(c.AM.BaseURL), nil
}

// APIMContext returns the context the APIM SDK client is built from.
func (c CLIConfig) APIMContext() (apimsdk.APIContext, error) {
	if c.Gamma == nil {
		return apimsdk.APIContext{}, errors.New("no 'gamma' section in the configuration")
	}
	return apimsdk.APIContext{APIContext: c.Gamma.apiContext(c.Gamma.BaseURL), CloudToken: c.Gamma.CloudToken}, nil
}

// addMissingSections lets env vars fill a section the file does not have.
func (c *CLIConfig) addMissingSections() {
	if c.AM == nil {
		c.AM = &AM{}
	}
	if c.Gamma == nil {
		c.Gamma = &APIM{}
	}
}

// dropEmptySections removes the sections neither the file nor env vars set.
func (c *CLIConfig) dropEmptySections() {
	if *c.AM == (AM{}) {
		c.AM = nil
	}
	if *c.Gamma == (APIM{}) {
		c.Gamma = nil
	}
}

// apiContext returns the SDK context of a connection to baseURL.
func (c Context) apiContext(baseURL string) apicontext.APIContext {
	ac := apicontext.APIContext{
		BaseURL: baseURL,
		OrgID:   c.OrgID,
		EnvID:   c.EnvID,
		Auth:    apicontext.Auth{BearerToken: c.Auth.BearerToken},
	}
	if basic := c.Auth.BasicAuth; basic != nil {
		ac.Auth.BasicAuth = &apicontext.BasicAuth{Username: basic.Username, Password: basic.Password}
	}
	return ac
}
