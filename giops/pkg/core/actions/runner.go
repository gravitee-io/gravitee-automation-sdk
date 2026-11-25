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

package actions

import (
	"fmt"

	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	"github.com/spf13/cobra"
)

// Command is an action of a module.
//
// Run does the action. cmd's context holds the module client (read it with core.Client), and cmd's
// output is where to print; args are the positional arguments. An error is printed after "cannot execute
// action:" and giops exits non-zero.
//
// Options adds the action's flags, nil when it has none; Run reads their values from the same value, such as
// a *core.Option.
type Command struct {
	// Module is the module the action belongs to: "giops <module> <action>".
	Module core.Module
	// Run does the action. Required.
	Run func(cmd *cobra.Command, module core.Module, args []string) error
	// Options adds the action's flags. Nil when it has none.
	Options core.CommandFlagger
}

// add finds or creates "<module> <verb>" under main.
func (c Command) add(main *cobra.Command, verb string) *cobra.Command {
	return core.FindOrCreate(c.Module.Command(main), verb)
}

// execute runs the action, as cobra's RunE.
func (c Command) execute(cmd *cobra.Command, args []string) error {
	if err := c.Run(cmd, c.Module, args); err != nil {
		return fmt.Errorf("cannot execute action: %w", err)
	}
	return nil
}
