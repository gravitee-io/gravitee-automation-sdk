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

// Package naming turns resource names into Kubernetes object names and settles the
// collisions that follow, since names are not unique in APIM.
package naming

import (
	"regexp"
	"strings"
)

// MaxLength is the DNS-1123 label limit. GKO derives an HRID from <namespace>-<name>, so short names win.
const MaxLength = 63

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slug converts a display name into a DNS-1123 label: lower case, runs of anything that is not
// a letter or a digit become one dash, no leading or trailing dash, at most 63 characters.
// "Petstore API (v2)" becomes "petstore-api-v2". An empty result becomes "unnamed".
func Slug(name string) string {
	s := nonAlnum.ReplaceAllString(strings.ToLower(name), "-")
	s = strings.Trim(s, "-")
	if len(s) > MaxLength {
		s = strings.Trim(s[:MaxLength], "-")
	}
	if s == "" {
		return "unnamed"
	}
	return s
}

// IsSlug reports whether s is already a valid object name as Slug would produce it.
func IsSlug(s string) bool {
	return s != "" && s == Slug(s)
}
