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

package authz

import (
	"context"

	apimsdk "github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	res "github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim/internal"
)

func PolicyColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Hrid"),
		res.NewColumnFromField("EntityId"),
		res.NewColumnFromField("PolicyStatus"),
	}
}

func GetPolicy(ctx context.Context, _ core.Module, _ res.Kind, id string) (any, error) {
	return res.HandleResponse(internal.Client(ctx).GetAuthzPolicyWithResponse(ctx, id))
}

func PutPolicy(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	_, err := putPolicy(ctx, false, kind)
	return err
}

func DryRunPolicy(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	resp, err := putPolicy(ctx, true, kind)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToAuthzErrorList(resp.JSON200.Errors)
}

func DeletePolicy(ctx context.Context, _ core.Module, _ res.Kind, id string) error {
	response, err := internal.Client(ctx).DeleteAuthzPolicyWithResponse(ctx, id, nil)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func putPolicy(ctx context.Context, dryRun bool, kind res.Kind) (*apimsdk.UpsertAuthzPolicyResponse, error) {
	body := *kind.GetObject().(*apimsdk.AuthzPolicy)
	params := &apimsdk.UpsertAuthzPolicyParams{DryRun: &dryRun}
	response, err := internal.Client(ctx).UpsertAuthzPolicyWithResponse(ctx, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}

// DryRunDeletePolicy asks AuthZ whether the policy can be deleted: it answers the
// state with its findings, or the error the delete would return.
func DryRunDeletePolicy(ctx context.Context, _ core.Module, _ res.Kind, id string) error {
	params := &apimsdk.DeleteAuthzPolicyParams{DryRun: new(true)}
	response, err := internal.Client(ctx).DeleteAuthzPolicyWithResponse(ctx, id, params)
	if err != nil {
		return err
	}
	if err := core.ExpectSuccess(response.StatusCode(), response.Body); err != nil {
		return err
	}
	if response.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToAuthzErrorList(response.JSON200.Errors)
}
