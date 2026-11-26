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

// Certificate is the file format of an AM certificate: the certificate itself,
// plus the key of the domain it belongs to, which the API takes in the path.
type Certificate struct {
	Domain string `json:"domain"`
	sdk.Certificate
}

func CertificateDomain(object store.Identifiable) string {
	return object.(*Certificate).Domain
}

func CertificateColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Key"),
		res.NewColumnFromField("Type"),
		res.NewColumnFromField("ExpiresAt"),
	}
}

func ListCertificate(ctx context.Context, _ core.Module, kind res.Kind) ([]any, error) {
	domain, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	return res.HandleResponseSlice(internal.Client(ctx).ListCertificatesWithResponse(ctx, domain))
}

func GetCertificate(ctx context.Context, _ core.Module, kind res.Kind, id string) (any, error) {
	domain, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	return res.HandleResponse(internal.Client(ctx).GetCertificateWithResponse(ctx, domain, id))
}

func DryRunCertificate(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	resp, err := putCertificate(ctx, true, kind)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToErrorList(resp.JSON200.DryRunErrors)
}

func PutCertificate(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	_, err := putCertificate(ctx, false, kind)
	return err
}

func DeleteCertificate(ctx context.Context, _ core.Module, kind res.Kind, id string) error {
	domain, err := kind.GetParent()
	if err != nil {
		return err
	}
	response, err := internal.Client(ctx).DeleteCertificateWithResponse(ctx, domain, id)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func putCertificate(ctx context.Context, dryRun bool, kind res.Kind) (*sdk.UpsertCertificateResponse, error) {
	domain, err := kind.GetParent()
	if err != nil {
		return nil, err
	}
	body := kind.GetObject().(*Certificate).Certificate
	params := &sdk.UpsertCertificateParams{DryRun: &dryRun}
	response, err := internal.Client(ctx).UpsertCertificateWithResponse(ctx, domain, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}
