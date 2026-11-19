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

func TestSpec_KeepsOnlyTheFieldsAPutSends(t *testing.T) {
	members := []GroupMember{{Source: "memory", SourceId: "john", Roles: map[string]string{"API": "USER"}}}
	roles := &GroupDefaultMemberRoles{Api: new("USER")}
	state := GroupState{
		Hrid:               "developers",
		Name:               "Developers",
		Members:            members,
		DefaultMemberRoles: roles,
		NotifyMembers:      new(true),
		Id:                 new("8f1c"),
		OrganizationId:     new("DEFAULT"),
		EnvironmentId:      new("DEFAULT"),
		MemberCount:        new(int64(1)),
		Errors:             &Errors{Warning: []string{"ignored"}},
	}

	assert.Equal(t, Group{
		Hrid:               "developers",
		Name:               "Developers",
		Members:            members,
		DefaultMemberRoles: roles,
		NotifyMembers:      new(true),
	}, state.Spec())
}

func TestSpec_ConvertsTheApplicationStatus(t *testing.T) {
	state := ApplicationState{Hrid: "app", Status: new(ApplicationStateStatusARCHIVED)}

	assert.Equal(t, new(ApplicationSpecStatusARCHIVED), state.Spec().Status)
}

func TestSpec_DropsTheMcpAuthPolicyRefs(t *testing.T) {
	server := AimCatalogMcpServerState{
		Hrid:          "github",
		EntityId:      "mcp-server.github",
		AuthPolicyRef: new(`MCPServer::"github"`),
		Tools:         []AimCatalogMcpToolStatus{{Name: new("create_issue"), AuthPolicyRef: new(`MCPTool::"github.create_issue"`)}},
	}
	assert.Equal(t, AimCatalogMcpServer{Hrid: "github", EntityId: "mcp-server.github"}, server.Spec())

	proxy := AimMcpProxyState{
		Hrid:  "studio",
		Mode:  AimMcpProxyModeSTUDIO,
		Tools: []AimMcpProxyToolStatus{{Server: new("github"), AuthPolicyRef: new(`MCPTool::"github.create_issue"`)}},
	}
	assert.Equal(t, AimMcpProxy{Hrid: "studio", Mode: AimMcpProxyModeSTUDIO}, proxy.Spec())
}

// GKO falls back on studio.tools when the state has no tools at all, not when it has none to report.
func TestAimMcpProxyState_KeepsAbsentToolsApartFromEmpty(t *testing.T) {
	var absent, empty AimMcpProxyState
	require.NoError(t, json.Unmarshal([]byte(`{"hrid":"studio","mode":"STUDIO"}`), &absent))
	require.NoError(t, json.Unmarshal([]byte(`{"hrid":"studio","mode":"STUDIO","tools":[]}`), &empty))

	assert.Nil(t, absent.Tools)
	assert.Nil(t, absent.WithDefaults().Tools)
	assert.NotNil(t, empty.Tools)
	assert.Empty(t, empty.Tools)
	assert.NotNil(t, empty.WithDefaults().Tools)
}
