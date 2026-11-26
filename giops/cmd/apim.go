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
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/config"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/actions"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim"
	"github.com/spf13/cobra"
)

func Gamma(c config.CLIConfig) (any, error) {
	ac, err := c.APIMContext()
	if err != nil {
		return nil, err
	}
	return apimsdk.NewAPIMClient(ac)
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

// registerAPIMActionConfigs adds the actions of a module that uses the APIM client: apim, portal, aim, authz.
func registerAPIMActionConfigs(gammaModule core.Module, main *cobra.Command) {
	registry := actions.NewRegistry()

	registry.RegisterAction("check", actions.Command{
		Module: gammaModule,
		Run:    apim.Probe,
	})

	registry.HookActionCommands(main)
}
