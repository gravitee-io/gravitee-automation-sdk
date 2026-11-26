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

// ClientSecret is the file format of an AM client secret: the secret itself, plus the keys of the application
// and the domain it belongs to, which the API takes in the path.
type ClientSecret struct {
	Domain      string `json:"domain"`
	Application string `json:"application"`
	sdk.ClientSecret
}

func ClientSecretDomain(object store.Identifiable) string {
	return object.(*ClientSecret).Domain
}

func ClientSecretApplication(object store.Identifiable) string {
	return object.(*ClientSecret).Application
}

func ClientSecretColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Key"),
		res.NewColumnFromField("ExpiresAt"),
	}
}

func ListClientSecret(ctx context.Context, _ core.Module, kind res.Kind) ([]any, error) {
	domain, application, err := secretParents(kind)
	if err != nil {
		return nil, err
	}
	return res.HandleResponseSlice(internal.Client(ctx).ListClientSecretsWithResponse(ctx, domain, application))
}

func GetClientSecret(ctx context.Context, _ core.Module, kind res.Kind, id string) (any, error) {
	domain, application, err := secretParents(kind)
	if err != nil {
		return nil, err
	}
	return res.HandleResponse(internal.Client(ctx).GetClientSecretWithResponse(ctx, domain, application, id))
}

func DryRunClientSecret(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	resp, err := putClientSecret(ctx, true, kind)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToErrorList(resp.JSON200.DryRunErrors)
}

func PutClientSecret(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	_, err := putClientSecret(ctx, false, kind)
	return err
}

func DeleteClientSecret(ctx context.Context, _ core.Module, kind res.Kind, id string) error {
	domain, application, err := secretParents(kind)
	if err != nil {
		return err
	}
	response, err := internal.Client(ctx).DeleteClientSecretWithResponse(ctx, domain, application, id)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func putClientSecret(ctx context.Context, dryRun bool, kind res.Kind) (*sdk.UpsertClientSecretResponse, error) {
	domain, application, err := secretParents(kind)
	if err != nil {
		return nil, err
	}
	body := kind.GetObject().(*ClientSecret).ClientSecret
	params := &sdk.UpsertClientSecretParams{DryRun: &dryRun}
	response, err := internal.Client(ctx).UpsertClientSecretWithResponse(ctx, domain, application, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}

// secretParents returns the domain and application of the secret, read in the order the Kind declares them.
func secretParents(kind res.Kind) (domain, application string, err error) {
	if application, err = kind.GetParent(); err != nil {
		return "", "", err
	}
	domain, err = kind.GetParent()
	return domain, application, err
}
