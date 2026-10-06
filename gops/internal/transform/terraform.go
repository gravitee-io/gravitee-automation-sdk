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

package transform

import (
	"fmt"

	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/resource"
)

// terraform is the slot for the Terraform provider's resource blocks (apim_apiv4, apim_mcp_proxy).
// Registered so --format=tf fails with a clear message rather than an unknown-format error.
type terraform struct{}

func (terraform) Format() string { return "tf" }
func (terraform) Ext() string    { return "tf" }
func (terraform) Transform(r resource.Resource, _ string) ([]byte, []string, error) {
	return nil, nil, fmt.Errorf("--format=tf for %s: %w", r.Kind.Name, ErrNotImplemented)
}
