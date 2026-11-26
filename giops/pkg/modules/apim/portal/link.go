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

package portal

import (
	"context"

	apimsdk "github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/store"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	res "github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim/internal"
)

// Link is the file format of a link of a portal: the link itself, plus the hrid of
// the portal, which the API takes in the path.
type Link struct {
	Portal string `json:"portal"`
	apimsdk.PortalLink
}

func LinkPortal(object store.Identifiable) string {
	return object.(*Link).Portal
}

func LinkColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Hrid"),
		res.NewColumnFromField("PortalHrid"),
	}
}

func GetLink(ctx context.Context, _ core.Module, kind res.Kind, id string) (any, error) {
	portal, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	return res.HandleResponse(internal.Client(ctx).GetPortalLinkWithResponse(ctx, portal, id))
}

func PutLink(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	_, err := putLink(ctx, false, kind)
	return err
}

func DryRunLink(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	resp, err := putLink(ctx, true, kind)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToErrorList(resp.JSON200.Errors)
}

func DeleteLink(ctx context.Context, _ core.Module, kind res.Kind, id string) error {
	portal, err := kind.GetParent()
	if err != nil {
		return err
	}
	response, err := internal.Client(ctx).DeletePortalLinkWithResponse(ctx, portal, id)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func putLink(ctx context.Context, dryRun bool, kind res.Kind) (*apimsdk.UpsertPortalLinkResponse, error) {
	portal, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	body := kind.GetObject().(*Link).PortalLink
	params := &apimsdk.UpsertPortalLinkParams{DryRun: &dryRun}
	response, err := internal.Client(ctx).UpsertPortalLinkWithResponse(ctx, portal, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}
