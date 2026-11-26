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

var inGiopsPortal = []string{"--portal", "giops-portal"}

// newPortalCLIWithPortal applies giops-portal, the parent of the portal's pages, links and listings, and
// giops-api, which the listing shows.
func newPortalCLIWithPortal(t *testing.T) cli {
	giops := newGammaCLI(t, "portal")
	giops.mustRun("portal", "apply", "portals", "-f", testdata(t, "portal/portal.yaml"), "--yes")
	giops.mustRun("apim", "apply", "apis", "-f", testdata(t, "apim/api.yaml"), "--yes")
	return giops
}

func TestPortalPortal(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "portal",
		resource: "portals",
		file:     "portal/portal.yaml",
		hrid:     "giops-portal",
		field:    "name",
		value:    "Giops portal",
		header:   `HRID\s+NAME`,
		row:      `giops-portal\s+Giops portal`,
	})
}

func TestPortalTheme(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "portal",
		resource: "themes",
		file:     "portal/theme.yaml",
		hrid:     "giops-theme",
		field:    "name",
		value:    "Giops theme",
		header:   `HRID\s+NAME`,
		row:      `giops-theme\s+Giops theme`,
	})
}

func TestPortalDocumentation(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "portal",
		resource: "documentations",
		file:     "portal/documentation.yaml",
		hrid:     "giops-guide",
		field:    "name",
		value:    "Giops guide",
		header:   `HRID\s+TYPE\s+LOCATION`,
		row:      `giops-guide\s+GRAVITEE_MARKDOWN`,
		parent:   "portal",
		parentID: "giops-portal",
		setup:    newPortalCLIWithPortal,
	})
}

func TestPortalLink(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "portal",
		resource: "links",
		file:     "portal/link.yaml",
		hrid:     "giops-link",
		field:    "href",
		value:    "https://www.gravitee.io",
		header:   `HRID\s+PORTALHRID`,
		row:      `giops-link\s+giops-portal`,
		parent:   "portal",
		parentID: "giops-portal",
		setup:    newPortalCLIWithPortal,
	})
}

func TestPortalListing(t *testing.T) {
	testGammaResource(t, gammaResource{
		module:   "portal",
		resource: "listings",
		file:     "portal/listing.yaml",
		hrid:     "giops-listing",
		field:    "hrid",
		value:    "giops-listing",
		header:   `HRID\s+PORTALHRID`,
		row:      `giops-listing\s+giops-portal`,
		parent:   "portal",
		parentID: "giops-portal",
		setup:    newPortalCLIWithPortal,
	})
}

func TestPortalFlagWinsOverTheFile(t *testing.T) {
	giops := newPortalCLIWithPortal(t)
	giops.mustRun("portal", "apply", "portals", "-f", testdata(t, "portal/other-portal.yaml"), "--yes")

	giops.mustRun("portal", "apply", "documentations", "-f", testdata(t, "portal/documentation.yaml"), "--portal", "giops-other-portal", "--yes")

	giops.getJSON("portal", "documentations", "giops-guide", "--portal", "giops-other-portal")
	giops.requireAbsent("portal", "documentations", "giops-guide", inGiopsPortal...)
}

func TestPortalDeletePortalRemovesItsChildren(t *testing.T) {
	giops := newPortalCLIWithPortal(t)
	children := []struct{ resource, file, hrid string }{
		{"documentations", "portal/documentation.yaml", "giops-guide"},
		{"links", "portal/link.yaml", "giops-link"},
		{"listings", "portal/listing.yaml", "giops-listing"},
	}
	for _, child := range children {
		giops.mustRun("portal", "apply", child.resource, "-f", testdata(t, child.file), "--yes")
	}

	out := giops.mustRun("portal", "delete", "portals", "-y", "giops-portal")

	assert.Contains(t, out, "giops-portal deleted")
	for _, child := range children {
		giops.requireAbsent("portal", child.resource, child.hrid, inGiopsPortal...)
	}
}
