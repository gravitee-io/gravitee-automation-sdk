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
	"net/http"
	"testing"

	"github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDryRunPut_NoRejectReturnsStateAndDoesNotPersist(t *testing.T) {
	_, srv := createAPIMServer(t)
	url := collectionURL(srv, "/apis")

	got := putJSON(t, url+"?dryRun=true", apiBody("test"))

	assert.NotEmpty(t, got["id"], "the State is computed as for a real PUT")
	assert.NotEmpty(t, got["crossId"])
	assert.Equal(t, []any{}, got["tags"], "defaults are applied")
	assert.NotContains(t, got, "errors")
	assertGet404(t, url, "test")
}

func TestDryRunPut_RejectAddsSevereError(t *testing.T) {
	_, srv := createAPIMServerWithDryRunReject(t, true)
	url := collectionURL(srv, "/apis")

	got := putJSON(t, url+"?dryRun=true", apiBody("test"))

	assert.Equal(t, jsonObject{"severe": []any{dryRunMessage}}, got["errors"])
	assert.NotEmpty(t, got["id"])
	assertGet404(t, url, "test")
}

func TestDryRunPut_FalseStillPersists(t *testing.T) {
	_, srv := createAPIMServerWithDryRunReject(t, true)
	url := collectionURL(srv, "/apis")

	got := putJSON(t, url+"?dryRun=false", apiBody("test"))

	assert.NotContains(t, got, "errors")
	assert.Equal(t, got, getJSON(t, url, "test"))
}

func TestDryRunPut_LeavesExistingUnchanged(t *testing.T) {
	_, srv := createAPIMServerWithDryRunReject(t, true)
	url := collectionURL(srv, "/apis")
	existing := putJSON(t, url, apiBody("test"))
	changed := apiBody("test")
	changed["name"] = "Changed"

	got := putJSON(t, url+"?dryRun=true", changed)

	assert.Equal(t, "Changed", got["name"])
	assert.Equal(t, existing, getJSON(t, url, "test"))
}

func TestDryRunPut_Nested(t *testing.T) {
	_, srv := createAPIMServerWithParents(t)
	url := collectionURL(srv, "/apis/"+apiHrid+"/documentations")

	got := putJSON(t, url+"?dryRun=true", documentationBody("doc"))

	assert.Equal(t, apiHrid, got["apiHrid"])
	assertGet404(t, url, "doc")
}

func TestDryRunPut_SDKReadsSevereErrors(t *testing.T) {
	_, srv := createAPIMServerWithDryRunReject(t, true)
	client := newAPIMClient(t, srv)

	put, err := client.UpsertApiWithResponse(t.Context(), &sdk.UpsertApiParams{DryRun: new(true)}, as[sdk.ApiV4](t, apiBody("test")))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, put.StatusCode())
	require.NotNil(t, put.JSON200.Errors)
	assert.Equal(t, []string{dryRunMessage}, put.JSON200.Errors.Severe)

	get, err := client.GetApiWithResponse(t.Context(), "test")
	assertSDK404(t, get, err)
}

func TestDryRunDelete_SharedPolicyGroupIsKept(t *testing.T) {
	_, srv := createAPIMServer(t)
	url := collectionURL(srv, "/shared-policy-groups")
	putJSON(t, url, sharedPolicyGroupBody("test"))

	resp := httpDelete(t, url+"/test?dryRun=true")
	resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	getJSON(t, url, "test")
}

func TestDryRunDelete_SharedPolicyGroupMissing404(t *testing.T) {
	_, srv := createAPIMServer(t)

	assertHridNotFound(t, httpDelete(t, collectionURL(srv, "/shared-policy-groups")+"/test?dryRun=true"), "test")
}
