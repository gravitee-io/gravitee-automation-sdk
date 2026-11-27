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

func GetGroup(ctx context.Context, _ core.Module, _ res.Kind, id string) (any, error) {
	response, err := internal.Client(ctx).GetGroupWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	if err := core.ExpectSuccess(response.StatusCode(), response.Body); err != nil {
		return nil, err
	}
	return response.JSON200, nil
}

func GroupColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Hrid"),
		res.NewColumnFromField("Name"),
	}
}

const ignoreMembers = "ignore-members"

func GroupOptions() []core.Option {
	return []core.Option{
		{
			Name:      ignoreMembers,
			Shorthand: "m",
			Usage:     "don't save members to APIM",
			Default:   false,
		},
	}
}

func PutGroup(ctx context.Context, _ core.Module, kind res.Kind, options *res.AdditionalOptions) error {
	_, err := put(ctx, false, kind, options)
	return err
}

func DryRunGroup(ctx context.Context, _ core.Module, kind res.Kind, options *res.AdditionalOptions) error {
	resp, err := put(ctx, true, kind, options)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToErrorList(resp.JSON200.Errors)
}

func DeleteGroup(ctx context.Context, _ core.Module, _ res.Kind, id string) error {
	response, err := internal.Client(ctx).DeleteGroupWithResponse(ctx, id, nil)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func put(ctx context.Context, dryRun bool, kind res.Kind, options *res.AdditionalOptions) (*apimsdk.UpsertGroupResponse, error) {
	body := *kind.GetObject().(*apimsdk.Group)
	params := &apimsdk.UpsertGroupParams{
		DryRun:        &dryRun,
		IgnoreMembers: new(options.Get(ignoreMembers).GetValueAsBool()),
	}
	response, err := internal.Client(ctx).UpsertGroupWithResponse(ctx, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}
