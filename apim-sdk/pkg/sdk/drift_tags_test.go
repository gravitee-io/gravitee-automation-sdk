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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The drift overlay tags GKO relies on to accept values APIM fills in on its own; dropping one makes GKO report drift
// on every update.
func TestDriftTags_RemoteDefaults(t *testing.T) {
	const pluginConfig = "unstructured:ignore-remote-default"
	tests := []struct {
		value any
		field string
		want  string
	}{
		{HttpSelector{}, "Path", "ignore-remote-default:/"},
		{HttpSelector{}, "PathOperator", "ignore-remote-default:STARTS_WITH"},
		{Entrypoint{}, "Qos", "ignore-remote-default:AUTO"},
		{FlowExecution{}, "Mode", "ignore-remote-default:case-insensitive,DEFAULT"},
		{AimMcpProxyGenericIdentityProvider{}, "UserInfoEndpoint", "ignore-remote-default:,/userinfo"},
		{APIResource{}, "Configuration", pluginConfig},
		{BaseStep{}, "Configuration", pluginConfig},
		{StepV4{}, "Configuration", pluginConfig},
		{EndpointV4{}, "Configuration", pluginConfig},
		{EndpointV4{}, "SharedConfigurationOverride", pluginConfig},
		{EndpointGroupV4{}, "SharedConfiguration", pluginConfig},
		{PageSource{}, "Configuration", pluginConfig},
		{PlanSecurity{}, "Configuration", pluginConfig},
		{ServiceV4{}, "Configuration", pluginConfig},
		{AimMcpProxyFlowStep{}, "Configuration", pluginConfig},
		{SubscriptionConsumerConfiguration{}, "EntrypointConfiguration", pluginConfig},
	}
	for _, tt := range tests {
		typ := reflect.TypeOf(tt.value)
		t.Run(typ.Name()+"."+tt.field, func(t *testing.T) {
			f, ok := typ.FieldByName(tt.field)
			if !ok {
				t.Fatalf("%s has no field %s", typ.Name(), tt.field)
			}
			assert.Equal(t, tt.want, f.Tag.Get("drift"))
		})
	}
}
