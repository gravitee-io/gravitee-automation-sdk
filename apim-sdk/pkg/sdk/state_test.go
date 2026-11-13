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
