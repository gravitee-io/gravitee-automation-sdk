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

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/auth"
)

// ParentChecker reports whether the parent resource key exists in the org/env.
type ParentChecker func(org, env, key string) bool

// ParentCheck answers 404 when a route parameter names a parent (a key of parents) that does not exist,
// whatever the method, as APIM does.
func ParentCheck(extract auth.RouteInfoExtractor, parents map[string]ParentChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			params := extract(r).RouteParams
			for param, exists := range parents {
				if key, ok := params[param]; ok && !exists(params["orgId"], params["envId"], key) {
					auth.WriteError(w, http.StatusNotFound, hridNotFoundMessage(key))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
