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
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/store"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	res "github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim/internal"
)

// ApiLink is the file format of a link of an API: the link itself, plus the hrid of
// the API, which the API takes in the path.
type ApiLink struct {
	Api string `json:"api"`
	apimsdk.PortalLink
}

func ApiLinkApi(object store.Identifiable) string {
	return object.(*ApiLink).Api
}

func ApiLinkColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Hrid"),
		res.NewColumnFromField("ApiHrid"),
	}
}

func GetApiLink(ctx context.Context, _ core.Module, kind res.Kind, id string) (any, error) {
	api, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	return res.HandleResponse(internal.Client(ctx).GetApiLinkWithResponse(ctx, api, id))
}

func PutApiLink(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	_, err := putApiLink(ctx, false, kind)
	return err
}

func DryRunApiLink(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	resp, err := putApiLink(ctx, true, kind)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToErrorList(resp.JSON200.Errors)
}

func DeleteApiLink(ctx context.Context, _ core.Module, kind res.Kind, id string) error {
	api, err := kind.GetParent()
	if err != nil {
		return err
	}
	response, err := internal.Client(ctx).DeleteApiLinkWithResponse(ctx, api, id)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func putApiLink(ctx context.Context, dryRun bool, kind res.Kind) (*apimsdk.UpsertApiLinkResponse, error) {
	api, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	body := kind.GetObject().(*ApiLink).PortalLink
	params := &apimsdk.UpsertApiLinkParams{DryRun: &dryRun}
	response, err := internal.Client(ctx).UpsertApiLinkWithResponse(ctx, api, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}
