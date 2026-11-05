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
	assert.Contains(t, *decodeTo[Error](t, resp).Message, messagePart)
}

func TestUpsertDomainRejectsMissingRequiredProperty(t *testing.T) {
	_, srv := createAMServer(t)

	assertPut400(t, domainsURL(srv), jsonObject{"key": "test", "name": "Test"}, `property "path" is missing`)
}

func TestUpsertDomainRejectsWrongType(t *testing.T) {
	_, srv := createAMServer(t)

	assertPut400(t, domainsURL(srv), jsonObject{"key": "test", "name": "Test", "path": "/test", "enabled": "yes"}, `"/enabled"`)
}

func TestUpsertDomainRejectsUnknownEnumValue(t *testing.T) {
	_, srv := createAMServer(t)
	body := jsonObject{
		"key": "test", "name": "Test", "path": "/test",
		"webAuthnSettings": jsonObject{"userVerification": "sometimes"},
	}

	assertPut400(t, domainsURL(srv), body, `"/webAuthnSettings/userVerification"`)
}

func TestUpsertDomainRejectedBodyIsNotStored(t *testing.T) {
	am, srv := createAMServer(t)

	assertPut400(t, domainsURL(srv), jsonObject{"key": "test", "name": "Test"}, "path")

	_, exists := defaultTenant(am).Domains.Get("test")
	assert.False(t, exists)
}

func TestDryRunUpsertDomainRejectsInvalidBody(t *testing.T) {
	_, srv := createAMServerWithDryRunReject(t, true)

	assertPut400(t, domainsURL(srv)+"?dryRun=true", jsonObject{"key": "test", "name": "Test"}, `property "path" is missing`)
}
