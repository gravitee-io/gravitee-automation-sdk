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
	"time"

	"github.com/stretchr/testify/assert"
)

// Defaults are checked on the raw JSON so the test proves what goes over the wire, not what a Go type decodes to.
type jsonObject = map[string]any

func putJSON(t *testing.T, url string, body jsonObject) jsonObject {
	t.Helper()
	resp := httpPut(t, url, body)
	defer resp.Body.Close()
	failOnNotOK(t, resp)
	return decodeTo[jsonObject](t, resp)
}

func getJSON(t *testing.T, url, key string) jsonObject {
	t.Helper()
	resp := httpGet(t, url, key)
	defer resp.Body.Close()
	failOnNotOK(t, resp)
	return decodeTo[jsonObject](t, resp)
}

func TestUpsertDomainReturnsDefaults(t *testing.T) {
	_, srv := createAMServer(t)

	got := putJSON(t, domainsURL(srv), jsonObject{"key": "test", "name": "Test", "path": "/test", "oidc": jsonObject{}})

	assert.Equal(t, true, got["enabled"])
	assert.Equal(t, false, got["master"])
	assert.Equal(t, jsonObject{"redirectUriStrictMatching": false}, got["oidc"], "nested settings get their defaults")
	assert.NotContains(t, got, "accountSettings", "unset nested settings stay unset")
}

func TestGetDomainReturnsStoredDefaults(t *testing.T) {
	_, srv := createAMServer(t)
	putJSON(t, domainsURL(srv), jsonObject{"key": "test", "name": "Test", "path": "/test"})

	got := getJSON(t, domainsURL(srv), "test")

	assert.Equal(t, true, got["enabled"])
	assert.Equal(t, false, got["master"])
}

func TestUpsertDomainKeepsExplicitValues(t *testing.T) {
	_, srv := createAMServer(t)

	got := putJSON(t, domainsURL(srv), jsonObject{"key": "test", "name": "Test", "path": "/test", "enabled": false})

	assert.Equal(t, false, got["enabled"])
}

func TestUpsertReporterReturnsDefaults(t *testing.T) {
	_, srv := createAMServerWithDomain(t)

	got := putJSON(t, reportersURL(srv), jsonObject{"key": "test", "name": "Test", "type": "reporter-am-file"})

	assert.Equal(t, true, got["enabled"])
	assert.Equal(t, false, got["system"])
}

func TestUpsertApplicationGeneratesClientId(t *testing.T) {
	_, srv := createAMServerWithDomain(t)
	app := jsonObject{"key": "test", "name": "Test", "type": "web", "certificate": "cert"}

	created := putJSON(t, applicationsURL(srv), app)
	updated := putJSON(t, applicationsURL(srv), app)

	assert.NotEmpty(t, created["clientId"])
	assert.Equal(t, created["clientId"], updated["clientId"], "an omitted clientId keeps the current value")
}

func TestUpsertApplicationKeepsExplicitClientId(t *testing.T) {
	_, srv := createAMServerWithDomain(t)

	got := putJSON(t, applicationsURL(srv), jsonObject{"key": "test", "name": "Test", "type": "web", "certificate": "cert", "clientId": "my-client"})

	assert.Equal(t, "my-client", got["clientId"])
}

func TestUpsertDomainSetsTimestamps(t *testing.T) {
	_, srv := createAMServer(t)

	got := putJSON(t, domainsURL(srv), jsonObject{"key": "test", "name": "Test", "path": "/test"})

	assert.NotEmpty(t, got["createdAt"])
	assert.Equal(t, got["createdAt"], got["updatedAt"], "a new resource is created and updated at the same instant")
}

func TestUpsertDomainKeepsCreatedAtAndIgnoresClientTimestamps(t *testing.T) {
	_, srv := createAMServer(t)
	created := putJSON(t, domainsURL(srv), jsonObject{"key": "test", "name": "Test", "path": "/test"})
	time.Sleep(2 * time.Millisecond)

	updated := putJSON(t, domainsURL(srv), jsonObject{
		"key": "test", "name": "Renamed", "path": "/test",
		"createdAt": "2000-01-01T00:00:00Z", "updatedAt": "2000-01-01T00:00:00Z",
	})

	assert.Equal(t, created["createdAt"], updated["createdAt"])
	assert.NotEqual(t, created["updatedAt"], updated["updatedAt"])
	assert.NotEqual(t, "2000-01-01T00:00:00Z", updated["updatedAt"])
}

func TestUpsertCertificateKeepsCreatedAt(t *testing.T) {
	_, srv := createAMServerWithDomain(t)
	created := putJSON(t, certificatesURL(srv), jsonObject{"key": "test", "name": "Test"})
	time.Sleep(2 * time.Millisecond)

	updated := putJSON(t, certificatesURL(srv), jsonObject{"key": "test", "name": "Renamed"})

	assert.Equal(t, created["createdAt"], updated["createdAt"])
	assert.NotEqual(t, created["updatedAt"], updated["updatedAt"])
}

func TestUpsertDataPlaneSetsOrgAndEnvFromURL(t *testing.T) {
	_, srv := createAMServer(t)
	url := srv.URL + BasePath + "/organizations/org-alpha/environments/env-alpha/dataplanes"

	got := putJSON(t, url, jsonObject{"id": "test", "name": "Test", "organizationId": "other-org"})

	assert.Equal(t, "org-alpha", got["organizationId"], "client value is overwritten")
	assert.Equal(t, "env-alpha", got["environmentId"])
}
