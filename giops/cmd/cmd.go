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

	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
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

	amModule := core.Module{Name: "am", ClientFactory: AM}
	registerAMResourceConfigs(amModule, main)
	registerAMActionConfigs(amModule, main)

	apimModule := core.Module{Name: "apim", ClientFactory: Gamma}
	registerAPIMResourceConfigs(apimModule, main)
	registerAPIMActionConfigs(apimModule, main)

	portalModule := core.Module{Name: "portal", ClientFactory: Gamma}
	registerPortalResourceConfigs(portalModule, main)
	registerAPIMActionConfigs(portalModule, main)

	aimModule := core.Module{Name: "aim", ClientFactory: Gamma}
	registerAIMResourceConfigs(aimModule, main)
	registerAPIMActionConfigs(aimModule, main)

	authzModule := core.Module{Name: "authz", ClientFactory: Gamma}
	registerAuthzResourceConfigs(authzModule, main)
	registerAPIMActionConfigs(authzModule, main)

	return main
}
