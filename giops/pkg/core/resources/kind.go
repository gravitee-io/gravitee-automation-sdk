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
	"fmt"

	"github.com/gravitee-io/gravitee-automation-sdk/common/pkg/store"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	"github.com/spf13/pflag"
)

// IdentifiableExtractor reads a value from a decoded file: Kind gives it the object NewKind created, so it can
// assert the object to *T. It returns "" when the file does not hold the value.
type IdentifiableExtractor func(store.Identifiable) string

// Kind is the type of a resource: its name on the command line (the API endpoint's plural, "domains"), the
// object a file decodes into, and its parents, the resources it lives under.
//
// The resource functions of a Command receive the Kind by value: each call has its own copy, so the parents
// GetParent consumes on one call are all there on the next.
type Kind struct {
	name    string
	object  store.Identifiable
	parents []parent
}

type parent struct {
	extractor IdentifiableExtractor
	option    *core.Option
}

type pointerOf[T any] interface {
	*T
	store.Identifiable
}

// NewKind returns the Kind named name, whose files decode into a new T. Its pointer *T must be Identifiable:
// the Identity of the decoded file is the id apply reports and delete --file deletes.
func NewKind[T any, PT pointerOf[T]](name string) *Kind {
	return &Kind{
		name:    name,
		object:  PT(new(T)),
		parents: make([]parent, 0),
	}
}

// flag adds one string flag per parent to f.
func (k *Kind) flag(f *pflag.FlagSet) {
	for _, ancestor := range k.parents {
		ancestor.option.Flag(f)
	}
}

// GetObject returns the object files decode into, a *T for NewKind[T]. In apply, and in delete --file, it holds
// the file; otherwise it is the zero T.
func (k *Kind) GetObject() store.Identifiable {
	return k.object
}

// WithParent adds a parent to the Kind: a --<parentName> flag on get, apply and delete, and the extractor that
// reads the same value from a file. Call it once per parent, from the direct parent up to the root; it returns
// k, so the calls chain (a client secret: WithParent(application).WithParent(domain)). It panics on an empty
// name or a nil extractor.
func (k *Kind) WithParent(extractor IdentifiableExtractor, parentName string) *Kind {
	k.parents = append(k.parents, k.newParent(extractor, parentName))
	return k
}

// GetParent returns the id of the next parent, in the order WithParent added them: its flag when set, else the
// value its extractor reads from GetObject. It fails when neither gives one. Each call moves to the next parent
// of this copy of the Kind; it panics when there is none left.
func (k *Kind) GetParent() (string, error) {
	ancestor := k.peekParent()
	if value := ancestor.option.GetValueAsString(); value != "" {
		return value, nil
	}
	value := ancestor.extractor(k.object)
	if value == "" {
		name := ancestor.option.Name
		return "", fmt.Errorf("%s identifier required for resource %s: file should contain a '%s' property or provide it via the --%s option", name, k.name, name, name)
	}
	return value, nil
}

func (k *Kind) newParent(extractor IdentifiableExtractor, parentName string) parent {
	k.mustHaveValues(extractor, parentName)
	return parent{
		extractor: extractor,
		option:    &core.Option{Name: parentName, Usage: fmt.Sprintf("sets the %s %s id", k.name, parentName), Default: ""},
	}
}

func (k *Kind) mustHaveValues(extractor IdentifiableExtractor, parentName string) {
	if parentName == "" {
		panic(fmt.Sprintf("parent name for %s cannot be empty", k.name))
	}
	if extractor == nil {
		panic(fmt.Sprintf("parent extractor for %s cannot be nil", parentName))
	}
}

func (k *Kind) peekParent() parent {
	if len(k.parents) > 0 {
		a := k.parents[0]
		k.parents = k.parents[1:]
		return a
	}
	panic(fmt.Sprintf("kind %s has no (more?) parent", k.name))
}
