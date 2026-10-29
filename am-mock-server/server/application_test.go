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

	"github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/stretchr/testify/assert"
)

func testApplication() Application {
	return Application{Key: "test", Name: "Test application", Type: ApplicationTypeWeb, Certificate: "cert", ClientId: new("test-client")}.WithDefaults()
}

func testApplicationSDK() sdk.Application {
	return sdk.Application{Key: "test", Name: "Test application", Type: sdk.ApplicationTypeWeb, Certificate: "cert", ClientId: new("test-client")}.WithDefaults()
}

func TestListApplications(t *testing.T) {
	am, srv := createAMServerWithDomain(t)
	defaultTenant(am).Applications.Put(newChild(testApplication(), defaultDomainKey))

	assertListEqual(t, applicationsURL(srv), []Application{testApplication()})
}

func TestGetApplication(t *testing.T) {
	am, srv := createAMServerWithDomain(t)
	defaultTenant(am).Applications.Put(newChild(testApplication(), defaultDomainKey))

	assertGetEqual(t, applicationsURL(srv), "test", testApplication())
}

func TestGetApplication404(t *testing.T) {
	_, srv := createAMServerWithDomain(t)

	assertGet404(t, applicationsURL(srv), "test", "Application")
}

func TestPutGetApplication(t *testing.T) {
	_, srv := createAMServerWithDomain(t)
	body := testApplication()

	assertPutEqual(t, applicationsURL(srv), body)
	assertGetEqual(t, applicationsURL(srv), "test", body)
}

func TestDeleteApplication(t *testing.T) {
	am, srv := createAMServerWithDomain(t)
	defaultTenant(am).Applications.Put(newChild(testApplication(), defaultDomainKey))

	assertDeleteGone(t, applicationsURL(srv), "test", "Application")
}

func TestListApplicationsSDK(t *testing.T) {
	am, srv := createAMServerWithDomain(t)
	defaultTenant(am).Applications.Put(newChild(testApplication(), defaultDomainKey))
	client := newAMClient(t, srv)

	res, err := client.ListApplicationsWithResponse(t.Context(), defaultDomainKey)
	assertSDKOK(t, res, err, []sdk.Application{testApplicationSDK()})
}

func TestGetApplicationSDK(t *testing.T) {
	am, srv := createAMServerWithDomain(t)
	defaultTenant(am).Applications.Put(newChild(testApplication(), defaultDomainKey))
	client := newAMClient(t, srv)

	res, err := client.GetApplicationWithResponse(t.Context(), defaultDomainKey, "test")
	assertSDKOK(t, res, err, testApplicationSDK())
}

func TestGetApplication404SDK(t *testing.T) {
	_, srv := createAMServerWithDomain(t)
	client := newAMClient(t, srv)

	res, err := client.GetApplicationWithResponse(t.Context(), defaultDomainKey, "test")
	assertSDK404(t, res, err)
}

func TestPutGetApplicationSDK(t *testing.T) {
	_, srv := createAMServerWithDomain(t)
	client := newAMClient(t, srv)
	body := testApplicationSDK()

	put, err := client.UpsertApplicationWithResponse(t.Context(), defaultDomainKey, nil, body)
	assertSDKOK(t, put, err, body)

	get, err := client.GetApplicationWithResponse(t.Context(), defaultDomainKey, "test")
	assertSDKOK(t, get, err, body)
}

func TestDeleteApplicationSDK(t *testing.T) {
	am, srv := createAMServerWithDomain(t)
	defaultTenant(am).Applications.Put(newChild(testApplication(), defaultDomainKey))
	client := newAMClient(t, srv)

	del, err := client.DeleteApplicationWithResponse(t.Context(), defaultDomainKey, "test")
	assertSDKNoContent(t, del, err)

	get, err := client.GetApplicationWithResponse(t.Context(), defaultDomainKey, "test")
	assertSDK404(t, get, err)
}

func TestUpsertApplicationRejectsClientIdChange(t *testing.T) {
	_, srv := createAMServerWithDomain(t)
	putJSON(t, applicationsURL(srv), jsonObject{"key": "test", "name": "Test", "type": "web", "certificate": "cert", "clientId": "first"})

	assertPut400(t, applicationsURL(srv), jsonObject{"key": "test", "name": "Test", "type": "web", "certificate": "cert", "clientId": "second"}, "clientId cannot be changed")
}

func TestUpsertApplicationRejectsClientIdOfAnotherApplication(t *testing.T) {
	_, srv := createAMServerWithDomain(t)
	putJSON(t, applicationsURL(srv), jsonObject{"key": "first", "name": "First", "type": "web", "certificate": "cert", "clientId": "shared"})

	assertPut400(t, applicationsURL(srv), jsonObject{"key": "second", "name": "Second", "type": "web", "certificate": "cert", "clientId": "shared"}, "already used by application [first]")
}

func TestUpsertApplicationRejectsTypeChangeToAgent(t *testing.T) {
	_, srv := createAMServerWithDomain(t)
	putJSON(t, applicationsURL(srv), jsonObject{"key": "test", "name": "Test", "type": "web", "certificate": "cert"})

	assertPut400(t, applicationsURL(srv), jsonObject{"key": "test", "name": "Test", "type": "agent", "kind": "AUTONOMOUS", "certificate": "cert"}, "to or from agent")
}

func TestUpsertApplicationAcceptsTypeChangeBetweenNonAgentTypes(t *testing.T) {
	_, srv := createAMServerWithDomain(t)
	putJSON(t, applicationsURL(srv), jsonObject{"key": "test", "name": "Test", "type": "web", "certificate": "cert"})

	got := putJSON(t, applicationsURL(srv), jsonObject{"key": "test", "name": "Test", "type": "native", "certificate": "cert"})

	assert.Equal(t, "native", got["type"])
}
