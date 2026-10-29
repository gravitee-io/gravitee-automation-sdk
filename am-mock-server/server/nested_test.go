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
)

func TestNested_IsolatedByDomain(t *testing.T) {
	_, srv := createAMServer(t)
	client := newAMClient(t, srv)

	d1 := sdk.Domain{Key: "dom-a", Name: "A", Path: "/dom-a"}.WithDefaults()
	d2 := sdk.Domain{Key: "dom-b", Name: "B", Path: "/dom-b"}.WithDefaults()
	put, err := client.UpsertDomainWithResponse(t.Context(), nil, d1)
	assertSDKOK(t, put, err, d1)
	put, err = client.UpsertDomainWithResponse(t.Context(), nil, d2)
	assertSDKOK(t, put, err, d2)

	certA := sdk.Certificate{Key: "cert", Name: new("A cert")}.WithDefaults()
	certB := sdk.Certificate{Key: "cert", Name: new("B cert")}.WithDefaults()
	up, err := client.UpsertCertificateWithResponse(t.Context(), "dom-a", nil, certA)
	assertSDKOK(t, up, err, certA)
	up, err = client.UpsertCertificateWithResponse(t.Context(), "dom-b", nil, certB)
	assertSDKOK(t, up, err, certB)

	listA, err := client.ListCertificatesWithResponse(t.Context(), "dom-a")
	assertSDKOK(t, listA, err, []sdk.Certificate{certA})
	listB, err := client.ListCertificatesWithResponse(t.Context(), "dom-b")
	assertSDKOK(t, listB, err, []sdk.Certificate{certB})

	getA, err := client.GetCertificateWithResponse(t.Context(), "dom-a", "cert")
	assertSDKOK(t, getA, err, certA)
	getB, err := client.GetCertificateWithResponse(t.Context(), "dom-b", "cert")
	assertSDKOK(t, getB, err, certB)
}

func TestNested_DeleteDomainCascades(t *testing.T) {
	_, srv := createAMServer(t)
	client := newAMClient(t, srv)

	keep := sdk.Domain{Key: "keep", Name: "Keep", Path: "/keep"}.WithDefaults()
	drop := sdk.Domain{Key: "drop", Name: "Drop", Path: "/drop"}.WithDefaults()
	put, err := client.UpsertDomainWithResponse(t.Context(), nil, keep)
	assertSDKOK(t, put, err, keep)
	put, err = client.UpsertDomainWithResponse(t.Context(), nil, drop)
	assertSDKOK(t, put, err, drop)

	certKeep := sdk.Certificate{Key: "ck", Name: new("keep cert")}.WithDefaults()
	certDrop := sdk.Certificate{Key: "cd", Name: new("drop cert")}.WithDefaults()
	idpDrop := sdk.IdentityProvider{Key: "id", Name: new("drop idp")}.WithDefaults()
	repDrop := sdk.Reporter{Key: "rd", Name: new("drop reporter")}.WithDefaults()
	appDrop := sdk.Application{Key: "ad", Name: "drop app", Type: sdk.Web, Certificate: "cd", ClientId: new("ad")}.WithDefaults()

	up, err := client.UpsertCertificateWithResponse(t.Context(), "keep", nil, certKeep)
	assertSDKOK(t, up, err, certKeep)
	up, err = client.UpsertCertificateWithResponse(t.Context(), "drop", nil, certDrop)
	assertSDKOK(t, up, err, certDrop)
	idp, err := client.UpsertIdentityProviderWithResponse(t.Context(), "drop", nil, idpDrop)
	assertSDKOK(t, idp, err, idpDrop)
	rep, err := client.UpsertReporterWithResponse(t.Context(), "drop", nil, repDrop)
	assertSDKOK(t, rep, err, repDrop)
	app, err := client.UpsertApplicationWithResponse(t.Context(), "drop", nil, appDrop)
	assertSDKOK(t, app, err, appDrop)

	del, err := client.DeleteDomainWithResponse(t.Context(), "drop")
	assertSDKNoContent(t, del, err)

	gone, err := client.GetCertificateWithResponse(t.Context(), "drop", "cd")
	assertSDK404(t, gone, err)
	goneIDP, err := client.GetIdentityProviderWithResponse(t.Context(), "drop", "id")
	assertSDK404(t, goneIDP, err)
	goneRep, err := client.GetReporterWithResponse(t.Context(), "drop", "rd")
	assertSDK404(t, goneRep, err)
	goneApp, err := client.GetApplicationWithResponse(t.Context(), "drop", "ad")
	assertSDK404(t, goneApp, err)

	still, err := client.GetCertificateWithResponse(t.Context(), "keep", "ck")
	assertSDKOK(t, still, err, certKeep)
}

func TestNested_MissingDomain404_WithAuth(t *testing.T) {
	_, srv := createAMServerWithAuth(t)
	client := newTestClient(t, srv, "", "", "admin-token")

	list, err := client.ListCertificatesWithResponse(t.Context(), "missing")
	assertSDK404(t, list, err)
	get, err := client.GetCertificateWithResponse(t.Context(), "missing", "any")
	assertSDK404(t, get, err)
}

func TestNested_DeleteMissing_204(t *testing.T) {
	_, srv := createAMServer(t)
	client := newAMClient(t, srv)

	del, err := client.DeleteDomainWithResponse(t.Context(), "no-such")
	assertSDKNoContent(t, del, err)
	delCert, err := client.DeleteCertificateWithResponse(t.Context(), "no-such", "no-such")
	assertSDK404(t, delCert, err)
}
