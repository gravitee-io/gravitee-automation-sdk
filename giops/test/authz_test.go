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

func TestAuthzCheck(t *testing.T) {
	newGammaCLI(t, "authz").mustRun("authz", "check")
}

func TestAuthzSchema(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "authz",
		resource: "schemas",
		file:     "authz/schema.yaml",
		hrid:     "giops-schema",
		field:    "name",
		value:    "Giops schema",
		header:   `HRID\s+TARGET`,
		row:      `giops-schema\s+giops@eu`,
	})
}

func TestAuthzPdp(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "authz",
		resource: "pdps",
		file:     "authz/pdp.yaml",
		hrid:     "giops-pdp",
		field:    "name",
		value:    "Giops PDP",
		header:   `HRID\s+ROUTINGSCOPE\s+ENDPOINT\s+STRICT`,
		row:      `giops-pdp\s+.*/authz/giops-eu`,
	})
}

func TestAuthzEntity(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "authz",
		resource: "entities",
		file:     "authz/entity.yaml",
		hrid:     "giops-entity",
		field:    "entityId",
		value:    "giops-agents",
		header:   `HRID\s+ENTITYID\s+ENTITYTYPE`,
		row:      `giops-entity\s+giops-agents\s+Group`,
	})
}

func TestAuthzPolicy(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "authz",
		resource: "policies",
		file:     "authz/policy.yaml",
		hrid:     "giops-policy",
		field:    "name",
		value:    "Giops policy",
		header:   `HRID\s+ENTITYID\s+POLICYSTATUS`,
		row:      `giops-policy\s+policy.giops`,
	})
}

func TestAuthzDeleteWithItsDryRun(t *testing.T) {
	for _, r := range []struct{ resource, file, hrid string }{
		{"pdps", "authz/pdp.yaml", "giops-pdp"},
		{"entities", "authz/entity.yaml", "giops-entity"},
		{"policies", "authz/policy.yaml", "giops-policy"},
	} {
		t.Run(r.resource, func(t *testing.T) {
			testDeleteDryRun(t, "authz", r.resource, r.file, r.hrid)
		})
	}
}
