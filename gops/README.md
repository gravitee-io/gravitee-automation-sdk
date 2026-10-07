# gops

Gravitee automation from the command line, built on the Automation SDKs. This proof of concept
ships one command: `export`.

```bash
gops export --format=crd --resources=apis,mcp-proxies --config=gops.yaml --output=./gravitee
```

`gops` lists each requested kind through the Automation API and writes one file per resource
under `<output>/<kind>/<name>.<ext>`.

| `--format` | Writes |
|---|---|
| `json`, `yaml` | The state the Automation API returns, as the SDK models serialize it |
| `crd` | The GKO manifest of the matching kind |
| `tf` | Reserved: the Terraform provider's resource block (not implemented, exit 2) |

| `--resources` | Automation API | GKO kind |
|---|---|---|
| `apis` | `GET /apis` (`ApiV4State`) | `ApiV4Definition` |
| `mcp-proxies` | `GET /aim/mcp-proxies` (`AimMcpProxyState`) | `McpProxy` |

## Configuration

`gops.yaml` (see [gops.yaml.example](gops.yaml.example)) holds the API URL, organization,
environment and credentials. `GOPS_TOKEN` overrides the token so CI never writes a secret to disk.

## Humans and agents

Every input a prompt can gather has a flag. In a terminal, `gops` asks for what you did not type;
without one (CI, an agent, `--no-input`), a missing input fails with exit 2 and the flag named.

| Situation | Terminal | No terminal or `--no-input` |
|---|---|---|
| `--format` missing | select | exit 2 |
| `--resources` missing | multi-select | exit 2 |
| two resources share a name | one input per resource (`--on-collision=prompt`) | exit 2 with the collisions; `--on-collision=suffix` or `--rename <id>=<name>` settles them |
| `--format=crd` without `--ids` | confirm: keep or strip the APIM identifiers | `keep` |
| a file exists, no `--force` | confirm | exit 2 naming the file |
| report | one line per file | `--json`: the report, or the collisions, on stdout |

Exit codes: `0` done, `1` reserved (a failed gate, `gops score`), `2` the run failed or the input
was incomplete.

## Naming

Resources created in the Console have no HRID, so the object name (and file name) is the resource
name as a Kubernetes label: lower case, dashes, 63 characters at most. "Petstore API (v2)" becomes
`petstore-api-v2`. Names are not unique in APIM: two "Petstore" APIs are a collision, settled by
the person who knows which is which, or deterministically by `--on-collision=suffix`
(`petstore`, `petstore-2`).

## CRD mapping

The mapping is driven by the two schemas rather than by per-kind code:

1. Properties the OpenAPI state schema marks `readOnly` are status (`id`, `environmentId`,
   `organizationId`, `crossId`, `errors`) and are dropped.
2. Properties absent from the CRD schema are dropped and reported (`primaryOwner`).
3. A list the CRD keys by name becomes a map (`plans`, `pages`), keyed by each item's name.
4. CRD nodes with `x-kubernetes-preserve-unknown-fields` (`listeners`, flow selectors) pass through.
5. `metadata.name` is the settled object name; an empty `hrid` is dropped, GKO derives it from
   the namespace and name.
6. The identifiers nested in the manifest (`plans.*.id`, `pages.*.id`) are a choice, not a rule:
   kept (`--ids keep`, the default), GKO adopts the existing plans and pages when the manifest is
   applied in the same environment; stripped (`--ids strip`), the manifest moves to another
   environment and GKO creates them. In a terminal, `gops` asks.

The CRDs are copied from gravitee-kubernetes-operator (`internal/transform/crd/manifests`); the
OpenAPI document is the one `apim-sdk` embeds.

## Known limits

- A manifest applied by GKO gets `hrid = <namespace>-<name>`, which is not the exported
  resource's identity: GKO creates a sibling rather than adopting the Console-created one.
- No `contextRef` is written.
- Credentials come back masked by the API and are exported as such.
- A required field the API returned as `null` (`hrid`) is serialized as `""` by the SDK model.
