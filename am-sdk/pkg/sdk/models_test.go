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
	"github.com/stretchr/testify/require"
)

func TestWithDefaults_FillsListItems(t *testing.T) {
	scopes := []AutomationScopeSettings{{Scope: "openid"}}
	app := Application{Settings: &AutomationApplicationSettings{Oauth: &AutomationApplicationOAuthSettings{ScopeSettings: scopes}}}

	got := app.WithDefaults().Settings.Oauth.ScopeSettings

	require.Len(t, got, 1)
	assert.Equal(t, "openid", got[0].Scope)
	assert.Equal(t, new(false), got[0].DefaultScope)
	assert.Equal(t, new(false), got[0].RequiredScope)
	assert.Nil(t, scopes[0].RequiredScope, "WithDefaults must not change the caller's list")
}
