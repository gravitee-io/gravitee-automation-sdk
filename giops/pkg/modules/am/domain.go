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

func DomainColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Key"),
		res.NewColumnFromField("Name"),
		res.NewColumnFromField("Path"),
		res.NewColumnFromField("Enabled"),
	}
}

func GetDomain(ctx context.Context, _ core.Module, _ res.Kind, id string) (any, error) {
	return res.HandleResponse(internal.Client(ctx).GetDomainWithResponse(ctx, id))
}

func ListDomain(ctx context.Context, _ core.Module, _ res.Kind) ([]any, error) {
	return res.HandleResponseSlice(internal.Client(ctx).ListDomainsWithResponse(ctx))
}

func DryRunDomain(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	resp, err := put(ctx, true, kind)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToErrorList(resp.JSON200.DryRunErrors)
}

func PutDomain(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	_, err := put(ctx, false, kind)
	return err
}

func DeleteDomain(ctx context.Context, _ core.Module, _ res.Kind, id string) error {
	response, err := internal.Client(ctx).DeleteDomainWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func put(ctx context.Context, dryRun bool, kind res.Kind) (*sdk.UpsertDomainResponse, error) {
	body := *kind.GetObject().(*sdk.Domain)
	response, err := internal.Client(ctx).UpsertDomainWithResponse(ctx, &sdk.UpsertDomainParams{DryRun: &dryRun}, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}
