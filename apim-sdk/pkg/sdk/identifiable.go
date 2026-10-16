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

// Identities of list items, the keys GKO's drift pairs them by.

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
