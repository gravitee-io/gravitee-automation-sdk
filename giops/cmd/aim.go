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
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim/aim"
	"github.com/spf13/cobra"
)

func registerAIMResourceConfigs(aimModule core.Module, main *cobra.Command) {
	registry := resources.NewRegistry()

	registry.RegisterResource(resources.Command{
		Module: aimModule,
		Kind:   *resources.NewKind[apimsdk.AimMcpProxy]("mcp-proxies"),
		Getter: resources.Getter{
			Get:     aim.GetMcpProxy,
			Columns: aim.McpProxyColumns,
		},
		Applier: resources.Applier{
			Put:    aim.PutMcpProxy,
			DryRun: aim.DryRunMcpProxy,
		},
		Deleter: resources.Deleter{
			Delete: aim.DeleteMcpProxy,
		},
	})

	registry.RegisterResource(resources.Command{
		Module: aimModule,
		Kind:   *resources.NewKind[apimsdk.AimCatalogMcpServer]("mcp-servers"),
		Getter: resources.Getter{
			Get:     aim.GetMcpServer,
			Columns: aim.McpServerColumns,
		},
		Applier: resources.Applier{
			Put:    aim.PutMcpServer,
			DryRun: aim.DryRunMcpServer,
		},
		Deleter: resources.Deleter{
			Delete: aim.DeleteMcpServer,
		},
	})

	registry.HookCommands(main)
}
