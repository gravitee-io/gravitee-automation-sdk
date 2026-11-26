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
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/store"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	res "github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/am/internal"
)

// Application is the file format of an AM application: the application itself,
// plus the key of the domain it belongs to, which the API takes in the path.
type Application struct {
	Domain string `json:"domain"`
	sdk.Application
}

func ApplicationDomain(object store.Identifiable) string {
	return object.(*Application).Domain
}

func ApplicationColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Key"),
		res.NewColumnFromField("Type"),
		res.NewColumnFromField("ClientId"),
		res.NewColumnFromField("Enabled"),
	}
}

func ListApplication(ctx context.Context, _ core.Module, kind res.Kind) ([]any, error) {
	domain, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	return res.HandleResponseSlice(internal.Client(ctx).ListApplicationsWithResponse(ctx, domain))
}

func GetApplication(ctx context.Context, _ core.Module, kind res.Kind, id string) (any, error) {
	domain, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	return res.HandleResponse(internal.Client(ctx).GetApplicationWithResponse(ctx, domain, id))
}

func DryRunApplication(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	resp, err := putApplication(ctx, true, kind)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToErrorList(resp.JSON200.DryRunErrors)
}

func PutApplication(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	_, err := putApplication(ctx, false, kind)
	return err
}

func DeleteApplication(ctx context.Context, _ core.Module, kind res.Kind, id string) error {
	domain, err := kind.GetParent()
	if err != nil {
		return err
	}
	response, err := internal.Client(ctx).DeleteApplicationWithResponse(ctx, domain, id)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func putApplication(ctx context.Context, dryRun bool, kind res.Kind) (*sdk.UpsertApplicationResponse, error) {
	domain, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	body := kind.GetObject().(*Application).Application
	params := &sdk.UpsertApplicationParams{DryRun: &dryRun}
	response, err := internal.Client(ctx).UpsertApplicationWithResponse(ctx, domain, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}
