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

	"github.com/stretchr/testify/assert"
)

func TestUpsertApiReturnsDefaults(t *testing.T) {
	_, srv := createAPIMServer(t)
	body := apiBody("test")
	body["flowExecution"] = jsonObject{"mode": "DEFAULT"}

	got := putJSON(t, collectionURL(srv, "/apis"), body)

	assert.Equal(t, []any{}, got["tags"])
	assert.Equal(t, []any{}, got["flows"])
	assert.Equal(t, true, got["notifyMembers"])
	assert.Equal(t, jsonObject{"mode": "DEFAULT", "matchRequired": false}, got["flowExecution"], "nested objects get their defaults")
	assert.NotContains(t, got, "analytics", "unset nested objects stay unset")
}

func TestGetApiReturnsStoredDefaults(t *testing.T) {
	_, srv := createAPIMServer(t)
	putJSON(t, collectionURL(srv, "/apis"), apiBody("test"))

	got := getJSON(t, collectionURL(srv, "/apis"), "test")

	assert.Equal(t, []any{}, got["tags"])
	assert.Equal(t, true, got["notifyMembers"])
}

func TestUpsertApiKeepsExplicitValues(t *testing.T) {
	_, srv := createAPIMServer(t)
	body := apiBody("test")
	body["tags"] = []any{"internal"}
	body["notifyMembers"] = false

	got := putJSON(t, collectionURL(srv, "/apis"), body)

	assert.Equal(t, []any{"internal"}, got["tags"])
	assert.Equal(t, false, got["notifyMembers"])
}

func TestUpsertIdsAreStable(t *testing.T) {
	_, srv := createAPIMServer(t)
	url := collectionURL(srv, "/apis")

	first := putJSON(t, url, apiBody("test"))
	second := putJSON(t, url, apiBody("test"))

	assert.Equal(t, first["id"], second["id"])
	assert.Equal(t, first["crossId"], second["crossId"])
	assert.NotEqual(t, first["id"], first["crossId"])
}

func TestUpsertIdsDifferPerEnv(t *testing.T) {
	_, srv := createAPIMServer(t)

	dev := putJSON(t, orgEnvURL(srv, "DEFAULT", "dev", "/apis"), apiBody("test"))
	prod := putJSON(t, orgEnvURL(srv, "DEFAULT", "prod", "/apis"), apiBody("test"))

	assert.NotEqual(t, dev["id"], prod["id"])
}

func TestUpsertKeepsGivenCrossId(t *testing.T) {
	_, srv := createAPIMServer(t)
	body := sharedPolicyGroupBody("test")
	body["crossId"] = "legacy-cross-id"

	got := putJSON(t, collectionURL(srv, "/shared-policy-groups"), body)

	assert.Equal(t, "legacy-cross-id", got["crossId"])
}

func TestUpsertSetsOrgAndEnvFromURL(t *testing.T) {
	_, srv := createAPIMServer(t)
	body := groupBody("test")
	body["organizationId"] = "other-org"

	got := putJSON(t, orgEnvURL(srv, "org-alpha", "env-alpha", "/groups"), body)

	assert.Equal(t, "org-alpha", got["organizationId"], "client value is overwritten")
	assert.Equal(t, "env-alpha", got["environmentId"])
}
