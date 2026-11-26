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

package aim

import (
	"context"

	apimsdk "github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	res "github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core/resources"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/modules/apim/internal"
)

func McpServerColumns() []*res.Column {
	return []*res.Column{
		res.NewColumnFromField("Hrid"),
		res.NewColumnFromField("EntityId"),
		res.NewColumnFromExtractor("Endpoint", endpoint),
		res.NewColumnFromField("LastSyncedAt"),
	}
}

func GetMcpServer(ctx context.Context, _ core.Module, _ res.Kind, id string) (any, error) {
	return res.HandleResponse(internal.Client(ctx).GetAimCatalogMcpServerWithResponse(ctx, id))
}

func PutMcpServer(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	_, err := putMcpServer(ctx, false, kind)
	return err
}

func DryRunMcpServer(ctx context.Context, _ core.Module, kind res.Kind, _ *res.AdditionalOptions) error {
	resp, err := putMcpServer(ctx, true, kind)
	if err != nil {
		return err
	}
	if resp.JSON200 == nil {
		return core.ErrNoDryRunResult
	}
	return internal.ToAimErrorList(resp.JSON200.Errors)
}

func DeleteMcpServer(ctx context.Context, _ core.Module, _ res.Kind, id string) error {
	response, err := internal.Client(ctx).DeleteAimCatalogMcpServerWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return core.ExpectSuccess(response.StatusCode(), response.Body)
}

func putMcpServer(ctx context.Context, dryRun bool, kind res.Kind) (*apimsdk.UpsertAimCatalogMcpServerResponse, error) {
	body := *kind.GetObject().(*apimsdk.AimCatalogMcpServer)
	params := &apimsdk.UpsertAimCatalogMcpServerParams{DryRun: &dryRun}
	response, err := internal.Client(ctx).UpsertAimCatalogMcpServerWithResponse(ctx, params, body)
	if err != nil {
		return nil, err
	}
	return response, core.ExpectSuccess(response.StatusCode(), response.Body)
}

// endpoint reads the URL of the MCP server, a field of its connection.
func endpoint(item any) (string, bool) {
	server, ok := item.(*apimsdk.AimCatalogMcpServerState)
	if !ok || server == nil {
		return "", false
	}
	return server.Connection.Endpoint, true
}
