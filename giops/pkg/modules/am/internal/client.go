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

// Package internal holds what the AM resources share.
package internal

import (
	"context"

	"github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
)

// Client returns the AM client of the module command running in ctx.
func Client(ctx context.Context) *sdk.AMClient {
	return core.Client[sdk.AMClient](ctx)
}
