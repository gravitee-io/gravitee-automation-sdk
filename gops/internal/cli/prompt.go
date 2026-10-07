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

package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/naming"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/resource"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/transform"
)

// The prompts are the human-in-the-loop side of the CLI. Each one fills exactly one input a flag
// could have set, and all of them run on the command's own streams.

func run(s streams, groups ...*huh.Group) error {
	return huh.NewForm(groups...).WithInput(s.in).WithOutput(s.out).Run()
}

func askFormat(s streams) (string, error) {
	var format string
	err := run(s, huh.NewGroup(
		huh.NewSelect[string]().
			Title("Export format").
			Description("json and yaml write the Automation API state; crd writes GKO manifests; tf writes Terraform").
			Options(huh.NewOptions(transform.Formats()...)...).
			Value(&format),
	))
	return format, err
}

func askResources(s streams) ([]string, error) {
	var options []huh.Option[string]
	for _, k := range resource.Kinds() {
		options = append(options, huh.NewOption(fmt.Sprintf("%s (%s)", k.Name, k.CRDKind), k.Name))
	}
	var picked []string
	err := run(s, huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Resources to export").
			Options(options...).
			Validate(func(v []string) error {
				if len(v) == 0 {
					return fmt.Errorf("pick at least one kind")
				}
				return nil
			}).
			Value(&picked),
	))
	return picked, err
}

// askStripIDs is a judgment call no rule can make: identifiers tie the manifest to this
// environment (GKO adopts the plans and pages) or go so the manifest moves to another one.
func askStripIDs(s streams) (bool, error) {
	var strip bool
	err := run(s, huh.NewGroup(
		huh.NewConfirm().
			Title("Strip the APIM identifiers from the manifests?").
			Description("Keep them and GKO adopts the existing plans and pages when the manifest is applied in this environment.\nStrip them for a portable manifest: GKO creates new plans and pages.").
			Affirmative("Strip").
			Negative("Keep").
			Value(&strip),
	))
	return strip, err
}

func confirmOverwrite(s streams, paths []string) (bool, error) {
	var ok bool
	err := run(s, huh.NewGroup(
		huh.NewConfirm().
			Title(fmt.Sprintf("%d file(s) already exist. Overwrite?", len(paths))).
			Description(strings.Join(paths, "\n")).
			Affirmative("Overwrite").
			Negative("Stop").
			Value(&ok),
	))
	return ok, err
}

// promptResolver asks the human for one name per colliding resource. Names are not unique in APIM,
// and no rule can tell which "Petstore" is which: the person who knows the environment does.
type promptResolver struct {
	s streams
}

func (p promptResolver) Resolve(collisions []naming.Collision, taken map[string]bool) (map[string]string, error) {
	var groups []*huh.Group
	var slots []*slot
	for _, c := range collisions {
		var fields []huh.Field
		group := []*slot{}
		for i, r := range c.Resources {
			sl := &slot{id: r.ID, value: c.Slug}
			if i > 0 {
				sl.value = fmt.Sprintf("%s-%d", c.Slug, i+1)
			}
			slots = append(slots, sl)
			group = append(group, sl)
			fields = append(fields, huh.NewInput().
				Title(fmt.Sprintf("%q  (id %s)", r.Name, r.ID)).
				Description(fmt.Sprintf("%d resources are named %q; give this one its object name", len(c.Resources), r.Name)).
				Validate(validName(sl, group, taken)).
				Value(&sl.value))
		}
		groups = append(groups, huh.NewGroup(fields...).Title(fmt.Sprintf("Name collision: %s", c.Slug)))
	}
	if err := run(p.s, groups...); err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, sl := range slots {
		out[sl.id] = sl.value
		taken[sl.value] = true
	}
	return out, nil
}

type slot struct {
	id    string
	value string
}

func validName(self *slot, group []*slot, taken map[string]bool) func(string) error {
	return func(v string) error {
		if !naming.IsSlug(v) {
			return fmt.Errorf("use lower-case letters, digits and dashes (%q)", naming.Slug(v))
		}
		if taken[v] {
			return fmt.Errorf("%q is already used by another resource", v)
		}
		for _, other := range group {
			if other != self && other.value == v {
				return fmt.Errorf("%q is already used in this group", v)
			}
		}
		return nil
	}
}
