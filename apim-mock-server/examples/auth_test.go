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

// Every APIM Automation route in this repo, with the APIM permission for each method.
var sampleRoutes = []struct {
	path   string
	method string
	perm   string
}{
	{"/organizations/{orgId}/environments/{envId}/apis", http.MethodPut, "API_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/apis/{hrid}", http.MethodGet, "API_READ"},
	{"/organizations/{orgId}/environments/{envId}/apis/{hrid}", http.MethodDelete, "API_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/applications", http.MethodPut, "APPLICATION_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/applications/{hrid}", http.MethodGet, "APPLICATION_READ"},
	{"/organizations/{orgId}/environments/{envId}/applications/{hrid}", http.MethodDelete, "APPLICATION_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/groups", http.MethodPut, "GROUP_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/groups/{hrid}", http.MethodGet, "GROUP_READ"},
	{"/organizations/{orgId}/environments/{envId}/groups/{hrid}", http.MethodDelete, "GROUP_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/dictionaries", http.MethodPut, "DICTIONARY_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/dictionaries/{hrid}", http.MethodGet, "DICTIONARY_READ"},
	{"/organizations/{orgId}/environments/{envId}/dictionaries/{hrid}", http.MethodDelete, "DICTIONARY_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/shared-policy-groups", http.MethodPut, "SHARED_POLICY_GROUP_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/shared-policy-groups/{hrid}", http.MethodGet, "SHARED_POLICY_GROUP_READ"},
	{"/organizations/{orgId}/environments/{envId}/shared-policy-groups/{hrid}", http.MethodDelete, "SHARED_POLICY_GROUP_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/portals", http.MethodPut, "PORTAL_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/portals/{hrid}", http.MethodGet, "PORTAL_READ"},
	{"/organizations/{orgId}/environments/{envId}/portals/{hrid}", http.MethodDelete, "PORTAL_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/portal-themes", http.MethodPut, "PORTAL_THEME_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/portal-themes/{hrid}", http.MethodGet, "PORTAL_THEME_READ"},
	{"/organizations/{orgId}/environments/{envId}/portal-themes/{hrid}", http.MethodDelete, "PORTAL_THEME_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/aim/catalog/mcp-servers", http.MethodPut, "AIM_CATALOG_MCP_SERVER_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/aim/catalog/mcp-servers/{hrid}", http.MethodGet, "AIM_CATALOG_MCP_SERVER_READ"},
	{"/organizations/{orgId}/environments/{envId}/aim/catalog/mcp-servers/{hrid}", http.MethodDelete, "AIM_CATALOG_MCP_SERVER_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/aim/mcp-proxies", http.MethodPut, "AIM_MCP_PROXY_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/aim/mcp-proxies/{hrid}", http.MethodGet, "AIM_MCP_PROXY_READ"},
	{"/organizations/{orgId}/environments/{envId}/aim/mcp-proxies/{hrid}", http.MethodDelete, "AIM_MCP_PROXY_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/apis/{apiHrid}/subscriptions", http.MethodPut, "API_SUBSCRIPTION_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/apis/{apiHrid}/subscriptions/{hrid}", http.MethodGet, "API_SUBSCRIPTION_READ"},
	{"/organizations/{orgId}/environments/{envId}/apis/{apiHrid}/subscriptions/{hrid}", http.MethodDelete, "API_SUBSCRIPTION_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/apis/{apiHrid}/documentations", http.MethodPut, "API_DOCUMENTATION_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/apis/{apiHrid}/documentations/{hrid}", http.MethodGet, "API_DOCUMENTATION_READ"},
	{"/organizations/{orgId}/environments/{envId}/apis/{apiHrid}/documentations/{hrid}", http.MethodDelete, "API_DOCUMENTATION_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/apis/{apiHrid}/links", http.MethodPut, "API_LINK_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/apis/{apiHrid}/links/{hrid}", http.MethodGet, "API_LINK_READ"},
	{"/organizations/{orgId}/environments/{envId}/apis/{apiHrid}/links/{hrid}", http.MethodDelete, "API_LINK_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/portals/{portalHrid}/listings", http.MethodPut, "PORTAL_LISTING_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/portals/{portalHrid}/listings/{hrid}", http.MethodGet, "PORTAL_LISTING_READ"},
	{"/organizations/{orgId}/environments/{envId}/portals/{portalHrid}/listings/{hrid}", http.MethodDelete, "PORTAL_LISTING_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/portals/{portalHrid}/documentations", http.MethodPut, "PORTAL_DOCUMENTATION_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/portals/{portalHrid}/documentations/{hrid}", http.MethodGet, "PORTAL_DOCUMENTATION_READ"},
	{"/organizations/{orgId}/environments/{envId}/portals/{portalHrid}/documentations/{hrid}", http.MethodDelete, "PORTAL_DOCUMENTATION_DELETE"},
	{"/organizations/{orgId}/environments/{envId}/portals/{portalHrid}/links", http.MethodPut, "PORTAL_LINK_UPDATE"},
	{"/organizations/{orgId}/environments/{envId}/portals/{portalHrid}/links/{hrid}", http.MethodGet, "PORTAL_LINK_READ"},
	{"/organizations/{orgId}/environments/{envId}/portals/{portalHrid}/links/{hrid}", http.MethodDelete, "PORTAL_LINK_DELETE"},
}

func TestSampleAuthYAML(t *testing.T) {
	cfg, err := auth.LoadConfig("auth.yaml")
	require.NoError(t, err)

	assert.Len(t, cfg.Users, 3)
	for _, name := range []string{"admin", "readonly", "updater"} {
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
	assert.True(t, readonly.HasPermission("API_READ"))
	assert.True(t, readonly.HasPermission("AIM_MCP_PROXY_READ"))
	assert.False(t, readonly.HasPermission("API_UPDATE"))
	assert.False(t, readonly.HasPermission("API_DELETE"))

	updater, ok := reg.AuthenticateBearer("updater-token")
	require.True(t, ok)
	assert.True(t, updater.HasPermission("API_UPDATE"))
	assert.True(t, updater.HasPermission("PORTAL_LINK_UPDATE"))
	assert.False(t, updater.HasPermission("API_READ"))
	assert.False(t, updater.HasPermission("API_DELETE"))
}
