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

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/store"
	"github.com/stretchr/testify/assert"
)

var (
	_ store.Identifiable = Member{}
	_ store.Identifiable = GroupMember{}
	_ store.Identifiable = Metadata{}
	_ store.Identifiable = ApiKey{}
	_ store.Identifiable = ClientCertificate{}
)

func TestIdentity(t *testing.T) {
	tests := []struct {
		name string
		item store.Identifiable
		want string
	}{
		{"member", Member{Source: "memory", SourceId: "john", Role: "OWNER"}, "memory:john"},
		{"group member", GroupMember{Source: "gravitee", SourceId: "jane"}, "gravitee:jane"},
		{"metadata", Metadata{Name: "team", Key: new("team")}, "team"},
		{"api key", ApiKey{Key: "secret-key"}, "secret-key"},
		{"named certificate", ClientCertificate{Name: "client", Content: "PEM"}, "client"},
		{"unnamed certificate", ClientCertificate{Content: "\n  PEM\n"}, "PEM"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.item.Identity())
		})
	}
}
