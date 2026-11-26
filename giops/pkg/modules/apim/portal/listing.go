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

// Listing is the file format of a listing of a portal: the listing itself, plus the hrid of
// the portal, which the API takes in the path.
type Listing struct {
	Portal string `json:"portal"`
	apimsdk.PortalListing
}

func ListingPortal(object store.Identifiable) string {
	return object.(*Listing).Portal
}

func ListingColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Hrid"),
		res.NewColumnFromField("PortalHrid"),
	}
}

func GetListing(ctx context.Context, _ core.Module, kind res.Kind, id string) (any, error) {
	portal, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	return res.HandleResponse(internal.Client(ctx).GetPortalListingWithResponse(ctx, portal, id))
}

func PutListing(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	_, err := putListing(ctx, false, kind)
	return err
}

func DryRunListing(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	resp, err := putListing(ctx, true, kind)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToErrorList(resp.JSON200.Errors)
}

func DeleteListing(ctx context.Context, _ core.Module, kind res.Kind, id string) error {
	portal, err := kind.GetParent()
	if err != nil {
		return err
	}
	response, err := internal.Client(ctx).DeletePortalListingWithResponse(ctx, portal, id)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func putListing(ctx context.Context, dryRun bool, kind res.Kind) (*apimsdk.UpsertPortalListingResponse, error) {
	portal, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	body := kind.GetObject().(*Listing).PortalListing
	params := &apimsdk.UpsertPortalListingParams{DryRun: &dryRun}
	response, err := internal.Client(ctx).UpsertPortalListingWithResponse(ctx, portal, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}
