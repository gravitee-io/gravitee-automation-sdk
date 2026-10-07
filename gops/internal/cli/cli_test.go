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
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Without a terminal (buffers here) nothing may prompt: a missing input is exit 2 with the flag named.
func execute(t *testing.T, args ...string) (int, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Run(context.Background(), append([]string{"gops"}, args...), bytes.NewReader(nil), &out, &errOut)
	return code, out.String() + errOut.String()
}

func TestExport_NonInteractiveContract(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "gops.yaml")
	_ = os.WriteFile(cfg, []byte("url: http://localhost:8083/automation\nauth:\n  token: t\n"), 0o600)

	code, msg := execute(t, "export", "--config", cfg)
	assert.Equal(t, ExitFailure, code)
	assert.Contains(t, msg, "--format is required")

	code, msg = execute(t, "export", "--config", cfg, "--format", "crd")
	assert.Equal(t, ExitFailure, code)
	assert.Contains(t, msg, "--resources is required")

	code, msg = execute(t, "export", "--config", cfg, "--format", "hcl", "--resources", "apis")
	assert.Equal(t, ExitFailure, code)
	assert.Contains(t, msg, `unknown format "hcl"`)

	code, msg = execute(t, "export", "--config", cfg, "--format", "crd", "--resources", "apis,nope")
	assert.Equal(t, ExitFailure, code)
	assert.Contains(t, msg, `unknown resource kind "nope"`)

	code, msg = execute(t, "export", "--config", cfg, "--format", "crd", "--resources", "apis", "--on-collision", "prompt")
	assert.Equal(t, ExitFailure, code)
	assert.Contains(t, msg, "needs a terminal")

	code, msg = execute(t, "export", "--config", cfg, "--format", "crd", "--resources", "apis", "--rename", "nope")
	assert.Equal(t, ExitFailure, code)
	assert.Contains(t, msg, "expected <id>=<name>")

	code, msg = execute(t, "export", "--config", cfg, "--format", "crd", "--resources", "apis", "--ids", "drop")
	assert.Equal(t, ExitFailure, code)
	assert.Contains(t, msg, `unknown --ids "drop"`)

	code, msg = execute(t, "export", "--config", filepath.Join(t.TempDir(), "missing.yaml"), "--format", "crd", "--resources", "apis")
	assert.Equal(t, ExitFailure, code)
	assert.Contains(t, msg, "read config")
}

func TestHelp(t *testing.T) {
	code, msg := execute(t, "export", "--help")
	assert.Equal(t, ExitOK, code)
	assert.Contains(t, msg, "--on-collision")
}
