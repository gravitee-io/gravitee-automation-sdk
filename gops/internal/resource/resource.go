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

// Package resource lists what an environment holds, one Kind per Automation API collection,
// and hands each resource over as the tree the API returned.
package resource

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/response"
)

// Resource is one exported resource: its identity and the state the Automation API returned.
type Resource struct {
	Kind  *Kind
	ID    string
	Name  string
	State map[string]any
}

// Kind binds a --resources name to the SDK call that lists it, the OpenAPI schema of what the
// call returns and the GKO kind it maps to.
type Kind struct {
	// Name is the --resources value and the output sub-directory: "apis", "mcp-proxies".
	Name string
	// StateSchema is the components.schemas entry describing one listed item.
	StateSchema string
	// CRDKind is the GKO kind the item maps to.
	CRDKind string
	// List fetches the resources of the kind through the SDK.
	List func(ctx context.Context, client *sdk.APIMClient) ([]Resource, error)
}

// Kinds returns every kind gops can export, in display order.
func Kinds() []*Kind {
	return []*Kind{apis, mcpProxies}
}

// Lookup returns the kind registered under name.
func Lookup(name string) (*Kind, bool) {
	for _, k := range Kinds() {
		if k.Name == name {
			return k, true
		}
	}
	return nil, false
}

var apis = &Kind{
	Name:        "apis",
	StateSchema: "ApiV4State",
	CRDKind:     "ApiV4Definition",
}

var mcpProxies = &Kind{
	Name:        "mcp-proxies",
	StateSchema: "AimMcpProxyState",
	CRDKind:     "McpProxy",
}

func init() {
	apis.List = func(ctx context.Context, client *sdk.APIMClient) ([]Resource, error) {
		res, err := client.ListApisWithResponse(ctx)
		if err != nil {
			return nil, fmt.Errorf("list apis: %w", err)
		}
		states, ok := response.Payload[[]sdk.ApiV4State](res)
		if !ok {
			return nil, httpError("list apis", res.StatusCode(), res.Body)
		}
		return collect(apis, states)
	}
	mcpProxies.List = func(ctx context.Context, client *sdk.APIMClient) ([]Resource, error) {
		res, err := client.ListAimMcpProxiesWithResponse(ctx)
		if err != nil {
			return nil, fmt.Errorf("list mcp-proxies: %w", err)
		}
		states, ok := response.Payload[[]sdk.AimMcpProxyState](res)
		if !ok {
			return nil, httpError("list mcp-proxies", res.StatusCode(), res.Body)
		}
		return collect(mcpProxies, states)
	}
}

// collect turns typed SDK states into resources. The tree is the state as the SDK serializes it:
// what the generated model keeps is exactly what gops exports, which is the point of going through
// the SDK rather than raw JSON.
func collect[T any](kind *Kind, states []T) ([]Resource, error) {
	out := make([]Resource, 0, len(states))
	for _, s := range states {
		tree, err := toTree(s)
		if err != nil {
			return nil, err
		}
		id, _ := tree["id"].(string)
		name, _ := tree["name"].(string)
		out = append(out, Resource{Kind: kind, ID: id, Name: name, State: tree})
	}
	return out, nil
}

func toTree(v any) (map[string]any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var tree map[string]any
	if err := json.Unmarshal(raw, &tree); err != nil {
		return nil, err
	}
	return tree, nil
}

func httpError(op string, status int, body []byte) error {
	msg := string(body)
	if len(msg) > 300 {
		msg = msg[:300] + "…"
	}
	return fmt.Errorf("%s: HTTP %d: %s", op, status, msg)
}
