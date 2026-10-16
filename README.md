# Gravitee Automation SDKs

Go clients and mock servers for the Gravitee Access Management (AM) and API Management (APIM) Automation APIs.

Aim: talk to AM the same way GKO does (idempotent PUT, org/env scoped), and test that against a local mock.

## Layout

```
common/            shared: API context, errors, response helpers, in-memory store
am-sdk/            AM Automation client (generated + thin facade)
am-mock-server/    in-memory HTTP mock of the same API
apim-sdk/          APIM Automation client (generated, core APIM + AIM)
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

## APIM SDK

Import the generated package:

`github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk`

`sdk.NewAPIMClient` takes an `sdk.APIContext` and returns an `APIMClient`. `APIContext` embeds the common
`apicontext.APIContext` (base URL, org, env, **one** of bearer or basic) and adds `CloudToken`.
`BaseURL` is the full automation URL, as for AM (`http://localhost:8083/automation`).

```go
client, err := sdk.NewAPIMClient(sdk.APIContext{
	APIContext: apicontext.APIContext{Auth: apicontext.Auth{BearerToken: &cloudToken}},
	CloudToken: true, // org, env and URL come from the token
})
```

With `CloudToken`, `Auth.BearerToken` is a Gravitee Cloud token (GKO's rules):
- its signature is not checked; the claims `org`, `envs` and `cpg` are required
- `OrgID` is always the token's `org`
- an empty `EnvID` is the token's only env (several envs is an error); a given `EnvID` must be in `envs`
- an empty `BaseURL` becomes `https://{cpg}.cloudgate.gravitee.io/apim/automation`; a given one is kept

**Capabilities** — upsert (PUT) / get / delete (e.g. `client.UpsertApiWithResponse(ctx, params, body)`):

| Resource | Scope |
|----------|--------|
| APIs (v4), applications, groups, dictionaries, shared policy groups, portals, portal themes | environment |
| AIM catalog MCP servers, AIM MCP proxies | environment |
| Subscriptions, documentations, links | API |
| Listings, documentations, links | portal |

Errors and response helpers are the same as the AM SDK's.

### Differences from AM

- **Two sources.** Core APIM (`gravitee-api-management` `master`) and the AIM fragment (private
  `gravitee-gamma-module-aim`, pinned tag) are merged by `common/cmd/mergespec` into `apim-sdk/openapi/openapi.yaml`.
  `make sync-aim-oas` and `make verify-aim-oas` need a `gh` login with access to the AIM repo. CI runs the AIM check
  in CircleCI, which holds the token in Keeper (as terraform-provider-apim does). `make check-oas` covers AM and core APIM.
- **No list operations**: upstream has none.
- **Model names** drop `Spec` and keep `State` (`ApiV4` is the PUT body, `ApiV4State` the response).
- **Dates** are RFC 3339 strings (`*string`), not `time.Time`.
- **Free-form configurations** (plugin, endpoint, page source, subscription entrypoint) are `unstructured.Unstructured`.
- **`crossId` is writable** on `ApiV4` and `SharedPolicyGroup`. APIM honours it on PUT for legacy resources
  (first created with a UUID), but the spec omits it.
- **Versioning**: `main` tracks APIM `master`. As a rule, each APIM minor gets its own SDK release line, synced from
  that APIM release branch; pick the line that matches your APIM. Only `main` ships for now.
- **Polymorphic types** (`Listener`, `Selector`, and the AIM `AimCatalogMcpServerAuth`, `AimMcpProxyUpstreamAuth`,
  `AimMcpProxyPlanSecurity`, `AimMcpProxyFlowSelector`, `AimMcpProxyIdentityProvider`) are hand-written `OneOf*` structs
  (`pkg/sdk/oneof.go`) with one pointer per variant. On the wire they are the variant plus `type`; set exactly one.
  The AIM discriminators have no `oneOf` upstream: the overlay moves their fields to a `*Base*` schema and declares the
  `oneOf` of their variants, as core APIM does. `DynamicDictionary.provider` is flattened to `HttpDictionaryProvider`.
- **`WithDefaults()`** never sets non-pointer scalars (a `false` stays `false`) and never recurses into anonymous
  structs. The template is shared, so AM gets the same rules; its output is unchanged.
- **List items with an identity** (`Member`, `GroupMember`, `Metadata`, `ApiKey`, `ClientCertificate`) implement
  `store.Identifiable` (`pkg/sdk/identifiable.go`), with GKO's keys.
- **Drift tags** are ported from GKO's APIM DTOs (`overlays/drift.yaml`), with these changes:
  - `entrypointConnect` replaces GKO's flow `connect`
  - explicit `ignore` on `readOnly` properties next to a `$ref` (the parser drops that `readOnly`)
  - generic `readOnly`/`writeOnly` rules inside `allOf` branches (status fields, AIM secrets)
  - overrides (`ignore-only:…`, `ignore-namespace-prefix`) placed after the generic rules
  - extra tags beyond GKO's: `empty-is-nil` on fields that became pointers, `ignore` on `Metadata.key` and on
    dictionary `encryptedProperties`
  - targets use `['a','b']` unions: if APIM renames one property of a union, its tag is dropped without an error

### Using the SDK from GKO

- **Drift on the GET response**: decode it into the PUT type (e.g. `ApiV4State` → `ApiV4`) before drift. Comparing
  two different types panics.
- **Drift engine changes**: `ignore-only` must pair items through `store.Identifiable`; `ignore` must be registered
  for numeric kinds (`GroupStatus.memberCount`); client-certificate `Expired`/`Scheduled` must move into the engine.
- **Never call `WithDefaults()` on the CRD side.** Its defaults differ from the CRD's (e.g. API `visibility` PUBLIC
  vs PRIVATE), and it would defeat `ignore-remote-default` and `empty-is-nil`.
- **Map flow `connect` to `entrypointConnect`.** GKO sends `connect` today, and APIM ignores it.
- **Required fields are now sent as zero values** (`responseTemplates.*.status`, member `role`, endpoint and page
  `name`): the CRD validation must cover them.
- **Branches**: only GKO master and 4.13.x can adopt it (4.12.x has neither these DTOs nor `internal/drift`).
  Suggested order: SharedPolicyGroup → Dictionary → Application → API v4.
- **Keep GKO's HTTP client** until the spec declares `hridContainsUUID`.
- **Pin the SDK release line** matching the branch's APIM minor.

## Prerequisites

- Go 1.26+

## Build & Test

```bash
go test ./am-sdk/... ./apim-sdk/... ./am-mock-server/... ./common/...
go test ./am-mock-server/server/...
make sync-oas          # AM_OAS_BRANCH / APIM_OAS_BRANCH to override
make sync-aim-oas      # AIM_OAS_REF to override; needs a gh login with access to gravitee-gamma-module-aim
make generate
```

`go test ./...` does not work from the workspace root — pass module paths.

## Code generation

Specs: `am-sdk/openapi/openapi.yaml` and `apim-sdk/openapi/openapi.yaml` (merged from `apim-sdk/openapi/sources/`).
Overlays reshape them; `oapi-codegen` emits clients and the mock server.

Do not edit `*.gen.go`.

## License

[Apache License 2.0](LICENSE)
