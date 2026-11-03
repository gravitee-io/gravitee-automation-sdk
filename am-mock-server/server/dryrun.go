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
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const dryRunMessage = "Mock server is in dry-run-reject mode, all PUT request are rejected on purpose when ?dryRun=true"

var dryRunErrors = []DryRunError{{Severity: new(SeverityError), Message: new(dryRunMessage)}}

// DryRun skips PUT persistence when ?dryRun=true and echoes the payload back, without a client secret value.
// When reject is true, the echoed payload carries a fixed dryRunErrors list.
func DryRun(reject bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isDryRun(r) || r.Method != http.MethodPut {
				next.ServeHTTP(w, r)
				return
			}
			// Read the whole body first: HTTP/1.x stops reading the request once the response starts.
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !reject && !isClientSecret(r) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(body)
				return
			}
			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if isClientSecret(r) {
				delete(payload, "value") // AM never returns a client secret value
			}
			if reject {
				payload["dryRunErrors"] = dryRunErrors
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(payload)
		})
	}
}

func isClientSecret(r *http.Request) bool {
	return strings.HasSuffix(r.URL.Path, "/secrets")
}

func isDryRun(r *http.Request) bool {
	ok, err := strconv.ParseBool(r.URL.Query().Get("dryRun"))
	return err == nil && ok
}
