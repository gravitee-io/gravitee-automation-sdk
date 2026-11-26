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

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var onGiopsApi = []string{"--api", "giops-api"}

// newAPIMCLIWithApiAndApplication applies giops-api and giops-app, which giops-subscription subscribes.
func newAPIMCLIWithApiAndApplication(t *testing.T) cli {
	giops := newAPIMCLIWithApi(t)
	giops.mustRun("apim", "apply", "applications", "-f", testdata(t, "apim/application.yaml"), "--yes")
	return giops
}

func TestAPIMApplySubscription(t *testing.T) {
	giops := newAPIMCLIWithApiAndApplication(t)

	out := giops.mustRun("apim", "apply", "subscriptions", "-f", testdata(t, "apim/subscription.yaml"), "--yes")

	assert.Contains(t, out, "giops-subscription applied successfully")
	subscription := giops.getJSON("apim", "subscriptions", "giops-subscription", onGiopsApi...)
	assert.Equal(t, "giops-app", subscription["applicationHrid"])
	assert.Equal(t, "giops-plan", subscription["planHrid"])
}

func TestAPIMDryRunSubscriptionSavesNothing(t *testing.T) {
	giops := newAPIMCLIWithApiAndApplication(t)

	out := giops.mustRun("apim", "apply", "subscriptions", "-f", testdata(t, "apim/subscription.yaml"), "--dry-run")

	assert.Contains(t, out, "Dry run is not configured")
	giops.requireAbsent("apim", "subscriptions", "giops-subscription", onGiopsApi...)
}

func TestAPIMApplySubscriptionAfterSayingYes(t *testing.T) {
	giops := newAPIMCLIWithApiAndApplication(t)

	r := giops.runWithInput("y\n", "apim", "apply", "subscriptions", "-f", testdata(t, "apim/subscription.yaml"))

	assert.NoError(t, r.err, r.out)
	assert.Contains(t, r.out, "Dry run is not configured")
	assert.Contains(t, r.out, "Apply?")
	giops.getJSON("apim", "subscriptions", "giops-subscription", onGiopsApi...)
}

func TestAPIMGetSubscriptionAsTable(t *testing.T) {
	giops := newAPIMCLIWithApiAndApplication(t)
	giops.mustRun("apim", "apply", "subscriptions", "-f", testdata(t, "apim/subscription.yaml"), "--yes")

	out := giops.mustRun(append([]string{"apim", "get", "subscriptions", "giops-subscription"}, onGiopsApi...)...)

	assert.Regexp(t, `HRID\s+API\s+PLAN\s+APPLICATION`, out)
	assert.Regexp(t, `giops-subscription\s+giops-api\s+giops-plan\s+giops-app`, out)
}

func TestAPIMGetUnknownSubscription(t *testing.T) {
	r := newAPIMCLIWithApi(t).run(append([]string{"apim", "get", "subscriptions", "unknown"}, onGiopsApi...)...)

	assert.ErrorContains(t, r.err, "404")
}

func TestAPIMDeleteSubscription(t *testing.T) {
	cases := map[string]func(t *testing.T) []string{
		"by id":         func(*testing.T) []string { return append([]string{"giops-subscription"}, onGiopsApi...) },
		"from its file": func(t *testing.T) []string { return []string{"-f", testdata(t, "apim/subscription.yaml")} },
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			giops := newAPIMCLIWithApiAndApplication(t)
			giops.mustRun("apim", "apply", "subscriptions", "-f", testdata(t, "apim/subscription.yaml"), "--yes")

			out := giops.mustRun(append([]string{"apim", "delete", "subscriptions", "-y"}, target(t)...)...)

			assert.Contains(t, out, "giops-subscription deleted")
			giops.requireAbsent("apim", "subscriptions", "giops-subscription", onGiopsApi...)
		})
	}
}

func TestAPIMSubscriptionNeedsItsApi(t *testing.T) {
	for _, args := range [][]string{
		{"apim", "get", "subscriptions", "giops-subscription"},
		{"apim", "delete", "subscriptions", "-y", "giops-subscription"},
	} {
		r := newAPIMCLI(t).run(args...)

		assert.ErrorContains(t, r.err, "api identifier required", args)
	}
}

func TestAPIMApiFlagWinsOverTheSubscriptionFile(t *testing.T) {
	giops := newAPIMCLIWithApiAndApplication(t)
	giops.mustRun("apim", "apply", "apis", "-f", testdata(t, "apim/other-api.yaml"), "--yes")

	giops.mustRun("apim", "apply", "subscriptions", "-f", testdata(t, "apim/subscription.yaml"), "--api", "giops-other-api", "--yes")

	giops.getJSON("apim", "subscriptions", "giops-subscription", "--api", "giops-other-api")
	giops.requireAbsent("apim", "subscriptions", "giops-subscription", onGiopsApi...)
}

func TestAPIMSubscriptionOfAnUnknownApplication(t *testing.T) {
	giops := newAPIMCLIWithApi(t)

	r := giops.run("apim", "apply", "subscriptions", "-f", testdata(t, "apim/subscription.yaml"), "--yes")

	assert.ErrorContains(t, r.err, "404")
}
