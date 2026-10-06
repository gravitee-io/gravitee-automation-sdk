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

package naming

import (
	"fmt"
	"sort"
	"strings"
)

// Named is anything the plan names: a resource with a stable id and a display name.
type Named struct {
	ID   string
	Name string
}

// Collision groups the resources whose names slug to the same object name.
type Collision struct {
	Slug      string
	Resources []Named
}

// Plan assigns an object name to every resource. Overrides (id → name) win over the slug of the
// display name. It returns the names settled without conflict and the collisions left to resolve;
// a collision is two or more resources of one kind sharing a name, override or slug alike.
func Plan(resources []Named, overrides map[string]string) (names map[string]string, collisions []Collision, err error) {
	for id, name := range overrides {
		if !IsSlug(name) {
			return nil, nil, fmt.Errorf("rename %s=%s: %q is not a valid object name (expected %q)", id, name, name, Slug(name))
		}
	}
	byName := map[string][]Named{}
	for _, r := range resources {
		name, ok := overrides[r.ID]
		if !ok {
			name = Slug(r.Name)
		}
		byName[name] = append(byName[name], r)
	}
	names = map[string]string{}
	for name, rs := range byName {
		if len(rs) == 1 {
			names[rs[0].ID] = name
			continue
		}
		sort.Slice(rs, func(i, j int) bool { return rs[i].ID < rs[j].ID })
		collisions = append(collisions, Collision{Slug: name, Resources: rs})
	}
	sort.Slice(collisions, func(i, j int) bool { return collisions[i].Slug < collisions[j].Slug })
	return names, collisions, nil
}

// Resolver settles collisions into one name per resource id. taken holds the names already
// assigned in the run, which a resolution must not reuse.
type Resolver interface {
	Resolve(collisions []Collision, taken map[string]bool) (map[string]string, error)
}

// SuffixResolver keeps the first resource (by id) on the bare slug and numbers the others:
// petstore, petstore-2, petstore-3. Deterministic, so an agent gets the same names on every run.
type SuffixResolver struct{}

// Resolve implements Resolver.
func (SuffixResolver) Resolve(collisions []Collision, taken map[string]bool) (map[string]string, error) {
	out := map[string]string{}
	for _, c := range collisions {
		n := 1
		for _, r := range c.Resources {
			name := c.Slug
			for taken[name] {
				n++
				name = fmt.Sprintf("%s-%d", c.Slug, n)
			}
			taken[name] = true
			out[r.ID] = name
		}
	}
	return out, nil
}

// FailResolver refuses to guess: it returns a CollisionError naming every collision so the
// caller (an agent, a CI job) can rerun with explicit --rename flags or --on-collision=suffix.
type FailResolver struct{}

// Resolve implements Resolver.
func (FailResolver) Resolve(collisions []Collision, _ map[string]bool) (map[string]string, error) {
	return nil, &CollisionError{Collisions: collisions}
}

// CollisionError reports unresolved name collisions.
type CollisionError struct {
	Collisions []Collision
}

func (e *CollisionError) Error() string {
	var b strings.Builder
	b.WriteString("resource names collide; pass --rename <id>=<name> for each, or --on-collision=suffix:")
	for _, c := range e.Collisions {
		for _, r := range c.Resources {
			fmt.Fprintf(&b, "\n  %s: %q (id %s)", c.Slug, r.Name, r.ID)
		}
	}
	return b.String()
}
