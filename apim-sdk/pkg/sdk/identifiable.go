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

import "strings"

// Identities of resources and list items: the keys GKO's lifecycle stores resources by
// and its drift pairs list items by.

// Identity is shared by API and Application members.
func (m Member) Identity() string {
	return m.Source + ":" + m.SourceId
}

func (m GroupMember) Identity() string {
	return m.Source + ":" + m.SourceId
}

func (m Metadata) Identity() string {
	return m.Name
}

func (k ApiKey) Identity() string {
	return k.Key
}

// Identity is the certificate name, or its trimmed content when it has none.
func (c ClientCertificate) Identity() string {
	if c.Name != "" {
		return c.Name
	}
	return strings.TrimSpace(c.Content)
}

// A resource is identified by its hrid.

func (a AimCatalogMcpServer) Identity() string { return a.Hrid }
func (a AimMcpProxy) Identity() string         { return a.Hrid }
func (a ApiV4) Identity() string               { return a.Hrid }
func (a Application) Identity() string         { return a.Hrid }
func (d Dictionary) Identity() string          { return d.Hrid }
func (d Documentation) Identity() string       { return d.Hrid }
func (g Group) Identity() string               { return g.Hrid }
func (p Portal) Identity() string              { return p.Hrid }
func (p PortalLink) Identity() string          { return p.Hrid }
func (p PortalListing) Identity() string       { return p.Hrid }
func (p PortalTheme) Identity() string         { return p.Hrid }
func (s SharedPolicyGroup) Identity() string   { return s.Hrid }
func (s Subscription) Identity() string        { return s.Hrid }
