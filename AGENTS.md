# AGENTS.md — Gravitee Automation SDKs

Universal AI agent instructions for the Gravitee Automation SDKs repository.
Consumed by: Claude Code, OpenAI Codex, GitHub Copilot, Cursor, Gemini/Jules, Windsurf, Zed, Aider, and others.

---

# Agent Guide

If you are an AI agent operating in this repository:
- You MUST read this file fully before making changes.
- You MUST follow the rules defined here.
- If any instruction conflicts with other files, this file takes precedence.

## 1. Quick Reference Commands

### Prerequisites

- Go 1.26+
- [Task](https://taskfile.dev) 3.x (`brew install go-task` or `go install github.com/go-task/task/v3/cmd/task@latest`)

### Build & Test

```bash
# Install lint/dev tools (staticcheck, revive, addlicense, goimports)
task tools

# Sync AM Automation OAS from gravitee-access-management (AM_OAS_BRANCH, default master)
task sync-oas

# Regenerate all code (overlays + oapi-codegen)
task generate

# Run all linters (vet, staticcheck, revive, license headers)
task lint

# Auto-fix lint issues and add license headers
task lint:fix

# Run all tests across all modules
task test

# Run tests for a single module
go test ./am-mock-server/server/...

# Run a single test
go test ./am-mock-server/server/... -run TestGetDomain404SDK

# Run the mock server
go run ./am-mock-server --port 8080
```

---

## 2. Project Context

### What Is This?

Go SDK clients and a mock server for the Gravitee Access Management (AM) Automation API. The SDK is generated from an OpenAPI spec using `oapi-codegen`, with OpenAPI Overlay files applied to reshape the spec before generation.

### Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.26 |
| Build | Go workspace (`go.work`), [Task](https://taskfile.dev) (`Taskfile.yml`) |
| Code generation | oapi-codegen + OpenAPI Overlay 1.1.0 |
| HTTP (mock server) | chi v5 |
| CLI (mock server) | cobra |
| Tests | stdlib `testing` + testify |

### Modules

A Go workspace with four modules:

| Module | Role |
|--------|------|
| `common` | Shared utilities: `apicontext` (auth + base URL), `response` (status helpers + generic `Payload[T]` extractor), `store` (channel-based in-memory generic store), `errors`, `refs` |
| `am-sdk` | Generated SDK clients for AM resources (domains, certificates, identity providers, reporters). All resources are generated into the single `am-sdk/pkg/sdk` package |
| `am-mock-server` | Standalone mock HTTP server implementing the same OpenAPI spec with strict-server codegen. Used for integration-testing the SDK |
| `apim-sdk` | Generated SDK client for the APIM Automation API, from the APIM document merged with the AI Management fragment (`common/cmd/mergespec`), in `apim-sdk/pkg/sdk`; `apim-sdk/openapi` embeds the merged document |

---

## 3. Code Generation Pipeline

All generated code comes from a single OpenAPI spec at `am-sdk/openapi/openapi.yaml`. Generation is driven by `//go:generate` directives in `generate.go` files and involves two steps:

1. **Overlay merge** — `common/cmd/mergeoverlay` is a CLI tool that merges multiple YAML overlay files into one. Overlays rename models (strip `Automation` prefix via `x-go-type-name`), standardize operationIds (`get`, `list`, `upsert`, `delete`), and (for SDK clients) rewrite paths to bake `orgId`/`envId` into the server URL.
2. **oapi-codegen** — Generates typed Go clients or strict servers from the overlaid spec. Configs live under `<module>/gen/` so package directories only hold Go files (the SDK uses `am-sdk/gen/models.cfg.yaml` → `models.gen.go` and `am-sdk/gen/client.cfg.yaml` → `client.gen.go`; the mock server uses `am-mock-server/gen/server/cfg.yaml`). Paths inside a config are relative to the package directory, where `go generate` runs.

### Overlay Chains

The overlay chain differs between SDK and mock server:

| Target | Overlays merged (from `am-sdk/overlays/`, into a git-ignored `overlay.gen.yaml` next to the target's config under `gen/`) | What is generated |
|--------|----------------|-------------------|
| SDK (`am-sdk/pkg/sdk/`) | `models.yaml` + `operations.yaml` + `paths.yaml` (bakes org/env into the server URL and strips them from paths) | Client + models (all tags), with `WithDefaults()` from `am-sdk/templates/typedef.tmpl` |
| Mock server (`am-mock-server/server/`) | `models.yaml` + `operations.yaml` | chi strict-server + models (all tags, original paths kept) |

**Files ending in `.gen.go` are generated — do not edit them.**

---

## 4. Key Patterns

- **`store.Identifiable` interface** — Any type stored in `store.Store[T]` must implement `Identity() string`. The mock server's `identifiers.go` wires generated types to this interface.
- **`response.Payload[T]`** — Generic extractor that pulls `JSON200` from oapi-codegen response structs via reflection. Used in tests and intended for SDK consumers.
- **Test helpers** — `am-mock-server/server/helpers_test.go` contains reusable assertion functions (`assertSDKOK`, `assertGet404`, etc.) used across all resource test files. Each resource's tests exercise both raw HTTP and SDK client paths.

---

## 5. Conventions

- Search for similar implementations first before writing new code.
- Prefer existing patterns over inventing new ones.
- Keep changes small and reviewable; avoid unrelated refactors or reformatting.
