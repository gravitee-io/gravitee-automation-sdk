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

package test

import "testing"

// newAPIMCLIWithApi applies giops-api, the parent of the API pages and subscriptions.
func newAPIMCLIWithApi(t *testing.T) cli {
	giops := newAPIMCLI(t)
	giops.mustRun("apim", "apply", "apis", "-f", testdata(t, "apim/api.yaml"), "--yes")
	return giops
}

func TestAPIMApiDocumentation(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "apim",
		resource: "documentations",
		file:     "apim/documentation.yaml",
		hrid:     "giops-doc",
		field:    "name",
		value:    "Giops doc",
		header:   `HRID\s+TYPE\s+LOCATION`,
		row:      `giops-doc\s+GRAVITEE_MARKDOWN`,
		parent:   "api",
		parentID: "giops-api",
		setup:    newAPIMCLIWithApi,
	})
}

func TestAPIMApiLink(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "apim",
		resource: "links",
		file:     "apim/link.yaml",
		hrid:     "giops-link",
		field:    "href",
		value:    "https://www.gravitee.io",
		header:   `HRID\s+APIHRID`,
		row:      `giops-link\s+giops-api`,
		parent:   "api",
		parentID: "giops-api",
		setup:    newAPIMCLIWithApi,
	})
}
