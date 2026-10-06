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
	"encoding/json"

	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/resource"
	"gopkg.in/yaml.v3"
)

// rawJSON writes the state as the Automation API returned it.
type rawJSON struct{}

func (rawJSON) Format() string { return "json" }
func (rawJSON) Ext() string    { return "json" }
func (rawJSON) Transform(r resource.Resource, _ string) ([]byte, []string, error) {
	out, err := json.MarshalIndent(r.State, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return append(out, '\n'), nil, nil
}

// rawYAML writes the same state as YAML.
type rawYAML struct{}

func (rawYAML) Format() string { return "yaml" }
func (rawYAML) Ext() string    { return "yaml" }
func (rawYAML) Transform(r resource.Resource, _ string) ([]byte, []string, error) {
	out, err := yaml.Marshal(r.State)
	return out, nil, err
}
