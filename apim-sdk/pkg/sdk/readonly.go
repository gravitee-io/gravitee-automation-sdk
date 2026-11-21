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

import "slices"

// clearReadOnly methods drop, from the PUT body a State's Spec() builds, the nested read-only fields the platform
// refuses on PUT. Spec() shares nested values with the State, so each method copies what it clears: the State keeps
// them. The read-only fields APIM honours on PUT stay in Spec(): readOnlyInPutBodies (readonly_test.go) lists them.

// clearReadOnly drops each studio tool's entityId, which AIM 5.1.2+ refuses on PUT.
func (p *AimMcpProxy) clearReadOnly() {
	if p.Studio == nil || p.Studio.Tools == nil {
		return
	}
	studio := *p.Studio
	studio.Tools = slices.Clone(studio.Tools)
	for i := range studio.Tools {
		studio.Tools[i].EntityId = nil
	}
	p.Studio = &studio
}
