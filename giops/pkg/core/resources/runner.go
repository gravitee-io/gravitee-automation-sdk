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
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/config"
	"github.com/gravitee-io/gravitee-automation-sdk/giops/pkg/core"
	"github.com/labstack/gommon/color"
	"github.com/spf13/cobra"
)

// verb is a resource command: "giops <module> <verb> <resource>".
type verb string

const (
	getVerb    verb = "get"
	applyVerb  verb = "apply"
	deleteVerb verb = "delete"
)

var verbs = []verb{getVerb, applyVerb, deleteVerb}

// Command is one resource of a module: how its files decode (Kind) and the functions that get, apply and
// delete it. Register it in a Registry.
//
// Every function receives the command context, which holds the module client (read it with core.Client),
// the Module, and its own copy of the Kind. It returns the SDK call's error, or core.ExpectSuccess of its
// response: giops prints the error and exits non-zero.
type Command struct {
	// Module is the module the resource belongs to: "giops <module> get <resource>".
	Module core.Module
	Getter
	Applier
	Deleter
	// Kind names the resource on the command line, holds its decoded file and its parents.
	Kind          Kind
	getOptions    *getOptions
	applyOptions  *applyOptions
	deleteOptions *deleteOptions
}

// PutterFunc sends the file, decoded into Kind.GetObject, to the API. options holds the resource's
// AdditionalOptions, nil when it has none.
//
// As Applier.Put, it creates or updates the resource and returns nil when the API saved it.
//
// As Applier.DryRun, it asks the API to validate the file without saving it, and returns:
//   - nil when the API found nothing;
//   - a core.DynRunResult with the API's findings: giops prints them, and stops when it has Errors;
//   - core.ErrNoDryRunResult when the answer has no body to read the findings from;
//   - any other error when the dry run could not run.
type PutterFunc func(ctx context.Context, module core.Module, kind Kind, options *AdditionalOptions) error

// DeleterFunc acts on the resource whose id is id. With delete --file, Kind.GetObject holds the file, so
// GetParent can read the parents from it; otherwise the object is empty and the parents come from their flags.
//
// As Deleter.Delete, it deletes the resource and returns nil when the API deleted it.
//
// As Deleter.DryRunDelete, it asks the API whether the delete would pass, without deleting, and returns what
// Applier.DryRun returns: nil, a core.DynRunResult, core.ErrNoDryRunResult, or the error that stops delete.
type DeleterFunc func(ctx context.Context, module core.Module, kind Kind, id string) error

// GetFunc returns the resource whose id is id, as a pointer to the SDK struct the columns read.
type GetFunc func(ctx context.Context, module core.Module, kind Kind, id string) (any, error)

// ListFunc returns every resource, under the parents GetParent gives, each as a pointer to the SDK struct the
// columns read.
type ListFunc func(ctx context.Context, module core.Module, kind Kind) ([]any, error)

// Applier holds what apply calls. Put is required. DryRun runs before Put unless --yes is given; it is nil
// when the API has no dry run for the resource: apply then says so and still asks. AdditionalOptions adds the
// resource's own apply flags, nil when it has none.
type Applier struct {
	// Put creates or updates the resource. Required.
	Put PutterFunc
	// DryRun validates the file without saving it. Nil when the API has no dry run.
	DryRun PutterFunc
	// AdditionalOptions are the resource's own apply flags. Nil when it has none.
	AdditionalOptions *AdditionalOptions
}

// Deleter holds what delete calls. Delete is required. DryRunDelete runs before Delete unless --yes is given,
// and returns what Applier.DryRun returns; it is nil when the API has no delete dry run for the resource.
// Either way, delete asks before deleting, unless --yes or --dry-run is given.
type Deleter struct {
	// Delete deletes the resource. Required.
	Delete DeleterFunc
	// DryRunDelete checks the delete without deleting. Nil when the API has no delete dry run.
	DryRunDelete DeleterFunc
}

// Getter holds what get calls. Get and Columns are required. List is nil when the API cannot list the
// resource: get without an id then fails. Columns returns the columns of the table get prints without -o.
type Getter struct {
	// Get returns one resource by id. Required.
	Get GetFunc
	// List returns every resource. Nil when the API cannot list them.
	List ListFunc
	// Columns returns the table columns, read from what Get and List return. Required.
	Columns func() []*Column
}

// add finds or creates "<module> <verb> <subject>" under main for each verb, in the order of verbs.
func (c *Command) add(main *cobra.Command, subject string) []*cobra.Command {
	commands := make([]*cobra.Command, 0)
	moduleCmd := c.Module.Command(main)
	for _, verb := range verbs {
		verbCmd := core.FindOrCreate(moduleCmd, string(verb))
		commands = append(commands, core.FindOrCreate(verbCmd, subject))
	}
	return commands
}

// runGet prints the resource named by the first argument, or every resource without one.
func (c *Command) runGet(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	if len(args) == 1 {
		resource, err := c.Get(cmd.Context(), c.Module, c.Kind, args[0])
		if err != nil {
			return fmt.Errorf("cannot get resource: %w", err)
		}
		return c.getOptions.Output.Format(out, resource, c.Columns())
	}
	if c.List == nil {
		return fmt.Errorf("%s list is not supported, id is required", cmd.Use)
	}
	resources, err := c.List(cmd.Context(), c.Module, c.Kind)
	if err != nil {
		return fmt.Errorf("cannot list resources: %w", err)
	}
	return c.getOptions.Output.formatList(out, resources, c.Columns())
}

// runApply decodes the file, runs the dry run and asks, unless --yes, then puts the resource.
func (c *Command) runApply(cmd *cobra.Command, _ []string) error {
	options := c.applyOptions
	file := options.File.GetValueAsString()
	if file == "" {
		return errors.New("no file specified, use -f or --file")
	}
	if err := config.FileToStruct(file, c.Kind.object); err != nil {
		return fmt.Errorf("cannot load or parse file: %w", err)
	}

	out := cmd.OutOrStdout()
	if !options.Force.GetValueAsBool() {
		if err := c.dryRun(cmd.Context(), out); err != nil {
			return err
		}
		if options.DryRunOnly.GetValueAsBool() || !core.Confirm(cmd.InOrStdin(), out, "Apply?") {
			return nil
		}
	}

	id := c.Kind.object.Identity()
	if err := c.Put(cmd.Context(), c.Module, c.Kind, c.AdditionalOptions); err != nil {
		return fmt.Errorf("could not apply %s %s: %w", cmd.Use, id, err)
	}
	_, err := fmt.Fprintf(out, "%s %s applied successfully\n", cmd.Use, id)
	return err
}

// runDelete deletes the resource named by --file, --id or the first argument, in that order.
func (c *Command) runDelete(cmd *cobra.Command, args []string) error {
	var id string
	if file := c.deleteOptions.File.GetValueAsString(); file != "" {
		if err := config.FileToStruct(file, c.Kind.object); err != nil {
			return fmt.Errorf("cannot load or parse file: %w", err)
		}
		id = c.Kind.object.Identity()
	} else {
		var err error
		id, err = core.Or(c.deleteOptions.ID.GetValueAsString(), args, 0, "no resource to delete, provide id with --id or --file -f to load the resource to delete")
		if err != nil {
			return err
		}
	}

	out := cmd.OutOrStdout()
	if !c.deleteOptions.Force.GetValueAsBool() {
		if err := c.dryRunDelete(cmd.Context(), id, out); err != nil {
			return err
		}
		if c.deleteOptions.DryRunOnly.GetValueAsBool() || !core.Confirm(cmd.InOrStdin(), out, "Apply?") {
			return nil
		}
	}
	if err := c.Delete(cmd.Context(), c.Module, c.Kind, id); err != nil {
		return fmt.Errorf("could not delete %s %s: %w", cmd.Use, id, err)
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s deleted\n", cmd.Use, id)
	return err
}

// dryRun prints the result of DryRun, and returns an error when apply must stop there.
func (c *Command) dryRun(ctx context.Context, out io.Writer) error {
	if c.DryRun == nil {
		_, err := fmt.Fprintf(out, "Dry run is not configured\n")
		return err
	}
	err := c.DryRun(ctx, c.Module, c.Kind, c.AdditionalOptions)
	return c.processError(err, out)
}

// dryRunDelete prints the result of DryRunDelete, and returns an error when delete must stop there.
func (c *Command) dryRunDelete(ctx context.Context, id string, out io.Writer) error {
	if c.DryRunDelete == nil {
		return nil
	}
	err := c.DryRunDelete(ctx, c.Module, c.Kind, id)
	return c.processError(err, out)
}

// processError prints the result of a dry run callback. It returns an error when the dry run did not complete
// or found errors; warnings alone do not stop.
func (c *Command) processError(err error, out io.Writer) error {
	var result core.DynRunResult
	switch {
	case err == nil:
		_, err = fmt.Fprintln(out, color.Green("Dry run result: successful", color.B))
		return err
	case !errors.As(err, &result):
		return fmt.Errorf("dry run did not complete: %w", err)
	case result.HasErrors():
		return fmt.Errorf("dry run result:\n%w", result)
	default:
		_, err = fmt.Fprint(out, "Dry run result:\n", result.Error())
		return err
	}
}
