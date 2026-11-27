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

// Package core holds what every module and command shares: the Module and its client, flags, prompts, and
// the errors of API calls and dry runs.
package core

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandFlagger adds flags to a command's flag set. *Option is one.
type CommandFlagger interface {
	Flag(flags *pflag.FlagSet)
}

// FindOrCreate returns the subcommand of super named use, adding an empty one when there is none, so that
// several registries can share "giops <module> get".
func FindOrCreate(super *cobra.Command, use string) *cobra.Command {
	for _, sub := range super.Commands() {
		if sub.Name() == use {
			return sub
		}
	}
	sub := &cobra.Command{Use: use}
	super.AddCommand(sub)
	return sub
}
