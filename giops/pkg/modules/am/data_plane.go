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
	"context"

	"github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	res "github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/am/internal"
)

// Data planes have no dry run in the API: they register no DryRun.

func DataPlaneColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Id"),
		res.NewColumnFromField("Name"),
		res.NewColumnFromField("Type"),
	}
}

func ListDataPlane(ctx context.Context, _ core.Module, _ res.Kind) ([]any, error) {
	return res.HandleResponseSlice(internal.Client(ctx).ListDataPlanesWithResponse(ctx))
}

func GetDataPlane(ctx context.Context, _ core.Module, _ res.Kind, id string) (any, error) {
	return res.HandleResponse(internal.Client(ctx).GetDataPlaneWithResponse(ctx, id))
}

func PutDataPlane(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	body := *kind.GetObject().(*sdk.DataPlane)
	response, err := internal.Client(ctx).UpsertDataPlaneWithResponse(ctx, body)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func DeleteDataPlane(ctx context.Context, _ core.Module, _ res.Kind, id string) error {
	response, err := internal.Client(ctx).DeleteDataPlaneWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}
