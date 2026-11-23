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

import "time"

// Validity window of a client certificate. APIM leaves expired and not yet
// active certificates out of its answers, so GKO's drift ignores them.

// Expired reports whether EndsAt is set and already in the past.
// A missing or unparseable EndsAt is not expired.
func (c ClientCertificate) Expired() bool {
	t, ok := parseRFC3339(c.EndsAt)
	return ok && time.Now().After(t)
}

// Scheduled reports whether StartsAt is set and still in the future.
// A missing or unparseable StartsAt is not scheduled.
func (c ClientCertificate) Scheduled() bool {
	t, ok := parseRFC3339(c.StartsAt)
	return ok && time.Now().Before(t)
}

func parseRFC3339(value *string) (time.Time, bool) {
	if value == nil {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, *value); err == nil {
		return t, true
	}
	t, err := time.Parse(time.RFC3339Nano, *value)
	return t, err == nil
}
