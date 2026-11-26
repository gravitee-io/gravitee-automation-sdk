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

package test

import (
	"path/filepath"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func example(t *testing.T, name string) string {
	path, err := filepath.Abs(filepath.Join("..", "examples", name))
	require.NoError(t, err)
	return path
}

func TestExampleConfig(t *testing.T) {
	c, err := config.Load(example(t, "config/config.yaml"))
	require.NoError(t, err)

	am, err := c.AMContext()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8093/automation", am.BaseURL)
	assert.NotNil(t, am.Auth.BearerToken)

	apim, err := c.APIMContext()
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8083/automation", apim.BaseURL)
	assert.NotNil(t, apim.Auth.BasicAuth)
}

func TestExampleAMResources(t *testing.T) {
	requireMock(t, "the examples are not giops-* resources the live tests clean up")
	giops := newAMCLI(t)

	giops.mustRun("am", "apply", "domains", "-f", example(t, "am/domain.yaml"), "--yes")
	giops.mustRun("am", "apply", "identityproviders", "-f", example(t, "am/identity-provider.yaml"), "--yes")
	for _, apply := range [][]string{
		{"reporters", "am/reporter.yaml"},
		{"certificates", "am/certificate.yaml"},
		{"applications", "am/application.yaml"},
		{"secrets", "am/secret.yaml"},
		{"dataplanes", "am/dataplane.yaml"},
	} {
		giops.mustRun("am", "apply", apply[0], "-f", example(t, apply[1]), "--yes")
	}

	domain := giops.getJSON("am", "domains", "example-domain")
	assert.Equal(t, "Example domain", domain["name"])
	out := giops.mustRun("am", "get", "identityproviders", "--domain", "example-domain")
	assert.Regexp(t, `example-users\s+inline-am-idp`, out)
	inDomain := []string{"--domain", "example-domain"}
	assert.Equal(t, "Example audit file", giops.getJSON("am", "reporters", "example-reporter", inDomain...)["name"])
	assert.Equal(t, "Example signing certificate", giops.getJSON("am", "certificates", "example-certificate", inDomain...)["name"])
	assert.Equal(t, "Example service", giops.getJSON("am", "applications", "example-app", inDomain...)["name"])
	assert.Equal(t, "example-secret", giops.getJSON("am", "secrets", "example-secret", "--domain", "example-domain", "--application", "example-app")["key"])
	assert.Equal(t, "Example data plane", giops.getJSON("am", "dataplanes", "example-dataplane")["name"])
}

func TestExampleAPIMResources(t *testing.T) {
	requireMock(t, "the examples are not giops-* resources the live tests clean up")
	giops := newAPIMCLI(t)

	giops.mustRun("apim", "apply", "groups", "-f", example(t, "apim/group.yaml"), "--yes")

	group := giops.getJSON("apim", "groups", "example-developers")
	assert.Equal(t, "Example developers", group["name"])
}
