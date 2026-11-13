# Gravitee Automation SDKs

Go clients and mock servers for the Gravitee Access Management (AM) and API Management (APIM) Automation APIs.

Aim: talk to AM the same way GKO does (idempotent PUT, org/env scoped), and test that against a local mock.

## Layout

```
common/            shared: API context, errors, response helpers, in-memory store
am-sdk/            AM Automation client (generated + thin facade)
am-mock-server/    in-memory HTTP mock of the same API
apim-sdk/          APIM Automation client (generated, core APIM + AIM)
apim-mock-server/  in-memory HTTP mock of the APIM Automation API
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
| Applications | domain |
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
A running mock can also answer an error instead of serving a request (fault injection), through `MockAM.Faults` in Go or `/_mock/faults` over HTTP: see [am-mock-server/README.md](am-mock-server/README.md#fault-injection).

## APIM SDK

Import the generated package:

`github.com/gravitee-io/gravitee-automation-sdk/apim-sdk/pkg/sdk`

`sdk.NewAPIMClient` takes an `sdk.APIContext` and returns an `APIMClient`. `APIContext` embeds the common
`apicontext.APIContext` (base URL, org, env, **one** of bearer or basic) and adds `CloudToken`. `BaseURL` is the full
automation URL (`http://localhost:8083/automation`). Org/env default to `DEFAULT`.

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

**Capabilities** — upsert (PUT) / get / delete, called directly on the client (e.g.
`client.UpsertApiWithResponse(ctx, params, body)`). APIM has no list operations.

| Resource | Scope |
|----------|--------|
| APIs (v4), applications, groups, dictionaries, shared policy groups, portals, portal themes | environment |
| AIM catalog MCP servers, AIM MCP proxies | environment |
| Subscriptions, documentations, links | API |
| Listings, documentations, links | portal |

Operations with query parameters (`dryRun`, the legacy UUID params) take a params argument, `nil` for none
(`client.GetApiWithResponse(ctx, hrid, nil)`).

**Models**

- `ApiV4` is the PUT body, `ApiV4State` the response: resource types drop `Spec` and keep `State`. Each State's
  `Spec()` returns the PUT body it holds (a shallow copy), e.g. `GroupState.Spec()` → `Group`.
- `WithDefaults()` fills unset pointers, slices and maps with their OpenAPI defaults, nested structs included. It never
  sets non-pointer scalars (a `false` stays `false`) and never recurses into anonymous structs.
- Resources (by hrid) and list items with an identity (`Member`, `GroupMember`, `Metadata`, `ApiKey`,
  `ClientCertificate`) implement `store.Identifiable` (`pkg/sdk/identifiable.go`), with GKO's keys.
- Dates are RFC 3339 strings (`*string`), not `time.Time`.
- Free-form configurations (plugin, endpoint, page source, subscription entrypoint) are `unstructured.Unstructured`.
- Polymorphic types (`Listener`, `Selector`, and the AIM `AimCatalogMcpServerAuth`, `AimMcpProxyUpstreamAuth`,
  `AimMcpProxyPlanSecurity`, `AimMcpProxyFlowSelector`, `AimMcpProxyIdentityProvider`) are hand-written `OneOf*` structs
  (`pkg/sdk/oneof.go`) with one pointer per variant: set exactly one. On the wire they are the variant plus `type`.
  `DynamicDictionary.provider` is an `HttpDictionaryProvider`.
- `crossId` is writable on `ApiV4` and `SharedPolicyGroup`: APIM honours it on PUT for legacy resources (first created
  with a UUID), though the spec omits it.
- Required fields are plain values, so an unset one is sent as its zero value instead of being left out (e.g. a group
  member without `role` is sent with `"role": ""`): check `responseTemplates.*.status`, member `role`, endpoint and
  page `name` before a PUT.
- The legacy UUID params (`hridContainsUUID`, `hridContainsApiUUID`, `hridContainsAppUUID`, the `X-Gravitee-Set-Hrid`
  header) treat an hrid as the UUID of a resource created before the automation API. The spec declares
  `hridContainsUUID` only on the API GET and DELETE and on the group and subscription operations: the API PUT and the
  application, dictionary and shared policy group operations cannot address a legacy resource by its UUID yet.
- The `drift` struct tags (`overlays/drift.yaml`) are read only by GKO's drift engine.

Calls return the generated `(resp, err)` pair. `err` is transport/construction only.

**Errors**

| When | What |
|------|------|
| `NewAPIMClient` | `errors.ClientError` (bad URL, invalid cloud token, etc.) |
| Auth missing or both set | `errors.NoAuthProvided` / `errors.ManyAuthProvided` |
| HTTP call | use `common/pkg/response`: `IsNotFound`, `IsUnauthorized`, `IsForbidden`, `IsServerError`, `IsNetworkError` |
| 200 body | `response.Payload[T](resp)` |
| Error body | `JSON400`, `JSON404`… on the response: APIM's `message` and `details` |

**Versioning**: `main` tracks APIM `master`. As a rule, each APIM minor gets its own SDK release line, synced from
that APIM release branch; pick the line that matches your APIM. Only `main` ships for now. Until v1 the SDK is not used
anywhere, so an API change is a plain `feat` or `fix` commit, never marked `!` or `BREAKING CHANGE:`. Drop this rule at
v1.

Not an API reference — generated methods and models live in `apim-sdk/pkg/sdk`.

### Updating the spec

The spec has two sources, merged by `common/cmd/mergespec` into `apim-sdk/openapi/openapi.yaml`: core APIM
(`gravitee-api-management` `master`, `make sync-oas`, checked by `make check-oas`) and the AIM fragment (private
`gravitee-gamma-module-aim`, pinned by `AIM_OAS_REF` in the `Makefile`). `make sync-aim-oas` and
`make verify-aim-oas` need a `gh` login with access to the AIM repo; CI runs the AIM check in CircleCI, which holds the
token in Keeper (as terraform-provider-apim does).

Nothing bumps the AIM fragment automatically, and `make verify-aim-oas` only checks the committed file against its
tag. To move it:

1. Set `AIM_OAS_REF` to the new tag, then run `make sync-aim-oas generate`.
2. If a discriminated type gained a variant (e.g. a new upstream auth `type`), add it to `overlays/oneof.yaml` and to
   its `OneOf*` struct in `pkg/sdk/oneof.go`, with a round-trip test. Decoding an unknown `type` fails, so a missing
   variant breaks every resource that uses it. The AIM discriminators have no `oneOf` upstream: the overlay moves
   their fields to a `*Base*` schema and declares the `oneOf` of their variants, as core APIM does.
3. Check new properties against `overlays/drift.yaml`: the generic rules tag `readOnly` properties and the
   `writeOnly` ones of variants (`allOf` branches); any other property may need a tag of its own. Targets use
   `['a','b']` unions: if APIM renames one property of a union, its tag is dropped without an error.
4. Run `make lint test`.

## APIM mock server

```bash
go run ./apim-mock-server --port 8080
go run ./apim-mock-server --auth-file apim-mock-server/examples/auth.yaml
go run ./apim-mock-server --dry-run-reject
```

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | HTTP listen port |
| `--base-path` | `/automation` | API base path |
| `--auth-file` | | Auth YAML (optional). Sample: `apim-mock-server/examples/auth.yaml` |
| `--dry-run-reject` | off | PUT `?dryRun=true` answers `200` with the message in `errors.severe`, and persists nothing. |

It serves every APIM SDK operation, AIM included, at `/automation/organizations/{orgId}/environments/{envId}/…`.
Requests are validated against the merged OpenAPI spec: an invalid body or parameter gets a `400` with the failing
field.

It behaves like APIM:

- **PUT returns the State**: the spec with its OpenAPI defaults plus `id`, `organizationId`, `environmentId` and, on
  sub-resources, `apiHrid`/`portalHrid`. Subscriptions also get `startingAt`, kept across PUTs. GET returns the same
  State.
- **404 on a missing resource**: GET or DELETE of an unknown hrid answers 404 `HRID [<hrid>] cannot be found.`
- **Deterministic ids**: `id` (and `crossId`) are UUIDv5, stable per org, env, kind, parent (for a sub-resource)
  and hrid, like APIM's but not byte-identical. A `crossId` sent on an API or shared policy group PUT is kept, as for
  legacy resources.
- **Dry run returns a computed State**, built as for a real PUT, and persists nothing. The shared policy group DELETE
  also supports `dryRun`: `404` if missing, else `204` with nothing deleted.
- **References**: sub-resources under a missing API (`apiHrid`) or portal (`portalHrid`) answer 404, whatever the
  method. A subscription PUT answers 404 when its application, or its plan on the API, does not exist.
- **Cascades**: deleting an API deletes its subscriptions, documentations and links; deleting a portal deletes its
  listings, documentations and links; deleting an application deletes its subscriptions (APIM closes them instead).
- **`ignoreMembers`** is applied on groups: an update keeps the stored members, a new group has none.
- **Legacy UUID params** (`hridContainsUUID`, `hridContainsApiUUID`, `hridContainsAppUUID`, the `X-Gravitee-Set-Hrid`
  header) and `ignoreGroups` are accepted and ignored: an hrid is always an hrid.
- **Polymorphic fields** (listeners, selectors, the AIM discriminated objects) are stored as unstructured maps, so
  every variant field is kept but no defaults are applied inside them.

## Prerequisites

- Go 1.26+

## Build & Test

```bash
go test ./am-sdk/... ./apim-sdk/... ./am-mock-server/... ./apim-mock-server/... ./common/...
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
