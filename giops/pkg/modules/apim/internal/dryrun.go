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
	"github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
)

// ToErrorList turns the errors of an APIM dry run into what a DryRun returns: nil when there is none, else a
// core.DynRunResult with the warning and severe findings.
func ToErrorList(errors *sdk.Errors) error {

	if errors == nil {
		return nil
	}

	if len(errors.Warning) == 0 && len(errors.Severe) == 0 {
		return nil
	}

	err := core.DynRunResult{}
	err.Warnings = make([]string, len(errors.Warning))
	err.Errors = make([]string, len(errors.Severe))
	copy(err.Warnings, errors.Warning)
	copy(err.Errors, errors.Severe)

	return err
}

// ToAimErrorList does what ToErrorList does for the findings of an AIM dry run.
func ToAimErrorList(errors *sdk.AimErrors) error {
	if errors == nil {
		return nil
	}
	return ToErrorList(&sdk.Errors{Severe: errors.Severe, Warning: errors.Warning})
}

// ToAuthzErrorList does what ToErrorList does for the findings of an AuthZ dry run.
func ToAuthzErrorList(errors *sdk.AuthzErrors) error {
	if errors == nil {
		return nil
	}
	return ToErrorList(&sdk.Errors{Severe: errors.Severe, Warning: errors.Warning})
}
