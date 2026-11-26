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
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim/authz"
	"github.com/spf13/cobra"
)

func registerAuthzResourceConfigs(authzModule core.Module, main *cobra.Command) {
	registry := resources.NewRegistry()

	registry.RegisterResource(resources.Command{
		Module: authzModule,
		Kind:   *resources.NewKind[apimsdk.AuthzSchema]("schemas"),
		Getter: resources.Getter{
			Get:     authz.GetSchema,
			Columns: authz.SchemaColumns,
		},
		Applier: resources.Applier{
			Put:    authz.PutSchema,
			DryRun: authz.DryRunSchema,
		},
		Deleter: resources.Deleter{
			Delete: authz.DeleteSchema,
		},
	})

	registry.RegisterResource(resources.Command{
		Module: authzModule,
		Kind:   *resources.NewKind[apimsdk.AuthzPdp]("pdps"),
		Getter: resources.Getter{
			Get:     authz.GetPdp,
			Columns: authz.PdpColumns,
		},
		Applier: resources.Applier{
			Put:    authz.PutPdp,
			DryRun: authz.DryRunPdp,
		},
		Deleter: resources.Deleter{
			Delete:       authz.DeletePdp,
			DryRunDelete: authz.DryRunDeletePdp,
		},
	})

	registry.RegisterResource(resources.Command{
		Module: authzModule,
		Kind:   *resources.NewKind[apimsdk.AuthzEntity]("entities"),
		Getter: resources.Getter{
			Get:     authz.GetEntity,
			Columns: authz.EntityColumns,
		},
		Applier: resources.Applier{
			Put:    authz.PutEntity,
			DryRun: authz.DryRunEntity,
		},
		Deleter: resources.Deleter{
			Delete:       authz.DeleteEntity,
			DryRunDelete: authz.DryRunDeleteEntity,
		},
	})

	registry.RegisterResource(resources.Command{
		Module: authzModule,
		Kind:   *resources.NewKind[apimsdk.AuthzPolicy]("policies"),
		Getter: resources.Getter{
			Get:     authz.GetPolicy,
			Columns: authz.PolicyColumns,
		},
		Applier: resources.Applier{
			Put:    authz.PutPolicy,
			DryRun: authz.DryRunPolicy,
		},
		Deleter: resources.Deleter{
			Delete:       authz.DeletePolicy,
			DryRunDelete: authz.DryRunDeletePolicy,
		},
	})

	registry.HookCommands(main)
}
