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
	"os"
	"path/filepath"

	amsdk "github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	apimsdk "github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/config"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/actions"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/am"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim"
	"github.com/spf13/cobra"
)

func MainCommand() *cobra.Command {
	var main = &cobra.Command{
		Use:           "giops",
		Short:         "Do all Gravitee Automation with giops.",
		SilenceUsage:  true,
		SilenceErrors: true, // main prints them
	}
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	main.PersistentFlags().StringP(core.ConfigFlag, "c", filepath.Join(dir, "config.yaml"), "configuration file")

	amModule := core.Module{
		Name:          "am",
		ClientFactory: AM,
	}
	gammaModule := core.Module{
		Name:          "apim",
		ClientFactory: Gamma,
	}

	registerAMResourceConfigs(amModule, main)
	registerAPIMResourceConfigs(gammaModule, main)
	registerAMActionConfigs(amModule, main)
	registerAPIMActionConfigs(gammaModule, main)

	return main
}

func registerAPIMResourceConfigs(gammaModule core.Module, main *cobra.Command) {
	// Init registry
	registry := resources.NewRegistry()

	// Register resources
	registry.RegisterResource(resources.Command{
		Module: gammaModule,
		Kind:   *resources.NewKind[apimsdk.Group]("groups"),
		Getter: resources.Getter{
			Get:     apim.GetGroup,
			Columns: apim.GroupColumns,
		},
		Applier: resources.Applier{
			Put:               apim.PutGroup,
			DryRun:            apim.DryRunGroup,
			AdditionalOptions: resources.NewSpecificOptions(apim.GroupOptions()),
		},
		Deleter: resources.Deleter{
			Delete: apim.DeleteGroup,
		},
	})

	// Hook
	registry.HookCommands(main)
}

func registerAPIMActionConfigs(gammaModule core.Module, main *cobra.Command) {
	registry := actions.NewRegistry()

	registry.RegisterAction("check", actions.Command{
		Module: gammaModule,
		Run:    apim.Probe,
	})

	registry.HookActionCommands(main)
}

func registerAMResourceConfigs(amModule core.Module, main *cobra.Command) {
	// Init registry
	registry := resources.NewRegistry()

	// Register resources
	registry.RegisterResource(resources.Command{
		Kind:   *resources.NewKind[amsdk.Domain]("domains"),
		Module: amModule,
		Getter: resources.Getter{
			List:    am.ListDomain,
			Get:     am.GetDomain,
			Columns: am.DomainColumns,
		},
		Applier: resources.Applier{
			Put:    am.PutDomain,
			DryRun: am.DryRunDomain,
		},
		Deleter: resources.Deleter{
			Delete: am.DeleteDomain,
		},
	})

	registry.RegisterResource(resources.Command{
		Kind:   *resources.NewKind[am.IdentityProvider]("identityproviders").WithParent(am.IdentityProviderDomain, "domain"),
		Module: amModule,
		Getter: resources.Getter{
			List:    am.ListIdentityProvider,
			Get:     am.GetIdentityProvider,
			Columns: am.IdentityProviderColumns,
		},
		Applier: resources.Applier{
			Put:    am.PutIdentityProvider,
			DryRun: am.DryRunIdentityProvider,
		},
		Deleter: resources.Deleter{
			Delete: am.DeleteIdentityProvider,
		},
	})

	// Hook into cobra
	registry.HookCommands(main)
}

func registerAMActionConfigs(amModule core.Module, main *cobra.Command) {
	registry := actions.NewRegistry()

	registry.RegisterAction("check", actions.Command{
		Module:  amModule,
		Options: am.ProbeOptions,
		Run:     am.Probe,
	})

	registry.HookActionCommands(main)
}

func AM(c config.CLIConfig) (any, error) {
	ac, err := c.AMContext()
	if err != nil {
		return nil, err
	}
	return amsdk.NewAMClient(ac)
}

func Gamma(c config.CLIConfig) (any, error) {
	ac, err := c.APIMContext()
	if err != nil {
		return nil, err
	}
	return apimsdk.NewAPIMClient(ac)
}
