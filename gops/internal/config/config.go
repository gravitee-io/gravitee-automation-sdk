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

// Package config reads gops.yaml: where the Automation API is and how to authenticate.
package config

import (
	"fmt"
	"os"

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/apicontext"
	"gopkg.in/yaml.v3"
)

// TokenEnv overrides auth.token so agents and CI never write a secret to disk.
const TokenEnv = "GOPS_TOKEN"

// Config is the content of gops.yaml.
type Config struct {
	URL          string `yaml:"url"`
	Organization string `yaml:"organization"`
	Environment  string `yaml:"environment"`
	Auth         Auth   `yaml:"auth"`
}

// Auth is one of a bearer token or a basic username/password.
type Auth struct {
	Token    string `yaml:"token"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Load reads the file and applies the GOPS_TOKEN override.
func Load(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var c Config
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	if token := os.Getenv(TokenEnv); token != "" {
		c.Auth = Auth{Token: token}
	}
	if c.URL == "" {
		return Config{}, fmt.Errorf("config %s: url is required", path)
	}
	return c, nil
}

// APIContext maps the config onto the SDK's connection settings.
func (c Config) APIContext() apicontext.APIContext {
	ac := apicontext.APIContext{BaseURL: c.URL, OrgID: c.Organization, EnvID: c.Environment}
	if c.Auth.Token != "" {
		token := c.Auth.Token
		ac.Auth.BearerToken = &token
	} else if c.Auth.Username != "" {
		ac.Auth.BasicAuth = &apicontext.BasicAuth{Username: c.Auth.Username, Password: c.Auth.Password}
	}
	return ac
}
