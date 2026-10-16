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

func TestOneOf_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		wire    string
		value   any
		variant func(any) any
	}{
		{
			name:    "listener",
			wire:    `{"type":"KAFKA","host":"kafka.local","port":9092}`,
			value:   &OneOfListener{},
			variant: func(v any) any { return v.(*OneOfListener).Kafka },
		},
		{
			name:    "selector",
			wire:    `{"type":"CONDITION","condition":"{#request.headers['x'] != null}"}`,
			value:   &OneOfSelector{},
			variant: func(v any) any { return v.(*OneOfSelector).Condition },
		},
		{
			name:    "catalog MCP server auth",
			wire:    `{"type":"HEADER","name":"X-Token"}`,
			value:   &OneOfAimCatalogMcpServerAuth{},
			variant: func(v any) any { return v.(*OneOfAimCatalogMcpServerAuth).Header },
		},
		{
			name:    "MCP proxy upstream auth",
			wire:    `{"type":"API_KEY","apiKeyHeader":"X-Api-Key"}`,
			value:   &OneOfAimMcpProxyUpstreamAuth{},
			variant: func(v any) any { return v.(*OneOfAimMcpProxyUpstreamAuth).ApiKey },
		},
		{
			name:    "MCP proxy plan security",
			wire:    `{"type":"API_KEY","source":"HEADER","apiKeyHeader":"X-Key","propagateApiKey":true}`,
			value:   &OneOfAimMcpProxyPlanSecurity{},
			variant: func(v any) any { return v.(*OneOfAimMcpProxyPlanSecurity).ApiKey },
		},
		{
			name:    "MCP proxy flow selector",
			wire:    `{"type":"MCP","methods":["tools/call"]}`,
			value:   &OneOfAimMcpProxyFlowSelector{},
			variant: func(v any) any { return v.(*OneOfAimMcpProxyFlowSelector).Mcp },
		},
		{
			name:    "MCP proxy identity provider",
			wire:    `{"type":"GRAVITEE_AM","name":"am"}`,
			value:   &OneOfAimMcpProxyIdentityProvider{},
			variant: func(v any) any { return v.(*OneOfAimMcpProxyIdentityProvider).GraviteeAm },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, json.Unmarshal([]byte(tt.wire), tt.value))
			assert.NotNil(t, tt.variant(tt.value))

			out, err := json.Marshal(tt.value)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wire, string(out))
		})
	}
}

func TestOneOf_UnknownTypeErrors(t *testing.T) {
	var listener OneOfListener
	err := json.Unmarshal([]byte(`{"type":"GRPC"}`), &listener)
	assert.ErrorContains(t, err, `unknown type "GRPC"`)
}

func TestOneOf_TwoVariantsErrors(t *testing.T) {
	listener := OneOfListener{Http: &HttpListener{}, Tcp: &TcpListener{}}
	_, err := json.Marshal(listener)
	assert.ErrorContains(t, err, "more than one variant set")
}

func TestOneOf_ZeroValueIsNull(t *testing.T) {
	out, err := json.Marshal(OneOfSelector{})
	require.NoError(t, err)
	assert.Equal(t, "null", string(out))
}

func TestOneOf_NestedVariantKeepsItsFields(t *testing.T) {
	wire := `{"serverUrl":"https://mcp.local","upstreamAuth":{"type":"API_KEY","apiKeyHeader":"X-Api-Key"}}`
	var proxy AimMcpProxyProxy
	require.NoError(t, json.Unmarshal([]byte(wire), &proxy))
	require.NotNil(t, proxy.UpstreamAuth.ApiKey)
	assert.Equal(t, "X-Api-Key", proxy.UpstreamAuth.ApiKey.ApiKeyHeader)

	out, err := json.Marshal(proxy)
	require.NoError(t, err)
	assert.JSONEq(t, wire, string(out))
}
