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

// Package cli is the gops command tree. Flags are the agent-to-machine interface: every input a
// prompt can gather has a flag, so a script or an agent never meets a prompt. Prompts (huh) only
// fill what a human did not type, and only when both stdin and stdout are terminals.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

// Exit codes, part of the contract: 0 done, 1 reserved for a failed gate (gops score), 2 the run
// failed or the input was incomplete.
const (
	ExitOK      = 0
	ExitFailure = 2
)

type streams struct {
	in       io.Reader
	out, err io.Writer
}

// interactive reports whether prompts may be shown: never with --no-input, never without a terminal.
func (s streams) interactive(noInput bool) bool {
	return !noInput && isTerminal(s.in) && isTerminal(s.out)
}

func isTerminal(v any) bool {
	f, ok := v.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// New builds the root command bound to the given streams.
func New(in io.Reader, out, errOut io.Writer) *cli.Command {
	s := streams{in: in, out: out, err: errOut}
	return &cli.Command{
		Name:            "gops",
		Usage:           "Gravitee automation from the command line",
		Reader:          in,
		Writer:          out,
		ErrWriter:       errOut,
		HideHelpCommand: true,
		ExitErrHandler:  func(context.Context, *cli.Command, error) {}, // Run prints and maps errors itself
		Commands:        []*cli.Command{exportCommand(s)},
	}
}

// Run executes args and returns the process exit code.
func Run(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer) int {
	err := New(in, out, errOut).Run(ctx, args)
	if err == nil {
		return ExitOK
	}
	var coder cli.ExitCoder
	code := ExitFailure
	if errors.As(err, &coder) {
		code = coder.ExitCode()
	}
	if msg := err.Error(); msg != "" {
		fmt.Fprintf(errOut, "gops: %s\n", msg)
	}
	return code
}

// usageError is an incomplete or invalid input: exit 2, message on stderr.
func usageError(format string, a ...any) error {
	return cli.Exit(fmt.Sprintf(format, a...), ExitFailure)
}
