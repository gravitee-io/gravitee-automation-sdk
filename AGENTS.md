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

### Build & Test

```bash
# Install lint/dev tools (staticcheck, revive, addlicense, goimports)
make tools

# Sync AM and APIM Automation OAS (AM_OAS_BRANCH / APIM_OAS_BRANCH, default master)
make sync-oas

# Sync the AIM fragment from the private gravitee-gamma-module-aim (AIM_OAS_REF); needs a gh login with access
make sync-aim-oas

# Fail if the committed AIM fragment differs from AIM_OAS_REF. CI runs it in CircleCI (.circleci/config.yml),
# triggered by the aim-oas GitHub Actions job through hack/scripts/verify-aim-oas.mjs, because only CircleCI
# holds (in Keeper) the token that reads the private repo
make verify-aim-oas

# Regenerate all code (overlays + oapi-codegen)
make generate

# Run all linters (vet, staticcheck, revive, license headers)
make lint

# Auto-fix lint issues and add license headers
make lint-fix

# Run all tests across all modules
make test

# Run tests for a single module
go test ./am-mock-server/server/...

# Run a single test
go test ./am-mock-server/server/... -run TestGetDomain404SDK

# Run the mock servers
go run ./am-mock-server --port 8080
go run ./apim-mock-server --port 8080
```

---

## 2. Project Context

### What Is This?

Go SDK clients and mock servers for the Gravitee Access Management (AM) and API Management (APIM) Automation APIs. The SDK is generated from an OpenAPI spec using `oapi-codegen`, with OpenAPI Overlay files applied to reshape the spec before generation.

### Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.26 |
| Build | Go workspace (`go.work`) |
| Code generation | oapi-codegen + OpenAPI Overlay 1.1.0 |
| HTTP (mock server) | chi v5 |
| CLI (mock server) | cobra |
| Tests | stdlib `testing` + testify |

### Modules

A Go workspace with five modules:

| Module | Role |
|--------|------|
| `common` | Shared utilities: `apicontext` (auth + base URL), `response` (status helpers + generic `Payload[T]` extractor), `store` (channel-based in-memory generic store), `errors`, `refs` |
| `am-sdk` | Generated SDK clients for AM resources (domains, certificates, identity providers, reporters). All resources are generated into the single `am-sdk/pkg/sdk` package |
| `am-mock-server` | Standalone mock HTTP server implementing the same OpenAPI spec with strict-server codegen. Used for integration-testing the SDK |
| `apim-sdk` | Generated SDK client for the APIM Automation API (core APIM + the AIM fragment) in `apim-sdk/pkg/sdk`, with hand-written `oneof.go` (polymorphic types), `identifiable.go` (list item identities) and `apicontext.go` (`APIContext` with cloud-token resolution) |
| `apim-mock-server` | Standalone mock HTTP server for the APIM Automation API (strict-server codegen from the merged APIM spec). Mirrors real APIM where the AM mock differs: 404 on missing hrids, deterministic ids, PUT returns the State, dry run in the handlers, parent and subscription reference checks, cascades |

---

## 3. Code Generation Pipeline

AM code comes from the OpenAPI spec at `am-sdk/openapi/openapi.yaml`, APIM code from `apim-sdk/openapi/openapi.yaml`. The APIM spec is itself generated: `common/cmd/mergespec` merges `apim-sdk/openapi/sources/apim.yaml` (core APIM) and `sources/aim.yaml` (AIM) into it (`apim-sdk/openapi/generate.go`), and it is committed. Generation is driven by `//go:generate` directives in `generate.go` files and involves two steps:

1. **Overlay merge** — `common/cmd/mergeoverlay` is a CLI tool that merges multiple YAML overlay files into one. Overlays rename models (strip `Automation` prefix via `x-go-type-name`), standardize operationIds (`get`, `list`, `upsert`, `delete`), and (for SDK clients) rewrite paths to bake `orgId`/`envId` into the server URL.
2. **oapi-codegen** — Generates typed Go clients or strict servers from the overlaid spec. Configs live under `<module>/gen/` so package directories only hold Go files (the SDK uses `am-sdk/gen/models.cfg.yaml` → `models.gen.go` and `am-sdk/gen/client.cfg.yaml` → `client.gen.go`; the mock server uses `am-mock-server/gen/server/cfg.yaml`). The experimental CRD package keeps its config and overlay in place. Paths inside a config are relative to the package directory, where `go generate` runs.

### Overlay Chains

The overlay chain differs between SDK and mock server:

| Target | Overlays merged (from `am-sdk/overlays/` or `apim-sdk/overlays/`, into a git-ignored `overlay.gen.yaml` next to the target's config under `gen/`) | What is generated |
|--------|----------------|-------------------|
| SDK (`am-sdk/pkg/sdk/`) | `models.yaml` + `operations.yaml` + `paths.yaml` (bakes org/env into the server URL and strips them from paths) | Client + models (all tags), with `WithDefaults()` from `common/templates/typedef.tmpl` |
| Mock server (`am-mock-server/server/`) | `models.yaml` + `operations.yaml` | chi strict-server + models (all tags, original paths kept) |
| CRD models (`am-sdk/pkg/crd/domain/`) | `models.yaml` + `operations.yaml` + its own `overlay.yaml` (experimental) | Models only |
| APIM SDK (`apim-sdk/pkg/sdk/`) | `models.yaml` + `oneof.yaml` (polymorphic types → hand-written `OneOf*`) + `drift.yaml` (GKO drift tags) + `operations.yaml` + `paths.yaml` | Client + models, with `WithDefaults()` |
| APIM mock server (`apim-mock-server/server/`) | `models.yaml` + `oneof.yaml` + the mock's own `gen/server/oneof.yaml` (polymorphic types → `unstructured.Unstructured`, so the server does not depend on the SDK's hand-written types) + `operations.yaml` | chi strict-server + models (original paths kept, no drift tags) |

`common/templates/typedef.tmpl` (`WithDefaults()`) is shared by every target. It fills unset pointers, nil slices and maps with their defaults, and recurses into named nested types. It never sets non-pointer scalars (zero is indistinguishable from unset) and never recurses into anonymous structs.

In `apim-sdk/pkg/sdk`, `oneof.go` must have one `OneOf<Name>` struct per `x-go-type: OneOf<Name>` in `oneof.yaml`, with a field per variant (`oneof:"<TYPE>"` tag). A new upstream variant needs a new field: unmarshal fails on an unknown `type`.

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
