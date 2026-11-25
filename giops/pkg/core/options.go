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

	"github.com/spf13/pflag"
)

// Or returns opt when it is set, else args[index] when it exists and is set, else an error with message.
// It lets a value come from a flag or from a positional argument.
func Or(opt string, args []string, index int, message string) (string, error) {
	if opt == "" {
		if len(args) > index && args[index] != "" {
			return args[index], nil
		}
		return "", errors.New(message)
	}
	return opt, nil
}

// Option is a command line flag. The type of Default, string, bool or int, is the type of the flag.
// Until a flag is added, its value is Default.
type Option struct {
	// Name is the long flag, without "--".
	Name string
	// Shorthand is the one-letter flag, without "-". Empty for none.
	Shorthand string
	// Usage is the flag's help text.
	Usage string
	// Default is the value before the flag is set: a string, bool or int, which sets the flag's type.
	Default   any
	recipient any // *string, *bool or *int, where the flag writes its value
}

// GetValueAsBool returns the value of a bool option. It panics on another type.
func (o *Option) GetValueAsBool() bool {
	return value[bool](o)
}

// GetValueAsString returns the value of a string option. It panics on another type.
func (o *Option) GetValueAsString() string {
	return value[string](o)
}

// GetValueAsInt returns the value of an int option. It panics on another type.
func (o *Option) GetValueAsInt() int {
	return value[int](o)
}

// value returns the flag value, or Default before the option is added to a flag set.
func value[T any](o *Option) T {
	if o.recipient == nil {
		return o.Default.(T)
	}
	r, ok := o.recipient.(*T)
	if !ok {
		panic(fmt.Sprintf("option --%s is a %T flag, not a %T", o.Name, o.Default, *new(T)))
	}
	return *r
}

// Flag adds the option to f. The commands that add the same option share its value,
// like get, apply and delete with the parent of a resource. It panics when Default is not a string, bool or int.
func (o *Option) Flag(f *pflag.FlagSet) {
	switch value := o.Default.(type) {
	case string:
		f.StringVarP(recipient[string](o), o.Name, o.Shorthand, value, o.Usage)
	case bool:
		f.BoolVarP(recipient[bool](o), o.Name, o.Shorthand, value, o.Usage)
	case int:
		f.IntVarP(recipient[int](o), o.Name, o.Shorthand, value, o.Usage)
	default:
		panic(fmt.Sprintf("option %s: unsupported default type %T", o.Name, o.Default))
	}
}

// recipient returns where the flag writes its value, creating it on the first call.
func recipient[T any](o *Option) *T {
	if r, ok := o.recipient.(*T); ok {
		return r
	}
	r := new(T)
	o.recipient = r
	return r
}
