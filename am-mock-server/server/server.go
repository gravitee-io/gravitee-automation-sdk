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

// Package server is the AM Automation mock HTTP server and in-memory StrictServer implementation.
package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/auth"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

// BasePath is the default API prefix ("/automation").
const BasePath = "/automation"

// New mounts the API at BasePath with no auth.
func New(impl *MockAM) http.Handler {
	return NewWithPath(impl, BasePath, nil, false)
}

// NewWithPath mounts the API at basePath, and the faults of mockAM at FaultsPath. nil Registry leaves the API open.
// nil impl serves Unimplemented, with no faults.
func NewWithPath(mockAM *MockAM, basePath string, reg *auth.Registry, dryRunReject bool) http.Handler {
	api := newAPI(mockAM, basePath, reg, dryRunReject)
	if mockAM == nil {
		return api
	}
	root := chi.NewRouter()
	routeFaultsEndpoint(root, mockAM.Faults)
	root.Mount("/", api)
	return root
}

func newAPI(mockAM *MockAM, basePath string, reg *auth.Registry, dryRunReject bool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	// First in the list is innermost: dry-run applies to requests that passed auth, parent and spec validation.
	// A fault answers after auth and the parent check, before spec validation and dry-run.
	middlewares := []MiddlewareFunc{DryRun(dryRunReject), requestValidator(basePath)}
	if mockAM != nil {
		middlewares = append(middlewares, injectFaults(mockAM.Faults))
	}
	if reg != nil {
		r.Use(auth.AuthnMiddleware(reg))
		middlewares = append(middlewares, auth.AuthzMiddleware(reg, chiRouteInfoExtractor()))
	}
	middlewares = append(middlewares, DomainParentCheck(chiRouteInfoExtractor(), func(org, env, key string) bool {
		_, exists := mockAM.getTenant(orgEnv{org: org, env: env}).Domains.Get(key)
		return exists
	}))

	var si ServerInterface = Unimplemented{}
	if mockAM != nil {
		si = NewStrictHandler(mockAM, nil)
	}

	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:     basePath,
		BaseRouter:  r,
		Middlewares: middlewares,
	})
}

// requestValidator rejects requests that do not match the OpenAPI spec with a 400.
// Read-only properties are accepted and ignored, as the server sets them.
func requestValidator(basePath string) MiddlewareFunc {
	spec, err := GetSwagger()
	if err != nil {
		panic(fmt.Errorf("loading embedded OpenAPI spec: %w", err))
	}
	spec.Servers = openapi3.Servers{{URL: basePath}}
	openapi3.SchemaErrorDetailsDisabled = true // keep 400 messages to the failing field, not the whole schema
	return nethttpmiddleware.OapiRequestValidatorWithOptions(spec, &nethttpmiddleware.Options{
		SilenceServersWarning: true,
		Options: openapi3filter.Options{
			AuthenticationFunc:         openapi3filter.NoopAuthenticationFunc,
			ExcludeReadOnlyValidations: true,
		},
		ErrorHandlerWithOpts: func(_ context.Context, err error, w http.ResponseWriter, _ *http.Request, opts nethttpmiddleware.ErrorHandlerOpts) {
			auth.WriteError(w, opts.StatusCode, err.Error())
		},
	})
}

func chiRouteInfoExtractor() auth.RouteInfoExtractor {
	return func(r *http.Request) auth.RouteInfo {
		context := chi.RouteContext(r.Context())
		return auth.RouteInfo{
			Method:       r.Method,
			RoutePattern: context.RoutePattern(),
			RouteParams: func(params chi.RouteParams) map[string]string {
				r := make(map[string]string)
				for i, key := range params.Keys {
					r[key] = params.Values[i]
				}
				return r
			}(context.URLParams),
		}
	}

}
