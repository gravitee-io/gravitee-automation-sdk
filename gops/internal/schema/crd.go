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

package schema

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// CRD is what the mapper needs from a CustomResourceDefinition: its identity and the spec schema.
type CRD struct {
	Group   string
	Version string
	Kind    string
	Spec    *Schema
}

// APIVersion returns the apiVersion a manifest of this CRD carries.
func (c *CRD) APIVersion() string {
	return c.Group + "/" + c.Version
}

// ParseCRD parses a CRD manifest and keeps the first served version's spec schema.
func ParseCRD(raw []byte) (*CRD, error) {
	var crd struct {
		Spec struct {
			Group string `yaml:"group"`
			Names struct {
				Kind string `yaml:"kind"`
			} `yaml:"names"`
			Versions []struct {
				Name   string `yaml:"name"`
				Schema struct {
					OpenAPIV3Schema map[string]any `yaml:"openAPIV3Schema"`
				} `yaml:"schema"`
			} `yaml:"versions"`
		} `yaml:"spec"`
	}
	if err := yaml.Unmarshal(raw, &crd); err != nil {
		return nil, fmt.Errorf("parse CRD: %w", err)
	}
	if len(crd.Spec.Versions) == 0 {
		return nil, fmt.Errorf("parse CRD %s: no versions", crd.Spec.Names.Kind)
	}
	v := crd.Spec.Versions[0]
	root := fromStructural(v.Schema.OpenAPIV3Schema)
	spec := root.Property("spec")
	if spec == nil {
		return nil, fmt.Errorf("parse CRD %s: no spec schema", crd.Spec.Names.Kind)
	}
	return &CRD{Group: crd.Spec.Group, Version: v.Name, Kind: crd.Spec.Names.Kind, Spec: spec}, nil
}

// fromStructural converts a structural (CRD) schema: no $ref, no allOf, Kubernetes extensions.
func fromStructural(m map[string]any) *Schema {
	s := &Schema{Type: typeOf(m)}
	if p, _ := m["x-kubernetes-preserve-unknown-fields"].(bool); p {
		s.PreserveUnknown = true
	}
	if props, ok := m["properties"].(map[string]any); ok {
		s.Properties = map[string]*Schema{}
		for k, v := range props {
			if vm, ok := v.(map[string]any); ok {
				s.Properties[k] = fromStructural(vm)
			}
		}
	}
	if items, ok := m["items"].(map[string]any); ok {
		s.ItemsSchema = fromStructural(items)
	}
	if ap, ok := m["additionalProperties"].(map[string]any); ok {
		s.AdditionalProperties = fromStructural(ap)
	}
	return s
}
