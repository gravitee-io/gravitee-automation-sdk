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

	"github.com/stretchr/testify/assert"
)

func assertPut400(t *testing.T, url string, body any, messagePart string) {
	t.Helper()
	resp := httpPut(t, url, body)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, decodeTo[jsonObject](t, resp)["message"], messagePart)
}

func TestUpsertApiRejectsMissingRequiredProperty(t *testing.T) {
	_, srv := createAPIMServer(t)
	body := apiBody("test")
	delete(body, "version")

	assertPut400(t, collectionURL(srv, "/apis"), body, `property "version" is missing`)
}

func TestUpsertApiRejectsWrongType(t *testing.T) {
	_, srv := createAPIMServer(t)
	body := apiBody("test")
	body["notifyMembers"] = "yes"

	assertPut400(t, collectionURL(srv, "/apis"), body, `"/notifyMembers"`)
}

func TestUpsertApiRejectsUnknownEnumValue(t *testing.T) {
	_, srv := createAPIMServer(t)
	body := apiBody("test")
	body["type"] = "SOAP"

	assertPut400(t, collectionURL(srv, "/apis"), body, `"/type"`)
}

func TestUpsertApiRejectedBodyIsNotStored(t *testing.T) {
	_, srv := createAPIMServer(t)
	body := apiBody("test")
	delete(body, "version")

	assertPut400(t, collectionURL(srv, "/apis"), body, "version")
	assertGet404(t, collectionURL(srv, "/apis"), "test")
}

func TestUpsertAimMcpProxyRejectsMissingRequiredProperty(t *testing.T) {
	_, srv := createAPIMServer(t)
	body := aimMcpProxyBody("test")
	delete(body, "contextPath")

	assertPut400(t, collectionURL(srv, "/aim/mcp-proxies"), body, `property "contextPath" is missing`)
}
