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

package am

import (
	"fmt"

	amsdk "github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	res "github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/spf13/cobra"
)

var ProbeOptions = &core.Option{
	Name:      "print-result",
	Shorthand: "p",
	Usage:     "output the server response as JSON",
	Default:   false,
}

func Probe(cmd *cobra.Command, module core.Module, _ []string) error {
	domains, err := ListDomain(cmd.Context(), module, *res.NewKind[amsdk.Domain]("domains"))
	if err != nil {
		return fmt.Errorf("cannot access API: %w", err)
	}
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "module %s is accessible\n", module.Name); err != nil {
		return err
	}
	if ProbeOptions.GetValueAsBool() {
		return new(res.JsonOutput).Format(out, domains, nil)
	}
	return nil
}
