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
	"fmt"
	"net/http"
	"strings"

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/auth"
)

type ParentChecker func(org, env, key string) bool

func DomainParentCheck(extract auth.RouteInfoExtractor, checker ParentChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := extract(r)
			if !strings.Contains(info.RoutePattern, "{domainKey}/") {
				next.ServeHTTP(w, r)
				return
			}
			if !parentExists(w, info.RouteParams, checker, "domainKey", "Domain key") {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func parentExists(w http.ResponseWriter, params map[string]string, checker ParentChecker, paramName string, paramLabel string) bool {
	if key, ok := params[paramName]; ok {
		if !checker(params["orgId"], params["envId"], key) {
			auth.WriteError(w, http.StatusNotFound, fmt.Sprintf("%s [%s] not found", paramLabel, key))
			return false
		}
	}
	return true
}

// applicationIdentity is the identity of a stored application, and so the parent identity of its client secrets.
func applicationIdentity(domainKey, applicationKey string) string {
	return domainKey + "/" + applicationKey
}
