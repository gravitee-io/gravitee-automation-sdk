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
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/config"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/export"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/naming"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/resource"
	"github.com/gravitee-io/gravitee-automation-sdk/gops/internal/transform"
	"github.com/urfave/cli/v3"
)

const (
	collisionPrompt = "prompt"
	collisionSuffix = "suffix"
	collisionFail   = "fail"
)

func exportCommand(s streams) *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "Write the resources of an environment as JSON, YAML, GKO manifests or Terraform",
		Description: `Lists each requested kind through the Automation API and writes one file per resource under
<output>/<kind>/<name>.<ext>. The name is the resource name as a Kubernetes object name; two resources
with the same name are a collision, settled by a prompt in a terminal or by --on-collision / --rename.

Exit codes: 0 done, 2 the run failed or the input was incomplete.`,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: "gops.yaml", Usage: "file holding the API url, organization, environment and credentials", Sources: cli.EnvVars("GOPS_CONFIG")},
			&cli.StringFlag{Name: "format", Aliases: []string{"f"}, Usage: "output format: " + strings.Join(transform.Formats(), ", ")},
			&cli.StringFlag{Name: "resources", Aliases: []string{"r"}, Usage: "comma-separated kinds to export: " + strings.Join(kindNames(), ", ")},
			&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Value: "./gravitee", Usage: "directory to write into"},
			&cli.BoolFlag{Name: "force", Usage: "overwrite existing files"},
			&cli.StringFlag{Name: "on-collision", Usage: "how to settle two resources slugging to the same name: prompt (terminal only), suffix (name, name-2, …), fail (default without a terminal)"},
			&cli.StringSliceFlag{Name: "rename", Usage: "force the object name of a resource, as <id>=<name>; repeatable"},
			&cli.BoolFlag{Name: "json", Usage: "print the report (or the collisions) as JSON on stdout"},
			&cli.BoolFlag{Name: "no-input", Usage: "never prompt; fail when an input is missing", Sources: cli.EnvVars("GOPS_NO_INPUT")},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExport(ctx, cmd, s)
		},
	}
}

func runExport(ctx context.Context, cmd *cli.Command, s streams) error {
	interactive := s.interactive(cmd.Bool("no-input"))
	asJSON := cmd.Bool("json")

	format := cmd.String("format")
	if format == "" {
		if !interactive {
			return usageError("--format is required (%s)", strings.Join(transform.Formats(), ", "))
		}
		var err error
		if format, err = askFormat(s); err != nil {
			return err
		}
	}
	if !slices.Contains(transform.Formats(), format) {
		return usageError("unknown format %q (expected one of %s)", format, strings.Join(transform.Formats(), ", "))
	}

	resources := splitList(cmd.String("resources"))
	if len(resources) == 0 {
		if !interactive {
			return usageError("--resources is required (%s)", strings.Join(kindNames(), ", "))
		}
		var err error
		if resources, err = askResources(s); err != nil {
			return err
		}
	}
	for _, r := range resources {
		if _, ok := resource.Lookup(r); !ok {
			return usageError("unknown resource kind %q (expected one of %s)", r, strings.Join(kindNames(), ", "))
		}
	}

	onCollision := cmd.String("on-collision")
	if onCollision == "" {
		onCollision = collisionFail
		if interactive {
			onCollision = collisionPrompt
		}
	}
	var resolver naming.Resolver
	switch onCollision {
	case collisionPrompt:
		if !interactive {
			return usageError("--on-collision=prompt needs a terminal; use suffix, fail or --rename")
		}
		resolver = promptResolver{s}
	case collisionSuffix:
		resolver = naming.SuffixResolver{}
	case collisionFail:
		resolver = naming.FailResolver{}
	default:
		return usageError("unknown --on-collision %q (prompt, suffix, fail)", onCollision)
	}

	renames, err := parseRenames(cmd.StringSlice("rename"))
	if err != nil {
		return usageError("%s", err)
	}

	cfg, err := config.Load(cmd.String("config"))
	if err != nil {
		return err
	}
	client, err := sdk.NewAPIMClient(cfg.APIContext())
	if err != nil {
		return err
	}

	var confirm func([]string) (bool, error)
	if interactive {
		confirm = func(paths []string) (bool, error) { return confirmOverwrite(s, paths) }
	}
	report, err := export.Run(ctx, export.Options{
		Format:    format,
		Resources: resources,
		Output:    cmd.String("output"),
		Force:     cmd.Bool("force"),
		Renames:   renames,
	}, export.Deps{Client: client, Resolver: resolver, ConfirmOverwrite: confirm})
	if err != nil {
		if ce, ok := export.IsCollision(err); ok && asJSON {
			// The collisions are data for the caller: print them where the report would go.
			printJSON(s, map[string]any{"collisions": ce.Collisions})
			return cli.Exit("", ExitFailure)
		}
		return cli.Exit(err.Error(), ExitFailure)
	}

	if asJSON {
		printJSON(s, report)
		return nil
	}
	for _, f := range report.Files {
		fmt.Fprintf(s.out, "wrote %s  (%s %s)\n", f.Path, f.Kind, f.ID)
	}
	for _, n := range report.Notes {
		fmt.Fprintf(s.err, "note: %s\n", n)
	}
	fmt.Fprintf(s.out, "%d file(s) written\n", len(report.Files))
	return nil
}

func printJSON(s streams, v any) {
	enc := json.NewEncoder(s.out)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func kindNames() []string {
	var names []string
	for _, k := range resource.Kinds() {
		names = append(names, k.Name)
	}
	return names
}

func splitList(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseRenames(values []string) (map[string]string, error) {
	out := map[string]string{}
	for _, v := range values {
		id, name, ok := strings.Cut(v, "=")
		if !ok || id == "" || name == "" {
			return nil, fmt.Errorf("--rename %q: expected <id>=<name>", v)
		}
		out[id] = name
	}
	return out, nil
}
