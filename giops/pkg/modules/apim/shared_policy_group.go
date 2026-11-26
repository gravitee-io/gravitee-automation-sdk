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

package apim

import (
	"context"

	apimsdk "github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	res "github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim/internal"
)

func SharedPolicyGroupColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Hrid"),
		res.NewColumnFromField("Name"),
		res.NewColumnFromField("Description"),
		res.NewColumnFromField("ApiType"),
	}
}

func GetSharedPolicyGroup(ctx context.Context, _ core.Module, _ res.Kind, id string) (any, error) {
	return res.HandleResponse(internal.Client(ctx).GetSharedPolicyGroupWithResponse(ctx, id))
}

func PutSharedPolicyGroup(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	_, err := putSharedPolicyGroup(ctx, false, kind)
	return err
}

func DryRunSharedPolicyGroup(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	resp, err := putSharedPolicyGroup(ctx, true, kind)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToErrorList(resp.JSON200.Errors)
}

func DeleteSharedPolicyGroup(ctx context.Context, _ core.Module, _ res.Kind, id string) error {
	response, err := internal.Client(ctx).DeleteSharedPolicyGroupWithResponse(ctx, id, nil)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func putSharedPolicyGroup(ctx context.Context, dryRun bool, kind res.Kind) (*apimsdk.UpsertSharedPolicyGroupResponse, error) {
	body := *kind.GetObject().(*apimsdk.SharedPolicyGroup)
	params := &apimsdk.UpsertSharedPolicyGroupParams{DryRun: &dryRun}
	response, err := internal.Client(ctx).UpsertSharedPolicyGroupWithResponse(ctx, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}

// DryRunDeleteSharedPolicyGroup asks APIM whether the group can be deleted: the API answers 204 when it can, and
// the error the delete would return when it cannot.
func DryRunDeleteSharedPolicyGroup(ctx context.Context, _ core.Module, _ res.Kind, id string) error {
	params := &apimsdk.DeleteSharedPolicyGroupParams{DryRun: new(true)}
	response, err := internal.Client(ctx).DeleteSharedPolicyGroupWithResponse(ctx, id, params)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}
