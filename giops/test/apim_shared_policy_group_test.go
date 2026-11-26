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

func TestAPIMSharedPolicyGroup(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "apim",
		resource: "shared-policy-groups",
		file:     "apim/shared-policy-group.yaml",
		hrid:     "giops-spg",
		field:    "name",
		value:    "Giops shared policy group",
		header:   `HRID\s+NAME\s+DESCRIPTION\s+APITYPE`,
		row:      `giops-spg\s+Giops shared policy group\s+PROXY`,
	})
}

func TestAPIMDeleteSharedPolicyGroupWithItsDryRun(t *testing.T) {
	testDeleteDryRun(t, "apim", "shared-policy-groups", "apim/shared-policy-group.yaml", "giops-spg")
}
