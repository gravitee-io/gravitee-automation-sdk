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

// Package openapi embeds the merged APIM Automation OpenAPI document the SDK is generated from,
// so tools built on the SDK (schema-driven exporters, validators) read the same contract.
package openapi

import _ "embed"

//go:embed openapi.yaml
var spec []byte

// Spec returns the merged OpenAPI document (APIM Automation API + AI Management fragment) as YAML.
func Spec() []byte {
	return spec
}
