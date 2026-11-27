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

package resources

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Output is the format get prints, the value of its -o flag. It implements pflag.Value.
type Output string

// The formats. None prints nothing; Default, the value without -o, prints a table.
const (
	None       Output = ""
	Default    Output = "default"
	YamlOutput Output = "yaml"
	JsonOutput Output = "json"
)

// String returns the format, as pflag.Value requires.
func (o *Output) String() string { return string(*o) }

// Set sets the format from the flag value, and refuses one that is not a format.
func (o *Output) Set(v string) error {
	switch Output(v) {
	case JsonOutput, YamlOutput, Default, None:
		*o = Output(v)
		return nil
	}
	return fmt.Errorf(`must be one of "json" or "yaml" when set, got %q`, v)
}

// Type names the flag value type in the help, as pflag.Value requires.
func (o *Output) Type() string { return "output" }

// formatList prints resources in the format: a table of columns, or one YAML or JSON list.
func (o *Output) formatList(w io.Writer, resources []any, columns []*Column) error {
	switch *o {
	case None:
		return nil
	case YamlOutput, JsonOutput:
		return o.encode(w, resources)
	}
	return printTable(w, resources, columns)
}

// Format prints resource in the format: a one-row table of columns, or a YAML or JSON document. YAML and
// JSON use the API field names, so the output can be applied back.
func (o *Output) Format(w io.Writer, resource any, columns []*Column) error {
	// no nil check, cobra sets it everytime
	if *o == YamlOutput || *o == JsonOutput {
		return o.encode(w, resource)
	}
	return printTable(w, []any{resource}, columns)
}

// encode writes v as indented JSON, or as YAML with the JSON field names.
func (o *Output) encode(w io.Writer, v any) error {
	if *o == JsonOutput {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(v)
	}
	// through JSON, so YAML keys are the JSON field names and empty fields are left out
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var tree any
	if err := json.Unmarshal(b, &tree); err != nil {
		return err
	}
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	return encoder.Encode(tree)
}
