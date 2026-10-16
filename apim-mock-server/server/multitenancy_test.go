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
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

func TestMultiTenancy_ApiIsolation(t *testing.T) {
	_, srv := createAPIMServer(t)
	t1 := newTestClient(t, srv, "org-alpha", "env-alpha", "test")
	t2 := newTestClient(t, srv, "org-beta", "env-beta", "test")
	api := func(name string) sdk.ApiV4 {
		body := apiBody("shared-hrid")
		body["name"] = name
		return as[sdk.ApiV4](t, body)
	}
	nameOf := func(state any) string { return state.(*sdk.ApiV4State).Name }

	t.Run("create in both tenants", func(t *testing.T) {
		put1, err := t1.UpsertApiWithResponse(t.Context(), nil, api("Tenant 1 API"))
		assert.Equal(t, "Tenant 1 API", nameOf(assertSDKOK(t, put1, err)))

		put2, err := t2.UpsertApiWithResponse(t.Context(), nil, api("Tenant 2 API"))
		assert.Equal(t, "Tenant 2 API", nameOf(assertSDKOK(t, put2, err)))
	})

	t.Run("get returns own tenant data", func(t *testing.T) {
		get1, err := t1.GetApiWithResponse(t.Context(), "shared-hrid")
		assert.Equal(t, "Tenant 1 API", nameOf(assertSDKOK(t, get1, err)))

		get2, err := t2.GetApiWithResponse(t.Context(), "shared-hrid")
		assert.Equal(t, "Tenant 2 API", nameOf(assertSDKOK(t, get2, err)))
	})

	t.Run("update in tenant 1 does not affect tenant 2", func(t *testing.T) {
		put, err := t1.UpsertApiWithResponse(t.Context(), nil, api("Tenant 1 Updated"))
		assertSDKOK(t, put, err)

		get2, err := t2.GetApiWithResponse(t.Context(), "shared-hrid")
		assert.Equal(t, "Tenant 2 API", nameOf(assertSDKOK(t, get2, err)))
	})

	t.Run("delete in tenant 1 does not affect tenant 2", func(t *testing.T) {
		del, err := t1.DeleteApiWithResponse(t.Context(), "shared-hrid")
		assertSDKNoContent(t, del, err)

		get1, err := t1.GetApiWithResponse(t.Context(), "shared-hrid")
		assertSDK404(t, get1, err)

		get2, err := t2.GetApiWithResponse(t.Context(), "shared-hrid")
		assert.Equal(t, "Tenant 2 API", nameOf(assertSDKOK(t, get2, err)))
	})
}
