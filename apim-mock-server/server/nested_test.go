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
)

var subResourceKinds = []struct {
	name       string
	parent     string // the parent's path
	parentHrid string
	parentKey  string
	path       string
	body       func(hrid string) jsonObject
}{
	{"API subscription", "/apis/" + apiHrid, apiHrid, "apiHrid", "/subscriptions", subscriptionBody},
	{"API documentation", "/apis/" + apiHrid, apiHrid, "apiHrid", "/documentations", documentationBody},
	{"API link", "/apis/" + apiHrid, apiHrid, "apiHrid", "/links", linkBody},
	{"portal listing", "/portals/" + portalHrid, portalHrid, "portalHrid", "/listings", listingBody},
	{"portal documentation", "/portals/" + portalHrid, portalHrid, "portalHrid", "/documentations", documentationBody},
	{"portal link", "/portals/" + portalHrid, portalHrid, "portalHrid", "/links", linkBody},
}

func TestNested_PutGetDelete(t *testing.T) {
	for _, kind := range subResourceKinds {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServerWithParents(t)
			url := collectionURL(srv, kind.parent+kind.path)

			put := putJSON(t, url, kind.body("test"))
			assert.NotEmpty(t, put["id"])
			assert.Equal(t, kind.parentHrid, put[kind.parentKey], "the parent is in the status")
			assert.Equal(t, put, getJSON(t, url, "test"))

			assertDeleteGone(t, url, "test")
		})
	}
}

func TestNested_MissingParent404(t *testing.T) {
	for _, kind := range subResourceKinds {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServer(t)
			url := collectionURL(srv, kind.parent+kind.path)

			assertHridNotFound(t, httpPut(t, url, kind.body("test")), kind.parentHrid)
			assertHridNotFound(t, httpGet(t, url, "test"), kind.parentHrid)
			assertHridNotFound(t, httpDelete(t, url+"/test"), kind.parentHrid)
		})
	}
}

func TestNested_Get404UnderExistingParent(t *testing.T) {
	for _, kind := range subResourceKinds {
		t.Run(kind.name, func(t *testing.T) {
			_, srv := createAPIMServerWithParents(t)
			url := collectionURL(srv, kind.parent+kind.path)

			assertGet404(t, url, "test")
			assertHridNotFound(t, httpDelete(t, url+"/test"), "test")
		})
	}
}

func TestNested_IsolatedByParent(t *testing.T) {
	_, srv := createAPIMServer(t)
	putJSON(t, collectionURL(srv, "/apis"), apiBody("api-a"))
	putJSON(t, collectionURL(srv, "/apis"), apiBody("api-b"))
	docA, docB := documentationBody("doc"), documentationBody("doc")
	docA["content"], docB["content"] = "A", "B"

	putJSON(t, collectionURL(srv, "/apis/api-a/documentations"), docA)
	putJSON(t, collectionURL(srv, "/apis/api-b/documentations"), docB)

	assert.Equal(t, "A", getJSON(t, collectionURL(srv, "/apis/api-a/documentations"), "doc")["content"])
	assert.Equal(t, "B", getJSON(t, collectionURL(srv, "/apis/api-b/documentations"), "doc")["content"])
}

func TestNested_SubscriptionStatus(t *testing.T) {
	_, srv := createAPIMServerWithParents(t)
	url := collectionURL(srv, "/apis/"+apiHrid+"/subscriptions")

	created := putJSON(t, url, subscriptionBody("sub"))
	updated := putJSON(t, url, subscriptionBody("sub"))

	assert.Equal(t, apiHrid, created["apiHrid"])
	assert.NotEmpty(t, created["startingAt"])
	assert.Equal(t, created["startingAt"], updated["startingAt"], "a subscription keeps its start")
}

func TestNested_SubscriptionUnknownApplication404(t *testing.T) {
	_, srv := createAPIMServerWithParents(t)
	body := subscriptionBody("sub")
	body["applicationHrid"] = "no-such-app"

	assertHridNotFound(t, httpPut(t, collectionURL(srv, "/apis/"+apiHrid+"/subscriptions"), body), "no-such-app")
}

func TestNested_SubscriptionUnknownPlan404(t *testing.T) {
	_, srv := createAPIMServerWithParents(t)
	body := subscriptionBody("sub")
	body["planHrid"] = "no-such-plan"

	assertHridNotFound(t, httpPut(t, collectionURL(srv, "/apis/"+apiHrid+"/subscriptions"), body), "no-such-plan")
}

func TestNested_SubscriptionUnknownPlan404SDK(t *testing.T) {
	_, srv := createAPIMServerWithParents(t)
	body := as[sdk.Subscription](t, subscriptionBody("sub"))
	body.PlanHrid = "no-such-plan"

	put, err := newAPIMClient(t, srv).UpsertApiSubscriptionWithResponse(t.Context(), apiHrid, body)
	assertSDK404(t, put, err)
}

func TestNested_DeleteApiCascades(t *testing.T) {
	_, srv := createAPIMServerWithParents(t)
	putJSON(t, collectionURL(srv, "/apis"), apiBody("keep"))
	for _, path := range []string{"/apis/" + apiHrid, "/apis/keep"} {
		putJSON(t, collectionURL(srv, path+"/subscriptions"), subscriptionBody("sub"))
		putJSON(t, collectionURL(srv, path+"/documentations"), documentationBody("doc"))
		putJSON(t, collectionURL(srv, path+"/links"), linkBody("link"))
	}

	resp := httpDelete(t, collectionURL(srv, "/apis/"+apiHrid))
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Recreate the API to look under it: the parent check would answer 404 for any child otherwise.
	putJSON(t, collectionURL(srv, "/apis"), apiBody(apiHrid))
	assertGet404(t, collectionURL(srv, "/apis/"+apiHrid+"/subscriptions"), "sub")
	assertGet404(t, collectionURL(srv, "/apis/"+apiHrid+"/documentations"), "doc")
	assertGet404(t, collectionURL(srv, "/apis/"+apiHrid+"/links"), "link")
	getJSON(t, collectionURL(srv, "/apis/keep/subscriptions"), "sub")
	getJSON(t, collectionURL(srv, "/apis/keep/documentations"), "doc")
	getJSON(t, collectionURL(srv, "/apis/keep/links"), "link")
}

func TestNested_DeletePortalCascades(t *testing.T) {
	_, srv := createAPIMServerWithParents(t)
	base := "/portals/" + portalHrid
	putJSON(t, collectionURL(srv, base+"/listings"), listingBody("listing"))
	putJSON(t, collectionURL(srv, base+"/documentations"), documentationBody("doc"))
	putJSON(t, collectionURL(srv, base+"/links"), linkBody("link"))

	resp := httpDelete(t, collectionURL(srv, base))
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	putJSON(t, collectionURL(srv, "/portals"), portalBody(portalHrid))
	assertGet404(t, collectionURL(srv, base+"/listings"), "listing")
	assertGet404(t, collectionURL(srv, base+"/documentations"), "doc")
	assertGet404(t, collectionURL(srv, base+"/links"), "link")
}

func TestNested_DeleteApplicationRemovesItsSubscriptions(t *testing.T) {
	_, srv := createAPIMServerWithParents(t)
	putJSON(t, collectionURL(srv, "/applications"), applicationBody("other-app"))
	url := collectionURL(srv, "/apis/"+apiHrid+"/subscriptions")
	putJSON(t, url, subscriptionBody("sub"))
	other := subscriptionBody("other-sub")
	other["applicationHrid"] = "other-app"
	putJSON(t, url, other)

	resp := httpDelete(t, collectionURL(srv, "/applications/"+applicationHrid))
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	assertGet404(t, url, "sub")
	getJSON(t, url, "other-sub")
}
