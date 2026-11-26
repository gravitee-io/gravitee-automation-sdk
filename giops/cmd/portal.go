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

package cmd

import (
	apimsdk "github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim/portal"
	"github.com/spf13/cobra"
)

func registerPortalResourceConfigs(portalModule core.Module, main *cobra.Command) {
	registry := resources.NewRegistry()

	registry.RegisterResource(resources.Command{
		Module: portalModule,
		Kind:   *resources.NewKind[apimsdk.Portal]("portals"),
		Getter: resources.Getter{
			Get:     portal.GetPortal,
			Columns: portal.PortalColumns,
		},
		Applier: resources.Applier{
			Put:    portal.PutPortal,
			DryRun: portal.DryRunPortal,
		},
		Deleter: resources.Deleter{
			Delete: portal.DeletePortal,
		},
	})

	registry.RegisterResource(resources.Command{
		Module: portalModule,
		Kind:   *resources.NewKind[apimsdk.PortalTheme]("themes"),
		Getter: resources.Getter{
			Get:     portal.GetTheme,
			Columns: portal.ThemeColumns,
		},
		Applier: resources.Applier{
			Put:    portal.PutTheme,
			DryRun: portal.DryRunTheme,
		},
		Deleter: resources.Deleter{
			Delete: portal.DeleteTheme,
		},
	})

	registry.RegisterResource(resources.Command{
		Module: portalModule,
		Kind:   *resources.NewKind[portal.Documentation]("documentations").WithParent(portal.DocumentationPortal, "portal"),
		Getter: resources.Getter{
			Get:     portal.GetDocumentation,
			Columns: portal.DocumentationColumns,
		},
		Applier: resources.Applier{
			Put:    portal.PutDocumentation,
			DryRun: portal.DryRunDocumentation,
		},
		Deleter: resources.Deleter{
			Delete: portal.DeleteDocumentation,
		},
	})

	registry.RegisterResource(resources.Command{
		Module: portalModule,
		Kind:   *resources.NewKind[portal.Link]("links").WithParent(portal.LinkPortal, "portal"),
		Getter: resources.Getter{
			Get:     portal.GetLink,
			Columns: portal.LinkColumns,
		},
		Applier: resources.Applier{
			Put:    portal.PutLink,
			DryRun: portal.DryRunLink,
		},
		Deleter: resources.Deleter{
			Delete: portal.DeleteLink,
		},
	})

	registry.RegisterResource(resources.Command{
		Module: portalModule,
		Kind:   *resources.NewKind[portal.Listing]("listings").WithParent(portal.ListingPortal, "portal"),
		Getter: resources.Getter{
			Get:     portal.GetListing,
			Columns: portal.ListingColumns,
		},
		Applier: resources.Applier{
			Put:    portal.PutListing,
			DryRun: portal.DryRunListing,
		},
		Deleter: resources.Deleter{
			Delete: portal.DeleteListing,
		},
	})

	registry.HookCommands(main)
}
