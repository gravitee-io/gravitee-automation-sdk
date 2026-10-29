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

func (d Domain) Identity() string {
	return d.Key
}

func (d DataPlane) Identity() string {
	return d.Id
}

func (a Application) Identity() string {
	return a.Key
}

func (c Certificate) Identity() string {
	return c.Key
}

func (i IdentityProvider) Identity() string {
	return i.Key
}

func (r Reporter) Identity() string {
	return r.Key
}
