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
	"net/http"
	"sync"

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/store"
)

// MockAPIM is an in-memory StrictServer. Tenants are created on first request, keyed by org+env. Data is process-local.
type MockAPIM struct {
	tenants      map[string]*tenant
	mutex        sync.Mutex
	dryRunReject bool
}

type tenant struct {
	Apis                 *store.Store[ApiV4State]
	Applications         *store.Store[ApplicationState]
	Groups               *store.Store[GroupState]
	Dictionaries         *store.Store[DictionaryState]
	SharedPolicyGroups   *store.Store[SharedPolicyGroupState]
	Portals              *store.Store[PortalState]
	PortalThemes         *store.Store[PortalThemeState]
	AimCatalogMcpServers *store.Store[AimCatalogMcpServerState]
	AimMcpProxies        *store.Store[AimMcpProxyState]
	Subscriptions        *store.Store[Child[SubscriptionState]]
	ApiDocumentations    *store.Store[Child[DocumentationState]]
	ApiLinks             *store.Store[Child[PortalLinkState]]
	PortalListings       *store.Store[Child[PortalListingState]]
	PortalDocumentations *store.Store[Child[DocumentationState]]
	PortalLinks          *store.Store[Child[PortalLinkState]]
}

func newTenant() *tenant {
	return &tenant{
		Apis:                 store.NewStore[ApiV4State](),
		Applications:         store.NewStore[ApplicationState](),
		Groups:               store.NewStore[GroupState](),
		Dictionaries:         store.NewStore[DictionaryState](),
		SharedPolicyGroups:   store.NewStore[SharedPolicyGroupState](),
		Portals:              store.NewStore[PortalState](),
		PortalThemes:         store.NewStore[PortalThemeState](),
		AimCatalogMcpServers: store.NewStore[AimCatalogMcpServerState](),
		AimMcpProxies:        store.NewStore[AimMcpProxyState](),
		Subscriptions:        store.NewStore[Child[SubscriptionState]](),
		ApiDocumentations:    store.NewStore[Child[DocumentationState]](),
		ApiLinks:             store.NewStore[Child[PortalLinkState]](),
		PortalListings:       store.NewStore[Child[PortalListingState]](),
		PortalDocumentations: store.NewStore[Child[DocumentationState]](),
		PortalLinks:          store.NewStore[Child[PortalLinkState]](),
	}
}

var _ StrictServerInterface = &MockAPIM{}

// NewMockAPIM returns an empty mock. With dryRunReject, a dry-run PUT answers its State with errors.severe set.
func NewMockAPIM(dryRunReject bool) *MockAPIM {
	return &MockAPIM{tenants: make(map[string]*tenant), dryRunReject: dryRunReject}
}

func (m *MockAPIM) tenant(org, env string) *tenant {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	key := org + "-" + env
	if _, ok := m.tenants[key]; !ok {
		m.tenants[key] = newTenant()
	}
	return m.tenants[key]
}

func (m *MockAPIM) apiExists(org, env, hrid string) bool {
	_, ok := m.tenant(org, env).Apis.Get(hrid)
	return ok
}

func (m *MockAPIM) portalExists(org, env, hrid string) bool {
	_, ok := m.tenant(org, env).Portals.Get(hrid)
	return ok
}

// APIs

func (m *MockAPIM) UpsertApi(_ context.Context, req UpsertApiRequestObject) (UpsertApiResponseObject, error) {
	scope := at(req.OrgId, req.EnvId, "api")
	status := m.status(scope, req.Body.Hrid, req.Params.DryRun)
	status["crossId"] = scope.crossID(req.Body.Hrid, req.Body.CrossId)
	api := state[ApiV4State](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).Apis, api, req.Params.DryRun)
	return UpsertApi200JSONResponse(api), nil
}

func (m *MockAPIM) GetApi(_ context.Context, req GetApiRequestObject) (GetApiResponseObject, error) {
	if api, ok := m.tenant(req.OrgId, req.EnvId).Apis.Get(req.Hrid); ok {
		return GetApi200JSONResponse(api), nil
	}
	return GetApi404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

// DeleteApi also deletes the API's subscriptions, documentations and links.
func (m *MockAPIM) DeleteApi(_ context.Context, req DeleteApiRequestObject) (DeleteApiResponseObject, error) {
	t := m.tenant(req.OrgId, req.EnvId)
	if !remove(t.Apis, req.Hrid) {
		return DeleteApi404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	removeChildren(t.Subscriptions, req.Hrid)
	removeChildren(t.ApiDocumentations, req.Hrid)
	removeChildren(t.ApiLinks, req.Hrid)
	return DeleteApi204Response{}, nil
}

// Applications

func (m *MockAPIM) UpsertApplication(_ context.Context, req UpsertApplicationRequestObject) (UpsertApplicationResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "application"), req.Body.Hrid, req.Params.DryRun)
	application := state[ApplicationState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).Applications, application, req.Params.DryRun)
	return UpsertApplication200JSONResponse(application), nil
}

func (m *MockAPIM) GetApplication(_ context.Context, req GetApplicationRequestObject) (GetApplicationResponseObject, error) {
	if application, ok := m.tenant(req.OrgId, req.EnvId).Applications.Get(req.Hrid); ok {
		return GetApplication200JSONResponse(application), nil
	}
	return GetApplication404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

// DeleteApplication also deletes the application's subscriptions, as APIM closes them.
func (m *MockAPIM) DeleteApplication(_ context.Context, req DeleteApplicationRequestObject) (DeleteApplicationResponseObject, error) {
	t := m.tenant(req.OrgId, req.EnvId)
	if !remove(t.Applications, req.Hrid) {
		return DeleteApplication404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	for _, subscription := range t.Subscriptions.GetAll() {
		if subscription.Self.ApplicationHrid == req.Hrid {
			t.Subscriptions.Delete(subscription)
		}
	}
	return DeleteApplication204Response{}, nil
}

// Groups

func (m *MockAPIM) UpsertGroup(_ context.Context, req UpsertGroupRequestObject) (UpsertGroupResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "group"), req.Body.Hrid, req.Params.DryRun)
	group := state[GroupState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).Groups, group, req.Params.DryRun)
	return UpsertGroup200JSONResponse(group), nil
}

func (m *MockAPIM) GetGroup(_ context.Context, req GetGroupRequestObject) (GetGroupResponseObject, error) {
	if group, ok := m.tenant(req.OrgId, req.EnvId).Groups.Get(req.Hrid); ok {
		return GetGroup200JSONResponse(group), nil
	}
	return GetGroup404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeleteGroup(_ context.Context, req DeleteGroupRequestObject) (DeleteGroupResponseObject, error) {
	if !remove(m.tenant(req.OrgId, req.EnvId).Groups, req.Hrid) {
		return DeleteGroup404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	return DeleteGroup204Response{}, nil
}

// Dictionaries

func (m *MockAPIM) UpsertDictionary(_ context.Context, req UpsertDictionaryRequestObject) (UpsertDictionaryResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "dictionary"), req.Body.Hrid, req.Params.DryRun)
	dictionary := state[DictionaryState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).Dictionaries, dictionary, req.Params.DryRun)
	return UpsertDictionary200JSONResponse(dictionary), nil
}

func (m *MockAPIM) GetDictionary(_ context.Context, req GetDictionaryRequestObject) (GetDictionaryResponseObject, error) {
	if dictionary, ok := m.tenant(req.OrgId, req.EnvId).Dictionaries.Get(req.Hrid); ok {
		return GetDictionary200JSONResponse(dictionary), nil
	}
	return GetDictionary404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeleteDictionary(_ context.Context, req DeleteDictionaryRequestObject) (DeleteDictionaryResponseObject, error) {
	if !remove(m.tenant(req.OrgId, req.EnvId).Dictionaries, req.Hrid) {
		return DeleteDictionary404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	return DeleteDictionary204Response{}, nil
}

// Portal themes

func (m *MockAPIM) UpsertPortalTheme(_ context.Context, req UpsertPortalThemeRequestObject) (UpsertPortalThemeResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "portal-theme"), req.Body.Hrid, req.Params.DryRun)
	theme := state[PortalThemeState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).PortalThemes, theme, req.Params.DryRun)
	return UpsertPortalTheme200JSONResponse(theme), nil
}

func (m *MockAPIM) GetPortalTheme(_ context.Context, req GetPortalThemeRequestObject) (GetPortalThemeResponseObject, error) {
	if theme, ok := m.tenant(req.OrgId, req.EnvId).PortalThemes.Get(req.Hrid); ok {
		return GetPortalTheme200JSONResponse(theme), nil
	}
	return GetPortalTheme404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeletePortalTheme(_ context.Context, req DeletePortalThemeRequestObject) (DeletePortalThemeResponseObject, error) {
	if !remove(m.tenant(req.OrgId, req.EnvId).PortalThemes, req.Hrid) {
		return DeletePortalTheme404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	return DeletePortalTheme204Response{}, nil
}

// AIM catalog MCP servers

func (m *MockAPIM) UpsertAimCatalogMcpServer(_ context.Context, req UpsertAimCatalogMcpServerRequestObject) (UpsertAimCatalogMcpServerResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "aim-catalog-mcp-server"), req.Body.Hrid, req.Params.DryRun)
	server := state[AimCatalogMcpServerState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).AimCatalogMcpServers, server, req.Params.DryRun)
	return UpsertAimCatalogMcpServer200JSONResponse(server), nil
}

func (m *MockAPIM) GetAimCatalogMcpServer(_ context.Context, req GetAimCatalogMcpServerRequestObject) (GetAimCatalogMcpServerResponseObject, error) {
	if server, ok := m.tenant(req.OrgId, req.EnvId).AimCatalogMcpServers.Get(req.Hrid); ok {
		return GetAimCatalogMcpServer200JSONResponse(server), nil
	}
	return GetAimCatalogMcpServer404JSONResponse{AimNotFoundJSONResponse(aimHridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeleteAimCatalogMcpServer(_ context.Context, req DeleteAimCatalogMcpServerRequestObject) (DeleteAimCatalogMcpServerResponseObject, error) {
	if !remove(m.tenant(req.OrgId, req.EnvId).AimCatalogMcpServers, req.Hrid) {
		return DeleteAimCatalogMcpServer404JSONResponse{AimNotFoundJSONResponse(aimHridNotFound(req.Hrid))}, nil
	}
	return DeleteAimCatalogMcpServer204Response{}, nil
}

// AIM MCP proxies

func (m *MockAPIM) UpsertAimMcpProxy(_ context.Context, req UpsertAimMcpProxyRequestObject) (UpsertAimMcpProxyResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "aim-mcp-proxy"), req.Body.Hrid, req.Params.DryRun)
	proxy := state[AimMcpProxyState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).AimMcpProxies, proxy, req.Params.DryRun)
	return UpsertAimMcpProxy200JSONResponse(proxy), nil
}

func (m *MockAPIM) GetAimMcpProxy(_ context.Context, req GetAimMcpProxyRequestObject) (GetAimMcpProxyResponseObject, error) {
	if proxy, ok := m.tenant(req.OrgId, req.EnvId).AimMcpProxies.Get(req.Hrid); ok {
		return GetAimMcpProxy200JSONResponse(proxy), nil
	}
	return GetAimMcpProxy404JSONResponse{AimNotFoundJSONResponse(aimHridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeleteAimMcpProxy(_ context.Context, req DeleteAimMcpProxyRequestObject) (DeleteAimMcpProxyResponseObject, error) {
	if !remove(m.tenant(req.OrgId, req.EnvId).AimMcpProxies, req.Hrid) {
		return DeleteAimMcpProxy404JSONResponse{AimNotFoundJSONResponse(aimHridNotFound(req.Hrid))}, nil
	}
	return DeleteAimMcpProxy204Response{}, nil
}

// Shared policy groups

func (m *MockAPIM) UpsertSharedPolicyGroup(_ context.Context, req UpsertSharedPolicyGroupRequestObject) (UpsertSharedPolicyGroupResponseObject, error) {
	scope := at(req.OrgId, req.EnvId, "shared-policy-group")
	status := m.status(scope, req.Body.Hrid, req.Params.DryRun)
	status["crossId"] = scope.crossID(req.Body.Hrid, req.Body.CrossId)
	group := state[SharedPolicyGroupState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).SharedPolicyGroups, group, req.Params.DryRun)
	return UpsertSharedPolicyGroup200JSONResponse(group), nil
}

func (m *MockAPIM) GetSharedPolicyGroup(_ context.Context, req GetSharedPolicyGroupRequestObject) (GetSharedPolicyGroupResponseObject, error) {
	if group, ok := m.tenant(req.OrgId, req.EnvId).SharedPolicyGroups.Get(req.Hrid); ok {
		return GetSharedPolicyGroup200JSONResponse(group), nil
	}
	return GetSharedPolicyGroup404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

// DeleteSharedPolicyGroup supports dry run: it then only checks the group exists.
func (m *MockAPIM) DeleteSharedPolicyGroup(_ context.Context, req DeleteSharedPolicyGroupRequestObject) (DeleteSharedPolicyGroupResponseObject, error) {
	groups := m.tenant(req.OrgId, req.EnvId).SharedPolicyGroups
	_, exists := groups.Get(req.Hrid)
	if !exists {
		return DeleteSharedPolicyGroup404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	if !isDryRun(req.Params.DryRun) {
		groups.DeleteByKey(req.Hrid)
	}
	return DeleteSharedPolicyGroup204Response{}, nil
}

// Portals

func (m *MockAPIM) UpsertPortal(_ context.Context, req UpsertPortalRequestObject) (UpsertPortalResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "portal"), req.Body.Hrid, req.Params.DryRun)
	portal := state[PortalState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).Portals, portal, req.Params.DryRun)
	return UpsertPortal200JSONResponse(portal), nil
}

func (m *MockAPIM) GetPortal(_ context.Context, req GetPortalRequestObject) (GetPortalResponseObject, error) {
	if portal, ok := m.tenant(req.OrgId, req.EnvId).Portals.Get(req.Hrid); ok {
		return GetPortal200JSONResponse(portal), nil
	}
	return GetPortal404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

// DeletePortal also deletes the portal's listings, documentations and links.
func (m *MockAPIM) DeletePortal(_ context.Context, req DeletePortalRequestObject) (DeletePortalResponseObject, error) {
	t := m.tenant(req.OrgId, req.EnvId)
	if !remove(t.Portals, req.Hrid) {
		return DeletePortal404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	removeChildren(t.PortalListings, req.Hrid)
	removeChildren(t.PortalDocumentations, req.Hrid)
	removeChildren(t.PortalLinks, req.Hrid)
	return DeletePortal204Response{}, nil
}

// Subscriptions

// UpsertApiSubscription answers 404 when the application or the API plan it references does not exist, as APIM does.
func (m *MockAPIM) UpsertApiSubscription(_ context.Context, req UpsertApiSubscriptionRequestObject) (UpsertApiSubscriptionResponseObject, error) {
	t := m.tenant(req.OrgId, req.EnvId)
	if missing := t.missingSubscriptionReference(req.ApiHrid, *req.Body); missing != "" {
		return UpsertApiSubscriptiondefaultJSONResponse{Body: hridNotFound(missing), StatusCode: http.StatusNotFound}, nil
	}
	status := m.status(at(req.OrgId, req.EnvId, "subscription").under("apiHrid", req.ApiHrid), req.Body.Hrid, nil)
	status["startingAt"] = startingAt(isChildOf(t.Subscriptions, req.ApiHrid, req.Body.Hrid))
	subscription := state[SubscriptionState](req.Body.WithDefaults(), status)
	t.Subscriptions.Put(newChild(subscription, req.ApiHrid))
	return UpsertApiSubscription200JSONResponse(subscription), nil
}

// missingSubscriptionReference returns the hrid of the application or plan the subscription names but that
// does not exist, checked in that order, or "" when both exist.
func (t *tenant) missingSubscriptionReference(apiHrid string, subscription Subscription) string {
	if _, ok := t.Applications.Get(subscription.ApplicationHrid); !ok {
		return subscription.ApplicationHrid
	}
	api, _ := t.Apis.Get(apiHrid)
	if api.Plans != nil {
		for _, plan := range *api.Plans {
			if plan.Hrid == subscription.PlanHrid {
				return ""
			}
		}
	}
	return subscription.PlanHrid
}

func (m *MockAPIM) GetApiSubscription(_ context.Context, req GetApiSubscriptionRequestObject) (GetApiSubscriptionResponseObject, error) {
	if subscription, ok := isChildOf(m.tenant(req.OrgId, req.EnvId).Subscriptions, req.ApiHrid, req.Hrid); ok {
		return GetApiSubscription200JSONResponse(subscription.Self), nil
	}
	return GetApiSubscription404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeleteApiSubscription(_ context.Context, req DeleteApiSubscriptionRequestObject) (DeleteApiSubscriptionResponseObject, error) {
	if !removeChild(m.tenant(req.OrgId, req.EnvId).Subscriptions, req.ApiHrid, req.Hrid) {
		return DeleteApiSubscription404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	return DeleteApiSubscription204Response{}, nil
}

// API documentations

func (m *MockAPIM) UpsertApiDocumentation(_ context.Context, req UpsertApiDocumentationRequestObject) (UpsertApiDocumentationResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "api-documentation").under("apiHrid", req.ApiHrid), req.Body.Hrid, req.Params.DryRun)
	documentation := state[DocumentationState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).ApiDocumentations, newChild(documentation, req.ApiHrid), req.Params.DryRun)
	return UpsertApiDocumentation200JSONResponse(documentation), nil
}

func (m *MockAPIM) GetApiDocumentation(_ context.Context, req GetApiDocumentationRequestObject) (GetApiDocumentationResponseObject, error) {
	if documentation, ok := isChildOf(m.tenant(req.OrgId, req.EnvId).ApiDocumentations, req.ApiHrid, req.Hrid); ok {
		return GetApiDocumentation200JSONResponse(documentation.Self), nil
	}
	return GetApiDocumentation404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeleteApiDocumentation(_ context.Context, req DeleteApiDocumentationRequestObject) (DeleteApiDocumentationResponseObject, error) {
	if !removeChild(m.tenant(req.OrgId, req.EnvId).ApiDocumentations, req.ApiHrid, req.Hrid) {
		return DeleteApiDocumentation404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	return DeleteApiDocumentation204Response{}, nil
}

// API links

func (m *MockAPIM) UpsertApiLink(_ context.Context, req UpsertApiLinkRequestObject) (UpsertApiLinkResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "api-link").under("apiHrid", req.ApiHrid), req.Body.Hrid, req.Params.DryRun)
	link := state[PortalLinkState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).ApiLinks, newChild(link, req.ApiHrid), req.Params.DryRun)
	return UpsertApiLink200JSONResponse(link), nil
}

func (m *MockAPIM) GetApiLink(_ context.Context, req GetApiLinkRequestObject) (GetApiLinkResponseObject, error) {
	if link, ok := isChildOf(m.tenant(req.OrgId, req.EnvId).ApiLinks, req.ApiHrid, req.Hrid); ok {
		return GetApiLink200JSONResponse(link.Self), nil
	}
	return GetApiLink404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeleteApiLink(_ context.Context, req DeleteApiLinkRequestObject) (DeleteApiLinkResponseObject, error) {
	if !removeChild(m.tenant(req.OrgId, req.EnvId).ApiLinks, req.ApiHrid, req.Hrid) {
		return DeleteApiLink404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	return DeleteApiLink204Response{}, nil
}

// Portal listings

func (m *MockAPIM) UpsertPortalListing(_ context.Context, req UpsertPortalListingRequestObject) (UpsertPortalListingResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "portal-listing").under("portalHrid", req.PortalHrid), req.Body.Hrid, req.Params.DryRun)
	listing := state[PortalListingState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).PortalListings, newChild(listing, req.PortalHrid), req.Params.DryRun)
	return UpsertPortalListing200JSONResponse(listing), nil
}

func (m *MockAPIM) GetPortalListing(_ context.Context, req GetPortalListingRequestObject) (GetPortalListingResponseObject, error) {
	if listing, ok := isChildOf(m.tenant(req.OrgId, req.EnvId).PortalListings, req.PortalHrid, req.Hrid); ok {
		return GetPortalListing200JSONResponse(listing.Self), nil
	}
	return GetPortalListing404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeletePortalListing(_ context.Context, req DeletePortalListingRequestObject) (DeletePortalListingResponseObject, error) {
	if !removeChild(m.tenant(req.OrgId, req.EnvId).PortalListings, req.PortalHrid, req.Hrid) {
		return DeletePortalListing404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	return DeletePortalListing204Response{}, nil
}

// Portal documentations

func (m *MockAPIM) UpsertPortalDocumentation(_ context.Context, req UpsertPortalDocumentationRequestObject) (UpsertPortalDocumentationResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "portal-documentation").under("portalHrid", req.PortalHrid), req.Body.Hrid, req.Params.DryRun)
	documentation := state[DocumentationState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).PortalDocumentations, newChild(documentation, req.PortalHrid), req.Params.DryRun)
	return UpsertPortalDocumentation200JSONResponse(documentation), nil
}

func (m *MockAPIM) GetPortalDocumentation(_ context.Context, req GetPortalDocumentationRequestObject) (GetPortalDocumentationResponseObject, error) {
	if documentation, ok := isChildOf(m.tenant(req.OrgId, req.EnvId).PortalDocumentations, req.PortalHrid, req.Hrid); ok {
		return GetPortalDocumentation200JSONResponse(documentation.Self), nil
	}
	return GetPortalDocumentation404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeletePortalDocumentation(_ context.Context, req DeletePortalDocumentationRequestObject) (DeletePortalDocumentationResponseObject, error) {
	if !removeChild(m.tenant(req.OrgId, req.EnvId).PortalDocumentations, req.PortalHrid, req.Hrid) {
		return DeletePortalDocumentation404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	return DeletePortalDocumentation204Response{}, nil
}

// Portal links

func (m *MockAPIM) UpsertPortalLink(_ context.Context, req UpsertPortalLinkRequestObject) (UpsertPortalLinkResponseObject, error) {
	status := m.status(at(req.OrgId, req.EnvId, "portal-link").under("portalHrid", req.PortalHrid), req.Body.Hrid, req.Params.DryRun)
	link := state[PortalLinkState](req.Body.WithDefaults(), status)
	save(m.tenant(req.OrgId, req.EnvId).PortalLinks, newChild(link, req.PortalHrid), req.Params.DryRun)
	return UpsertPortalLink200JSONResponse(link), nil
}

func (m *MockAPIM) GetPortalLink(_ context.Context, req GetPortalLinkRequestObject) (GetPortalLinkResponseObject, error) {
	if link, ok := isChildOf(m.tenant(req.OrgId, req.EnvId).PortalLinks, req.PortalHrid, req.Hrid); ok {
		return GetPortalLink200JSONResponse(link.Self), nil
	}
	return GetPortalLink404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
}

func (m *MockAPIM) DeletePortalLink(_ context.Context, req DeletePortalLinkRequestObject) (DeletePortalLinkResponseObject, error) {
	if !removeChild(m.tenant(req.OrgId, req.EnvId).PortalLinks, req.PortalHrid, req.Hrid) {
		return DeletePortalLink404JSONResponse{NotFoundJSONResponse(hridNotFound(req.Hrid))}, nil
	}
	return DeletePortalLink204Response{}, nil
}
