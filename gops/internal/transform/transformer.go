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

// Package transform turns the state the Automation API returned into one output format: the
// state itself (json, yaml), a GKO manifest (crd) or a Terraform resource block (tf).
package transform

import (
	"errors"
	"fmt"

	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/resource"
)

// Transformer renders one resource. name is the object name the export settled for it.
type Transformer interface {
	Format() string
	// Ext is the file extension written under <output>/<kind>/.
	Ext() string
	// Transform returns the file content and notes about anything it had to drop or reshape.
	Transform(r resource.Resource, name string) ([]byte, []string, error)
}

// ErrNotImplemented marks a format that is registered but not written yet.
var ErrNotImplemented = errors.New("not implemented")

// Formats lists the --format values, in display order.
func Formats() []string {
	return []string{"json", "yaml", "crd", "tf"}
}

// Lookup returns the transformer for a --format value.
func Lookup(format string) (Transformer, error) {
	switch format {
	case "json":
		return rawJSON{}, nil
	case "yaml":
		return rawYAML{}, nil
	case "crd":
		return newCRD()
	case "tf":
		return terraform{}, nil
	}
	return nil, fmt.Errorf("unknown format %q (expected one of %v)", format, Formats())
}
