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

// Package export runs one export: list the requested kinds, settle object names, render each
// resource through the transformer of the requested format and write one file per resource.
package export

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/naming"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/resource"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/transform"
)

// Options are the settled inputs of one export.
type Options struct {
	Format    string
	Resources []string
	Output    string
	Force     bool
	// Renames maps a resource id to the object name it must get, regardless of its display name.
	Renames map[string]string
}

// Deps are the collaborators the CLI wires: the SDK client, how collisions are settled and how
// an overwrite is approved. ConfirmOverwrite nil means no one can approve: existing files fail.
type Deps struct {
	Client           *sdk.APIMClient
	Resolver         naming.Resolver
	ConfirmOverwrite func(paths []string) (bool, error)
	// List overrides Kind.List; tests use it to export without a server.
	List func(ctx context.Context, kind *resource.Kind) ([]resource.Resource, error)
}

// File is one written file.
type File struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// Report is what the export did, for the human summary and the --json output.
type Report struct {
	Files []File   `json:"files"`
	Notes []string `json:"notes,omitempty"`
}

// ExistsError reports files the export would overwrite without --force or an approval.
type ExistsError struct {
	Paths    []string
	Declined bool
}

func (e *ExistsError) Error() string {
	if e.Declined {
		return "overwrite declined"
	}
	return fmt.Sprintf("%d file(s) exist; pass --force to overwrite:\n  %s", len(e.Paths), strings.Join(e.Paths, "\n  "))
}

// Run performs the export.
func Run(ctx context.Context, opts Options, deps Deps) (*Report, error) {
	transformer, err := transform.Lookup(opts.Format)
	if err != nil {
		return nil, err
	}
	list := deps.List
	if list == nil {
		list = func(ctx context.Context, kind *resource.Kind) ([]resource.Resource, error) {
			return kind.List(ctx, deps.Client)
		}
	}

	type target struct {
		res  resource.Resource
		name string
		path string
	}
	var targets []target
	for _, kindName := range opts.Resources {
		kind, ok := resource.Lookup(kindName)
		if !ok {
			return nil, fmt.Errorf("unknown resource kind %q", kindName)
		}
		resources, err := list(ctx, kind)
		if err != nil {
			return nil, err
		}
		names, err := settleNames(resources, opts.Renames, deps.Resolver)
		if err != nil {
			return nil, err
		}
		for _, r := range resources {
			name := names[r.ID]
			targets = append(targets, target{res: r, name: name, path: filepath.Join(opts.Output, kind.Name, name+"."+transformer.Ext())})
		}
	}

	if !opts.Force {
		var existing []string
		for _, t := range targets {
			if _, err := os.Stat(t.path); err == nil {
				existing = append(existing, t.path)
			}
		}
		if len(existing) > 0 {
			if deps.ConfirmOverwrite == nil {
				return nil, &ExistsError{Paths: existing}
			}
			ok, err := deps.ConfirmOverwrite(existing)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, &ExistsError{Paths: existing, Declined: true}
			}
		}
	}

	report := &Report{}
	for _, t := range targets {
		content, notes, err := transformer.Transform(t.res, t.name)
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Dir(t.path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(t.path, content, 0o644); err != nil {
			return nil, err
		}
		report.Files = append(report.Files, File{Kind: t.res.Kind.Name, ID: t.res.ID, Name: t.name, Path: t.path})
		for _, n := range notes {
			report.Notes = append(report.Notes, fmt.Sprintf("%s: %s", t.path, n))
		}
	}
	return report, nil
}

// settleNames names every resource of one kind: slugs first, then the resolver on collisions.
func settleNames(resources []resource.Resource, renames map[string]string, resolver naming.Resolver) (map[string]string, error) {
	named := make([]naming.Named, 0, len(resources))
	for _, r := range resources {
		named = append(named, naming.Named{ID: r.ID, Name: r.Name})
	}
	names, collisions, err := naming.Plan(named, renames)
	if err != nil {
		return nil, err
	}
	if len(collisions) == 0 {
		return names, nil
	}
	if resolver == nil {
		resolver = naming.FailResolver{}
	}
	taken := map[string]bool{}
	for _, n := range names {
		taken[n] = true
	}
	resolved, err := resolver.Resolve(collisions, taken)
	if err != nil {
		return nil, err
	}
	for id, n := range resolved {
		names[id] = n
	}
	return names, nil
}

// IsCollision reports whether err is an unresolved collision, which the CLI prints as data.
func IsCollision(err error) (*naming.CollisionError, bool) {
	var ce *naming.CollisionError
	if errors.As(err, &ce) {
		return ce, true
	}
	return nil, false
}

// SortedPaths is a helper for deterministic reports.
func SortedPaths(files []File) []string {
	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	sort.Strings(paths)
	return paths
}
