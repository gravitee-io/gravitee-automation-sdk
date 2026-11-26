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
	amsdk "github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/config"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/actions"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/am"
	"github.com/spf13/cobra"
)

func AM(c config.CLIConfig) (any, error) {
	ac, err := c.AMContext()
	if err != nil {
		return nil, err
	}
	return amsdk.NewAMClient(ac)
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

	registry.RegisterResource(resources.Command{
		Kind:   *resources.NewKind[am.Reporter]("reporters").WithParent(am.ReporterDomain, "domain"),
		Module: amModule,
		Getter: resources.Getter{
			List:    am.ListReporter,
			Get:     am.GetReporter,
			Columns: am.ReporterColumns,
		},
		Applier: resources.Applier{
			Put:    am.PutReporter,
			DryRun: am.DryRunReporter,
		},
		Deleter: resources.Deleter{
			Delete: am.DeleteReporter,
		},
	})

	registry.RegisterResource(resources.Command{
		Kind:   *resources.NewKind[am.Application]("applications").WithParent(am.ApplicationDomain, "domain"),
		Module: amModule,
		Getter: resources.Getter{
			List:    am.ListApplication,
			Get:     am.GetApplication,
			Columns: am.ApplicationColumns,
		},
		Applier: resources.Applier{
			Put:    am.PutApplication,
			DryRun: am.DryRunApplication,
		},
		Deleter: resources.Deleter{
			Delete: am.DeleteApplication,
		},
	})

	registry.RegisterResource(resources.Command{
		Kind: *resources.NewKind[am.ClientSecret]("secrets").
			WithParent(am.ClientSecretApplication, "application").
			WithParent(am.ClientSecretDomain, "domain"),
		Module: amModule,
		Getter: resources.Getter{
			List:    am.ListClientSecret,
			Get:     am.GetClientSecret,
			Columns: am.ClientSecretColumns,
		},
		Applier: resources.Applier{
			Put:    am.PutClientSecret,
			DryRun: am.DryRunClientSecret,
		},
		Deleter: resources.Deleter{
			Delete: am.DeleteClientSecret,
		},
	})

	registry.RegisterResource(resources.Command{
		Kind:   *resources.NewKind[am.Certificate]("certificates").WithParent(am.CertificateDomain, "domain"),
		Module: amModule,
		Getter: resources.Getter{
			List:    am.ListCertificate,
			Get:     am.GetCertificate,
			Columns: am.CertificateColumns,
		},
		Applier: resources.Applier{
			Put:    am.PutCertificate,
			DryRun: am.DryRunCertificate,
		},
		Deleter: resources.Deleter{
			Delete: am.DeleteCertificate,
		},
	})

	registry.RegisterResource(resources.Command{
		Kind:   *resources.NewKind[amsdk.DataPlane]("dataplanes"),
		Module: amModule,
		Getter: resources.Getter{
			List:    am.ListDataPlane,
			Get:     am.GetDataPlane,
			Columns: am.DataPlaneColumns,
		},
		Applier: resources.Applier{
			Put: am.PutDataPlane,
		},
		Deleter: resources.Deleter{
			Delete: am.DeleteDataPlane,
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
