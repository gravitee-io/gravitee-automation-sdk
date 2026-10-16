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

package server

// API and portal documentations share DocumentationState, API and portal links share PortalLinkState:
// they live in separate stores.

func (s ApiV4State) Identity() string               { return s.Hrid }
func (s ApplicationState) Identity() string         { return s.Hrid }
func (s GroupState) Identity() string               { return s.Hrid }
func (s DictionaryState) Identity() string          { return s.Hrid }
func (s SharedPolicyGroupState) Identity() string   { return s.Hrid }
func (s PortalState) Identity() string              { return s.Hrid }
func (s PortalThemeState) Identity() string         { return s.Hrid }
func (s AimCatalogMcpServerState) Identity() string { return s.Hrid }
func (s AimMcpProxyState) Identity() string         { return s.Hrid }
func (s SubscriptionState) Identity() string        { return s.Hrid }
func (s DocumentationState) Identity() string       { return s.Hrid }
func (s PortalLinkState) Identity() string          { return s.Hrid }
func (s PortalListingState) Identity() string       { return s.Hrid }
