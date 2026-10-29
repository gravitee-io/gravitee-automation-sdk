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

package examples

import (
	"net/http"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const basePath = "/automation"

// Every AM Automation route in this repo, with the real AM permission for each method.
var sampleRoutes = []struct {
	path   string
	method string
	perm   string
}{
	{"/organizations/{orgId}/environments/{envId}/domains", http.MethodGet, "DOMAIN_LIST"},
	{"/organizations/{orgId}/environments/{envId}/domains", http.MethodPut, "DOMAIN_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}", http.MethodGet, "DOMAIN_READ"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}", http.MethodDelete, "DOMAIN_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/dataplanes", http.MethodGet, "DATA_PLANE_LIST"},
	{"/organizations/{orgId}/environments/{envId}/dataplanes", http.MethodPut, "DATA_PLANE_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/dataplanes/{dataPlaneId}", http.MethodGet, "DATA_PLANE_READ"},
	{"/organizations/{orgId}/environments/{envId}/dataplanes/{dataPlaneId}", http.MethodDelete, "DATA_PLANE_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/applications", http.MethodGet, "APPLICATION_LIST"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/applications", http.MethodPut, "APPLICATION_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/applications/{applicationKey}", http.MethodGet, "APPLICATION_READ"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/applications/{applicationKey}", http.MethodDelete, "APPLICATION_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/certificates", http.MethodGet, "DOMAIN_CERTIFICATE_LIST"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/certificates", http.MethodPut, "DOMAIN_CERTIFICATE_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/certificates/{certKey}", http.MethodGet, "DOMAIN_CERTIFICATE_READ"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/certificates/{certKey}", http.MethodDelete, "DOMAIN_CERTIFICATE_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/identities", http.MethodGet, "DOMAIN_IDENTITY_PROVIDER_LIST"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/identities", http.MethodPut, "DOMAIN_IDENTITY_PROVIDER_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/identities/{identityKey}", http.MethodGet, "DOMAIN_IDENTITY_PROVIDER_READ"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/identities/{identityKey}", http.MethodDelete, "DOMAIN_IDENTITY_PROVIDER_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/reporters", http.MethodGet, "DOMAIN_REPORTER_LIST"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/reporters", http.MethodPut, "DOMAIN_REPORTER_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/reporters/{reporterKey}", http.MethodGet, "DOMAIN_REPORTER_READ"},
	{"/organizations/{orgId}/environments/{envId}/domains/{domainKey}/reporters/{reporterKey}", http.MethodDelete, "DOMAIN_REPORTER_DELETE"},
}

func TestSampleAuthYAML(t *testing.T) {
	cfg, err := auth.LoadConfig("auth.yaml")
	require.NoError(t, err)

	assert.Len(t, cfg.Users, 4)
	for _, name := range []string{"admin", "readonly", "updater", "admin-per-resource"} {
		_, ok := cfg.Users[name]
		assert.True(t, ok, "missing user %s", name)
	}

	reg := auth.NewRegistry(*cfg, basePath)
	for _, route := range sampleRoutes {
		perm, ok := reg.RequiredPermission(route.method, basePath+route.path)
		assert.True(t, ok, "%s %s", route.method, route.path)
		assert.Equal(t, route.perm, perm, "%s %s", route.method, route.path)
	}

	admin, ok := reg.AuthenticateBearer("admin-token")
	require.True(t, ok)
	assert.True(t, admin.AllPermissions)

	readonly, ok := reg.AuthenticateBearer("readonly-token")
	require.True(t, ok)
	assert.True(t, readonly.HasPermission("DOMAIN_LIST"))
	assert.True(t, readonly.HasPermission("DOMAIN_READ"))
	assert.True(t, readonly.HasPermission("DATA_PLANE_LIST"))
	assert.True(t, readonly.HasPermission("DATA_PLANE_READ"))
	assert.False(t, readonly.HasPermission("DOMAIN_UPDATE"))
	assert.False(t, readonly.HasPermission("DOMAIN_DELETE"))
	assert.False(t, readonly.HasPermission("DATA_PLANE_UPDATE"))
	assert.False(t, readonly.HasPermission("DATA_PLANE_DELETE"))

	updater, ok := reg.AuthenticateBearer("updater-token")
	require.True(t, ok)
	assert.True(t, updater.HasPermission("DOMAIN_UPDATE"))
	assert.True(t, updater.HasPermission("DATA_PLANE_UPDATE"))
	assert.False(t, updater.HasPermission("DOMAIN_READ"))
	assert.False(t, updater.HasPermission("DOMAIN_DELETE"))
	assert.False(t, updater.HasPermission("DATA_PLANE_READ"))

	resourceAdmin, ok := reg.AuthenticateBearer("admin-per-resource-token")
	require.True(t, ok)
	assert.False(t, resourceAdmin.AllPermissions)
	assert.True(t, resourceAdmin.HasPermission("DOMAIN_DELETE"))
	assert.True(t, resourceAdmin.HasPermission("DATA_PLANE_LIST"))
	assert.True(t, resourceAdmin.HasPermission("DATA_PLANE_READ"))
	assert.True(t, resourceAdmin.HasPermission("DATA_PLANE_UPDATE"))
	assert.True(t, resourceAdmin.HasPermission("DATA_PLANE_DELETE"))
	assert.True(t, resourceAdmin.HasPermission("DOMAIN_CERTIFICATE_UPDATE"))
	assert.False(t, resourceAdmin.HasPermission("UNKNOWN_PERM"))
}
