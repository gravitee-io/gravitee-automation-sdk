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

// Package resources builds the get, apply and delete commands of every resource a module serves, from the
// functions each resource gives in a Command.
package resources

import (
	"github.com/spf13/cobra"
)

// Registry holds the resource commands of one module, keyed by resource name, until HookCommands adds them.
type Registry struct {
	resources map[string]*Command
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	reg := &Registry{
		resources: make(map[string]*Command),
	}
	return reg
}

// RegisterResource adds cmd under its Kind name. A name registered twice is a bug, so it panics.
func (r *Registry) RegisterResource(cmd Command) {
	subject := cmd.Kind.name
	if _, exists := r.resources[subject]; exists {
		panic("resource command already registered " + subject)
	}
	cmd.getOptions = newGetOptions()
	cmd.applyOptions = newApplyOptions()
	cmd.deleteOptions = newDeleteOptions()
	r.resources[subject] = &cmd
}

// HookCommands adds "<module> get|apply|delete <resource>" under main for every registered resource, with their
// flags: the parents on all three, -o on get, -f -d -y and the AdditionalOptions on apply, -f -i on delete.
func (r *Registry) HookCommands(main *cobra.Command) {
	for subject, command := range r.resources {
		cmds := command.add(main, subject)
		for i, v := range verbs {
			switch v {
			case getVerb:
				cmd := cmds[i]
				cmd.RunE = command.runGet
				command.Kind.flag(cmd.Flags())
				command.getOptions.flag(cmd.Flags())
			case applyVerb:
				cmd := cmds[i]
				cmd.RunE = command.runApply
				command.applyOptions.flag(cmd.Flags())
				command.AdditionalOptions.flag(cmd.Flags())
				command.Kind.flag(cmd.Flags())
				cmd.MarkFlagsMutuallyExclusive(command.applyOptions.DryRunOnly.Name, command.applyOptions.Force.Name)
			case deleteVerb:
				cmd := cmds[i]
				cmd.RunE = command.runDelete
				command.Kind.flag(cmd.Flags())
				command.deleteOptions.flag(cmd.Flags())
			}
		}
	}
}
