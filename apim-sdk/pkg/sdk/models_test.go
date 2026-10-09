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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The AIM fragment models its unions as a base schema extended by subtypes; overlays/models.yaml
// keeps the subtype fields, which the generated base struct would otherwise drop on decode.
func TestDiscriminatedSubtypeFieldsSurviveARoundTrip(t *testing.T) {
	in := `{"name":"gold","security":{"type":"API_KEY","source":"HEADER","apiKeyHeader":"X-Key"}}`
	var plan AimMcpProxyPlan
	require.NoError(t, json.Unmarshal([]byte(in), &plan))

	out, err := json.Marshal(plan)
	require.NoError(t, err)
	assert.JSONEq(t, in, string(out))
}
