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

package internal

import (
	"strings"

	"github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
)

// ToErrorList turns the dryRunErrors of an AM dry run into what a DryRun returns: nil when there is none,
// else a core.DynRunResult with each WARNING and ERROR message. Other severities are left out.
func ToErrorList(runError []sdk.DryRunError) error {
	if len(runError) == 0 {
		return nil
	}

	err := core.DynRunResult{}
	err.Warnings = []string{}
	err.Errors = []string{}
	for _, e := range runError {
		if e.Severity != nil && e.Message != nil {
			if strings.ToUpper(string(*e.Severity)) == "WARNING" {
				err.Warnings = append(err.Warnings, *e.Message)
			}
			if strings.ToUpper(string(*e.Severity)) == "ERROR" {
				err.Errors = append(err.Errors, *e.Message)
			}
		}
	}
	return err
}
