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
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	"github.com/spf13/pflag"
)

// getOptions are the flags of get.
type getOptions struct {
	Output *Output
}

// alterOptions are the flags apply and delete share: the file, and the dry run and confirmation that come
// before the change.
type alterOptions struct {
	File       core.Option
	DryRunOnly core.Option
	Force      core.Option
}

// deleteOptions are the flags of delete.
type deleteOptions struct {
	alterOptions
	ID core.Option
}

// applyOptions are the flags of apply.
type applyOptions struct {
	alterOptions
}

func newGetOptions() *getOptions {
	return &getOptions{Output: new(Default)}
}

func newApplyOptions() *applyOptions {
	return &applyOptions{
		alterOptions: newAlterOptions(),
	}
}

func newDeleteOptions() *deleteOptions {
	return &deleteOptions{
		alterOptions: newAlterOptions(),
		ID:           core.Option{Name: "id", Shorthand: "i", Usage: "id to delete", Default: ""},
	}
}

func newAlterOptions() alterOptions {
	return alterOptions{
		File:       core.Option{Name: "file", Shorthand: "f", Usage: "file to apply", Default: ""},
		DryRunOnly: core.Option{Name: "dry-run", Shorthand: "d", Usage: "dry run and don't apply", Default: false},
		Force:      core.Option{Name: "yes", Shorthand: "y", Usage: "apply without asking and without a dry run", Default: false},
	}
}

// AdditionalOptions are the apply flags of one resource, such as --ignore-members on APIM groups. A resource
// without any has a nil AdditionalOptions. Put and DryRun receive it to read the values the user gave.
type AdditionalOptions struct {
	options map[string]*core.Option
}

// NewSpecificOptions keys the options by name. Two options with one name is a bug in the resource, so it panics.
func NewSpecificOptions(options []core.Option) *AdditionalOptions {
	byName := make(map[string]*core.Option, len(options))
	for _, option := range options {
		if _, exists := byName[option.Name]; exists {
			panic("option declared twice: " + option.Name)
		}
		byName[option.Name] = &option
	}
	return &AdditionalOptions{options: byName}
}

// flag adds every option to f. A nil AdditionalOptions adds none.
func (o *AdditionalOptions) flag(f *pflag.FlagSet) {
	if o == nil {
		return
	}
	for name, option := range o.options {
		option.Flag(f)
		o.options[name] = option
	}
}

// Get returns the option of that name, holding the value of its flag. An unknown name is a bug in the
// resource, so it panics.
func (o *AdditionalOptions) Get(name string) *core.Option {
	option, ok := o.options[name]
	if !ok {
		panic("no option " + name)
	}
	return option
}

func (o *getOptions) flag(f *pflag.FlagSet) {
	f.VarP(o.Output, "output", "o", "optional output formats: yaml, json")
}

func (o *alterOptions) flag(f *pflag.FlagSet) {
	o.File.Flag(f)
	o.DryRunOnly.Flag(f)
	o.Force.Flag(f)
}

func (o *applyOptions) flag(f *pflag.FlagSet) {
	o.alterOptions.flag(f)
}

func (o *deleteOptions) flag(f *pflag.FlagSet) {
	o.alterOptions.flag(f)
	o.ID.Flag(f)
}
