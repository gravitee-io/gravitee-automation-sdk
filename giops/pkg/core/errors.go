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

package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/labstack/gommon/color"
)

var (
	warn = color.Yellow("Warning", color.B)
	err  = color.Red("Error", color.B)
)

// ErrNoDryRunResult is a successful dry run response without the JSON body that carries the result.
var ErrNoDryRunResult = errors.New("the dry run response has no JSON body")

// ExpectSuccess turns a response status outside 2xx into an error carrying the status and the body.
func ExpectSuccess(status int, body []byte) error {
	if status >= 200 && status < 300 {
		return nil
	}
	return fmt.Errorf("HTTP %d: %s", status, strings.TrimSpace(string(body)))
}

// DynRunResult is the findings of a dry run, returned as an error by a resource's DryRun. Warnings let the
// apply go on; Errors stop it.
type DynRunResult struct {
	// Warnings are findings that let the change go on, after giops prints them.
	Warnings []string
	// Errors are findings that stop the change.
	Errors []string
}

// Error lists the warnings, then the errors, one per line.
func (r DynRunResult) Error() string {
	s := strings.Builder{}
	// print the warning in yellow
	for _, w := range r.Warnings {
		s.WriteString(fmt.Sprintf("%s: %s\n", warn, w))
	}
	for _, e := range r.Errors {
		s.WriteString(fmt.Sprintf("%s: %s\n", err, e))
	}
	return s.String()
}

// HasErrors tells whether the findings stop the apply.
func (r DynRunResult) HasErrors() bool {
	return len(r.Errors) > 0
}
