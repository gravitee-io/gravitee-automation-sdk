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
	"strings"

	"gopkg.in/yaml.v3"
)

// Document is a parsed OpenAPI document, resolving component schemas on demand.
type Document struct {
	schemas map[string]any
}

// ParseOpenAPI parses an OpenAPI 3.x document (YAML or JSON).
func ParseOpenAPI(raw []byte) (*Document, error) {
	var doc struct {
		Components struct {
			Schemas map[string]any `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse OpenAPI document: %w", err)
	}
	if doc.Components.Schemas == nil {
		return nil, fmt.Errorf("parse OpenAPI document: no components.schemas")
	}
	return &Document{schemas: doc.Components.Schemas}, nil
}

// Schema resolves a component schema by name, following $ref and flattening allOf.
func (d *Document) Schema(name string) (*Schema, error) {
	node, ok := d.schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema %q is not in components.schemas", name)
	}
	return d.resolve(node, map[string]bool{name: true}), nil
}

func (d *Document) resolve(node any, visiting map[string]bool) *Schema {
	m, ok := node.(map[string]any)
	if !ok {
		return &Schema{}
	}
	if ref, ok := m["$ref"].(string); ok {
		name := strings.TrimPrefix(ref, "#/components/schemas/")
		if visiting[name] {
			return &Schema{} // recursive schema: the mapper treats it as untyped below this point
		}
		target, ok := d.schemas[name]
		if !ok {
			return &Schema{}
		}
		visiting[name] = true
		s := d.resolve(target, visiting)
		delete(visiting, name)
		if ro, _ := m["readOnly"].(bool); ro {
			s.ReadOnly = true
		}
		return s
	}
	s := &Schema{Type: typeOf(m)}
	if ro, _ := m["readOnly"].(bool); ro {
		s.ReadOnly = true
	}
	if props, ok := m["properties"].(map[string]any); ok {
		s.Properties = map[string]*Schema{}
		for k, v := range props {
			s.Properties[k] = d.resolve(v, visiting)
		}
		if s.Type == "" {
			s.Type = "object"
		}
	}
	if items, ok := m["items"]; ok {
		s.ItemsSchema = d.resolve(items, visiting)
		if s.Type == "" {
			s.Type = "array"
		}
	}
	if ap, ok := m["additionalProperties"].(map[string]any); ok {
		s.AdditionalProperties = d.resolve(ap, visiting)
	}
	// allOf: the state schemas are allOf[Spec, Status]; merge every part into one object.
	if parts, ok := m["allOf"].([]any); ok {
		for _, part := range parts {
			merge(s, d.resolve(part, visiting))
		}
	}
	// oneOf/anyOf (discriminated unions such as listeners): keep the union of properties so
	// known fields pass through typed, which is what the CRD side does with preserve-unknown.
	for _, key := range []string{"oneOf", "anyOf"} {
		if parts, ok := m[key].([]any); ok {
			for _, part := range parts {
				merge(s, d.resolve(part, visiting))
			}
		}
	}
	return s
}

func merge(into, from *Schema) {
	if from == nil {
		return
	}
	if into.Type == "" {
		into.Type = from.Type
	}
	if len(from.Properties) > 0 && into.Properties == nil {
		into.Properties = map[string]*Schema{}
	}
	for k, v := range from.Properties {
		if _, exists := into.Properties[k]; !exists {
			into.Properties[k] = v
		}
	}
	if into.ItemsSchema == nil {
		into.ItemsSchema = from.ItemsSchema
	}
	if into.AdditionalProperties == nil {
		into.AdditionalProperties = from.AdditionalProperties
	}
	into.ReadOnly = into.ReadOnly || from.ReadOnly
}

// typeOf reads `type`, which OpenAPI 3.1 may write as a list (["string", "null"]).
func typeOf(m map[string]any) string {
	switch t := m["type"].(type) {
	case string:
		return t
	case []any:
		for _, v := range t {
			if s, ok := v.(string); ok && s != "null" {
				return s
			}
		}
	}
	return ""
}
