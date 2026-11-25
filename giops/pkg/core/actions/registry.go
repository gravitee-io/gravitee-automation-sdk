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

// Package actions builds the module commands that are not about one resource, such as "giops am check".
package actions

import (
	"github.com/spf13/cobra"
)

// Registry holds the actions of one module, keyed by verb, until HookActionCommands adds them.
type Registry struct {
	actions map[string]Command
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	reg := &Registry{
		actions: make(map[string]Command),
	}

	return reg
}

// RegisterAction adds cmd as "<module> <verb>". A verb registered twice is a bug, so it panics.
func (r *Registry) RegisterAction(verb string, cmd Command) {
	if _, exists := r.actions[verb]; exists {
		panic("action already registered " + verb)
	}
	r.actions[verb] = cmd
}

// HookActionCommands adds every registered action under main, with its Options.
func (r *Registry) HookActionCommands(main *cobra.Command) {
	for verb, command := range r.actions {
		cmd := command.add(main, verb)
		cmd.RunE = command.execute
		if command.Options != nil {
			command.Options.Flag(cmd.Flags())
		}
	}
}
