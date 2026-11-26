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

const ignoreGroups = "ignore-groups"

// ignoreGroupsOptions is the apply flag of the resources whose groups APIM can leave as they are.
func ignoreGroupsOptions() []core.Option {
	return []core.Option{
		{
			Name:      ignoreGroups,
			Shorthand: "g",
			Usage:     "don't save groups to APIM",
			Default:   false,
		},
	}
}

func ApiOptions() []core.Option {
	return ignoreGroupsOptions()
}

func ApiColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Hrid"),
		res.NewColumnFromField("State"),
		res.NewColumnFromField("LifecycleState"),
		res.NewColumnFromField("Version"),
	}
}

func GetApi(ctx context.Context, _ core.Module, _ res.Kind, id string) (any, error) {
	return res.HandleResponse(internal.Client(ctx).GetApiWithResponse(ctx, id, nil))
}

func PutApi(ctx context.Context, _ core.Module, kind res.Kind, options *res.AdditionalOptions) error {
	_, err := putApi(ctx, false, kind, options)
	return err
}

func DryRunApi(ctx context.Context, _ core.Module, kind res.Kind, options *res.AdditionalOptions) error {
	resp, err := putApi(ctx, true, kind, options)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToErrorList(resp.JSON200.Errors)
}

func DeleteApi(ctx context.Context, _ core.Module, _ res.Kind, id string) error {
	response, err := internal.Client(ctx).DeleteApiWithResponse(ctx, id, nil)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func putApi(ctx context.Context, dryRun bool, kind res.Kind, options *res.AdditionalOptions) (*apimsdk.UpsertApiResponse, error) {
	body := *kind.GetObject().(*apimsdk.ApiV4)
	params := &apimsdk.UpsertApiParams{
		DryRun:       &dryRun,
		IgnoreGroups: new(options.Get(ignoreGroups).GetValueAsBool()),
	}
	response, err := internal.Client(ctx).UpsertApiWithResponse(ctx, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}
