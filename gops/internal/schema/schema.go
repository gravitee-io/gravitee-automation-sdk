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

// Package schema holds the one structural model gops needs from both the Automation API OpenAPI
// document and a GKO CustomResourceDefinition: property names, types, read-only flags and the
// container shapes (array vs map). Both documents are JSON Schema dialects, so one struct covers
// them and the CRD mapper compares the two sides directly.
package schema

// Schema is the subset of JSON Schema the mapper reads.
type Schema struct {
	Type                 string
	Properties           map[string]*Schema
	ItemsSchema          *Schema
	AdditionalProperties *Schema
	ReadOnly             bool
	// PreserveUnknown marks a CRD node with x-kubernetes-preserve-unknown-fields: anything goes under it.
	PreserveUnknown bool
}

// Property returns the schema of a property, or nil when the schema does not declare it.
func (s *Schema) Property(name string) *Schema {
	if s == nil || s.Properties == nil {
		return nil
	}
	return s.Properties[name]
}

// IsMap reports whether the schema is an object whose keys are free (a Kubernetes map).
func (s *Schema) IsMap() bool {
	return s != nil && s.Type == "object" && s.AdditionalProperties != nil && len(s.Properties) == 0
}

// IsArray reports whether the schema is an array.
func (s *Schema) IsArray() bool {
	return s != nil && s.Type == "array"
}

// Typed reports whether the schema constrains its properties: an untyped node accepts anything.
func (s *Schema) Typed() bool {
	return s != nil && !s.PreserveUnknown && (len(s.Properties) > 0 || s.AdditionalProperties != nil || s.ItemsSchema != nil)
}

// Items returns the item schema of an array, or nil.
func (s *Schema) Items() *Schema {
	if s == nil {
		return nil
	}
	return s.ItemsSchema
}

// AdditionalProps returns the value schema of a map, or nil.
func (s *Schema) AdditionalProps() *Schema {
	if s == nil {
		return nil
	}
	return s.AdditionalProperties
}
