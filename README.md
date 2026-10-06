# Gravitee Automation SDKs

Go clients and a mock server for the Gravitee Access Management (AM) Automation API.

Aim: talk to AM the same way GKO does (idempotent PUT, org/env scoped), and test that against a local mock.

## Layout

```
common/            shared: API context, errors, response helpers, in-memory store
am-sdk/            AM Automation client (generated + thin facade)
am-mock-server/    in-memory HTTP mock of the same API
apim-sdk/          APIM Automation client (generated from the merged APIM + AIM document)
```

## AM SDK

Import the facade as `am`:

`github.com/gravitee-io/gravitee-automation-sdk/am-sdk/v2/pkg`

`am.NewClient` takes an `apicontext.APIContext` (base URL, org, env, **one** of bearer or basic) and returns an `AMClient`. Org/env default to `DEFAULT`.

**Capabilities** — list / upsert (PUT) / get / delete, called directly on the client (e.g. `client.ListDomainsWithResponse(ctx)`):

| Resource | Scope |
|----------|--------|
| Domains | environment |
| Data planes | environment |
| Certificates | domain |
| Identity providers | domain |
| Reporters | domain |

Models expose `WithDefaults()`, which fills unset fields with the OpenAPI defaults (nested structs included).

Calls return the generated `(resp, err)` pair. `err` is transport/construction only.

**Errors**

| When | What |
|------|------|
| `NewClient` | `errors.ClientError` (bad URL, etc.) |
| Auth missing or both set | `errors.NoAuthProvided` / `errors.ManyAuthProvided` |
| HTTP call | use `common/pkg/response`: `IsNotFound`, `IsUnauthorized`, `IsForbidden`, `IsServerError`, `IsNetworkError` |
| 200 body | `response.Payload[T](resp)` |

Not an API reference — generated methods and models live in `am-sdk/pkg/sdk`.

## Mock server

```bash
go run ./am-mock-server --port 8080
go run ./am-mock-server --auth-file am-mock-server/examples/auth.yaml
go run ./am-mock-server --dry-run-reject
```

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | HTTP listen port |
| `--base-path` | `/automation` | API base path |
| `--auth-file` | | Auth YAML (optional). Sample: `am-mock-server/examples/auth.yaml` |
| `--dry-run-reject` | off | PUT `?dryRun=true` returns `200` `[{severity: ERROR, message}]` and does **not** persist. Without this flag, `dryRun` is ignored. |

Like AM, upserts fill unset fields with their OpenAPI defaults, so the PUT and GET responses carry them.
Requests are validated against the OpenAPI spec: an invalid body or parameter gets a `400` with the failing field; read-only fields sent by clients are ignored.

## Prerequisites

- Go 1.26+
- [Task](https://taskfile.dev) 3.x (`brew install go-task` or `go install github.com/go-task/task/v3/cmd/task@latest`)

## Build & Test

```bash
go test ./am-sdk/... ./am-mock-server/... ./common/...
go test ./am-mock-server/server/...
task sync-oas          # AM_OAS_BRANCH to override
go generate ./am-sdk/... ./am-mock-server/...
```

`go test ./...` does not work from the workspace root — pass module paths.

## Code generation

Spec: `am-sdk/openapi/openapi.yaml`. Overlays reshape it; `oapi-codegen` emits clients and the mock server.

Do not edit `*.gen.go`.

## License

[Apache License 2.0](LICENSE)
