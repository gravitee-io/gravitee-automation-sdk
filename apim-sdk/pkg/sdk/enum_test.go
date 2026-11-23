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

package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// x-enum-varnames (overlays/models.yaml) names an enum's values by position: an upstream value added
// or reordered would shift every name after it, and still generate and compile. Each name is pinned
// to its value here.
func TestAuthzEnumValues(t *testing.T) {
	tests := []struct {
		name  string
		got   string
		want  string
		valid bool
	}{
		{"AuthzEntityKindPRINCIPAL", string(AuthzEntityKindPRINCIPAL), "PRINCIPAL", AuthzEntityKindPRINCIPAL.Valid()},
		{"AuthzEntityKindRESOURCE", string(AuthzEntityKindRESOURCE), "RESOURCE", AuthzEntityKindRESOURCE.Valid()},
		{"AuthzPolicyTypeMCP", string(AuthzPolicyTypeMCP), "MCP", AuthzPolicyTypeMCP.Valid()},
		{"AuthzPolicyTypeAGENT", string(AuthzPolicyTypeAGENT), "AGENT", AuthzPolicyTypeAGENT.Valid()},
		{"AuthzPolicyTypeMODEL", string(AuthzPolicyTypeMODEL), "MODEL", AuthzPolicyTypeMODEL.Valid()},
		{"AuthzPolicyTypeAPI", string(AuthzPolicyTypeAPI), "API", AuthzPolicyTypeAPI.Valid()},
		{"AuthzPolicyTypeCUSTOM", string(AuthzPolicyTypeCUSTOM), "CUSTOM", AuthzPolicyTypeCUSTOM.Valid()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
			assert.True(t, tt.valid)
		})
	}

	assert.False(t, AuthzEntityKind("GROUP").Valid())
	assert.False(t, AuthzPolicyType("TOOL").Valid())
}
