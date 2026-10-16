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
	"context"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/response"
	"github.com/stretchr/testify/assert"
)

var topLevelKinds = []struct {
	name string
	path string
	body func(hrid string) jsonObject
}{
	{"API", "/apis", apiBody},
	{"application", "/applications", applicationBody},
	{"group", "/groups", groupBody},
	{"dictionary", "/dictionaries", dictionaryBody},
	{"shared policy group", "/shared-policy-groups", sharedPolicyGroupBody},
	{"portal", "/portals", portalBody},
	{"portal theme", "/portal-themes", portalThemeBody},
	{"AIM catalog MCP server", "/aim/catalog/mcp-servers", aimCatalogMcpServerBody},
	{"AIM MCP proxy", "/aim/mcp-proxies", aimMcpProxyBody},
}

func TestPutGet(t *testing.T) {
	for _, kind := range topLevelKinds {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServer(t)
			url := collectionURL(srv, kind.path)

			put := putJSON(t, url, kind.body("test"))

			assert.Equal(t, "test", put["hrid"])
			assert.NotEmpty(t, put["id"])
			assert.Equal(t, "DEFAULT", put["organizationId"])
			assert.Equal(t, "DEFAULT", put["environmentId"])
			assert.Equal(t, put, getJSON(t, url, "test"))
		})
	}
}

func TestGet404(t *testing.T) {
	for _, kind := range topLevelKinds {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServer(t)

			assertGet404(t, collectionURL(srv, kind.path), "test")
		})
	}
}

func TestDelete(t *testing.T) {
	for _, kind := range topLevelKinds {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServer(t)
			url := collectionURL(srv, kind.path)
			putJSON(t, url, kind.body("test"))

			assertDeleteGone(t, url, "test")
		})
	}
}

func TestDelete404(t *testing.T) {
	for _, kind := range topLevelKinds {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServer(t)

			assertHridNotFound(t, httpDelete(t, collectionURL(srv, kind.path)+"/test"), "test")
		})
	}
}

type sdkCall func(ctx context.Context, client *sdk.APIMClient) (response.StatusCoder, error)

type sdkHridCall func(ctx context.Context, client *sdk.APIMClient, hrid string) (response.StatusCoder, error)

func sdkKinds(t *testing.T) []struct {
	name   string
	upsert sdkCall
	get    sdkHridCall
	delete sdkHridCall
} {
	return []struct {
		name   string
		upsert sdkCall
		get    sdkHridCall
		delete sdkHridCall
	}{
		{
			"API",
			func(ctx context.Context, c *sdk.APIMClient) (response.StatusCoder, error) {
				return c.UpsertApiWithResponse(ctx, nil, as[sdk.ApiV4](t, apiBody("test")))
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.GetApiWithResponse(ctx, hrid)
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.DeleteApiWithResponse(ctx, hrid)
			},
		},
		{
			"application",
			func(ctx context.Context, c *sdk.APIMClient) (response.StatusCoder, error) {
				return c.UpsertApplicationWithResponse(ctx, nil, as[sdk.Application](t, applicationBody("test")))
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.GetApplicationWithResponse(ctx, hrid)
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.DeleteApplicationWithResponse(ctx, hrid)
			},
		},
		{
			"group",
			func(ctx context.Context, c *sdk.APIMClient) (response.StatusCoder, error) {
				return c.UpsertGroupWithResponse(ctx, nil, as[sdk.Group](t, groupBody("test")))
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.GetGroupWithResponse(ctx, hrid, nil)
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.DeleteGroupWithResponse(ctx, hrid, nil)
			},
		},
		{
			"dictionary",
			func(ctx context.Context, c *sdk.APIMClient) (response.StatusCoder, error) {
				return c.UpsertDictionaryWithResponse(ctx, nil, as[sdk.Dictionary](t, dictionaryBody("test")))
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.GetDictionaryWithResponse(ctx, hrid)
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.DeleteDictionaryWithResponse(ctx, hrid)
			},
		},
		{
			"shared policy group",
			func(ctx context.Context, c *sdk.APIMClient) (response.StatusCoder, error) {
				return c.UpsertSharedPolicyGroupWithResponse(ctx, nil, as[sdk.SharedPolicyGroup](t, sharedPolicyGroupBody("test")))
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.GetSharedPolicyGroupWithResponse(ctx, hrid)
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.DeleteSharedPolicyGroupWithResponse(ctx, hrid, nil)
			},
		},
		{
			"portal",
			func(ctx context.Context, c *sdk.APIMClient) (response.StatusCoder, error) {
				return c.UpsertPortalWithResponse(ctx, nil, as[sdk.Portal](t, portalBody("test")))
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.GetPortalWithResponse(ctx, hrid)
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.DeletePortalWithResponse(ctx, hrid)
			},
		},
		{
			"portal theme",
			func(ctx context.Context, c *sdk.APIMClient) (response.StatusCoder, error) {
				return c.UpsertPortalThemeWithResponse(ctx, nil, as[sdk.PortalTheme](t, portalThemeBody("test")))
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.GetPortalThemeWithResponse(ctx, hrid)
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.DeletePortalThemeWithResponse(ctx, hrid)
			},
		},
		{
			"AIM catalog MCP server",
			func(ctx context.Context, c *sdk.APIMClient) (response.StatusCoder, error) {
				return c.UpsertAimCatalogMcpServerWithResponse(ctx, nil, as[sdk.AimCatalogMcpServer](t, aimCatalogMcpServerBody("test")))
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.GetAimCatalogMcpServerWithResponse(ctx, hrid)
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.DeleteAimCatalogMcpServerWithResponse(ctx, hrid)
			},
		},
		{
			"AIM MCP proxy",
			func(ctx context.Context, c *sdk.APIMClient) (response.StatusCoder, error) {
				return c.UpsertAimMcpProxyWithResponse(ctx, nil, as[sdk.AimMcpProxy](t, aimMcpProxyBody("test")))
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.GetAimMcpProxyWithResponse(ctx, hrid)
			},
			func(ctx context.Context, c *sdk.APIMClient, hrid string) (response.StatusCoder, error) {
				return c.DeleteAimMcpProxyWithResponse(ctx, hrid)
			},
		},
	}
}

func TestPutGetSDK(t *testing.T) {
	for _, kind := range sdkKinds(t) {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServer(t)
			client := newAPIMClient(t, srv)

			put, err := kind.upsert(t.Context(), client)
			state := assertSDKOK(t, put, err)

			get, err := kind.get(t.Context(), client, "test")
			assert.Equal(t, state, assertSDKOK(t, get, err))
		})
	}
}

func TestGet404SDK(t *testing.T) {
	for _, kind := range sdkKinds(t) {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServer(t)

			get, err := kind.get(t.Context(), newAPIMClient(t, srv), "test")
			assertSDK404(t, get, err)
		})
	}
}

func TestDeleteSDK(t *testing.T) {
	for _, kind := range sdkKinds(t) {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServer(t)
			client := newAPIMClient(t, srv)
			put, err := kind.upsert(t.Context(), client)
			assertSDKOK(t, put, err)

			del, err := kind.delete(t.Context(), client, "test")
			assertSDKNoContent(t, del, err)

			get, err := kind.get(t.Context(), client, "test")
			assertSDK404(t, get, err)
		})
	}
}

func TestDelete404SDK(t *testing.T) {
	for _, kind := range sdkKinds(t) {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServer(t)

			del, err := kind.delete(t.Context(), newAPIMClient(t, srv), "test")
			assertSDK404(t, del, err)
		})
	}
}
