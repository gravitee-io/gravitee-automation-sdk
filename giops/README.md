# giops

`giops` is a command line client for the Gravitee Automation APIs of Access Management (AM) and
API Management (APIM). It reads, applies and deletes resources described in YAML or JSON files.

> **Experimental.** giops is a proof of concept: commands, flags and file formats may change.

## Install

Download the archive for your platform from the `giops/vX.Y.Z` releases of this repository, or build it
from a clone:

```bash
go build -o giops ./giops
```

`giops --version` prints the release.

## Configuration

giops connects to one or both products, called modules: `am` and `apim`. Each module has its own section
in the configuration, `am` for AM and `gamma` for APIM. Declare at least one.

### File

giops reads `./config.yaml`, or the file given with `-c` / `--config`. YAML (`.yaml`, `.yml`) and JSON
(`.json`) are accepted.

```yaml
am:
  base_url: http://localhost:8093/automation
  org_id: DEFAULT
  env_id: DEFAULT
  auth:
    bearer_token: my-am-token

gamma:
  base_url: http://localhost:8083/automation
  auth:
    basic_auth:
      username: admin
      password: admin
```

| Key | Required | Meaning |
|-----|----------|---------|
| `base_url` | yes, except APIM with `cloud_token` | Automation API root, without organization and environment |
| `org_id` | no | Organization; `DEFAULT` when left out |
| `env_id` | no | Environment; `DEFAULT` when left out |
| `auth.bearer_token` | one of the two | Token sent as `Authorization: Bearer` |
| `auth.basic_auth.username`, `auth.basic_auth.password` | one of the two | Credentials sent as HTTP Basic |
| `cloud_token` | no, APIM only | `true` when `bearer_token` is a Gravitee Cloud token: org and env come from the token, and an empty `base_url` becomes the token's cloud gate |

[examples/config/config.yaml](examples/config/config.yaml) is a commented example with every key.

### Environment variables

Every key can also be set with an environment variable, read from the shell or from a `.env` file in the
current directory. A variable overrides the file. Variables alone can declare a module the file leaves
out.

| Variable | Key |
|----------|-----|
| `GIOPS_AM_BASE_URL` | `am.base_url` |
| `GIOPS_AM_ORG_ID` | `am.org_id` |
| `GIOPS_AM_ENV_ID` | `am.env_id` |
| `GIOPS_AM_AUTH_BEARER_TOKEN` | `am.auth.bearer_token` |
| `GIOPS_AM_AUTH_BASIC_USERNAME` | `am.auth.basic_auth.username` |
| `GIOPS_AM_AUTH_BASIC_PASSWORD` | `am.auth.basic_auth.password` |
| `GIOPS_GAMMA_BASE_URL` | `gamma.base_url` |
| `GIOPS_GAMMA_ORG_ID` | `gamma.org_id` |
| `GIOPS_GAMMA_ENV_ID` | `gamma.env_id` |
| `GIOPS_GAMMA_AUTH_BEARER_TOKEN` | `gamma.auth.bearer_token` |
| `GIOPS_GAMMA_AUTH_BASIC_USERNAME` | `gamma.auth.basic_auth.username` |
| `GIOPS_GAMMA_AUTH_BASIC_PASSWORD` | `gamma.auth.basic_auth.password` |
| `GIOPS_GAMMA_CLOUD_TOKEN` | `gamma.cloud_token` |

For example, to keep the token out of the file:

```bash
export GIOPS_AM_AUTH_BEARER_TOKEN=my-am-token
giops am check
```

## Commands

Commands read `giops <module> <verb> <resource>`:

```bash
giops am get domains
giops apim apply groups -f group.yaml
```

### check

`giops <module> check` connects to the module and tells whether the configuration works: the server is
reachable and accepts the credentials, organization and environment.

### get

`giops <module> get <resource>` lists the resources, where the API can list them (AM only);
`giops <module> get <resource> <id>` reads one.

| Flag | Meaning |
|------|---------|
| `-o`, `--output` | `yaml` or `json` instead of a table. The YAML and JSON output can be applied back |

### apply

`giops <module> apply <resource> -f <file>` creates the resource, or updates it when it exists. The file
holds one resource, in YAML or JSON, with the same field names as the API.

By default, apply first runs a dry run: the server validates the resource without saving it, and giops
prints its errors and warnings. Errors stop there. Otherwise giops asks before applying; any answer other
than `y` or `yes` applies nothing.

| Flag | Meaning |
|------|---------|
| `-f`, `--file` | File of the resource to apply (required) |
| `-d`, `--dry-run` | Stop after the dry run |
| `-y`, `--yes` | Apply at once, without the dry run or the question |

### delete

`giops <module> delete <resource> <id>` deletes the resource.

Like apply, delete asks before deleting; any answer other than `y` or `yes` deletes nothing. When the API
has a delete dry run for the resource, giops runs it first and prints its result; errors stop there. These
resources have one: APIM `shared-policy-groups`.

| Flag | Meaning |
|------|---------|
| `-i`, `--id` | Id of the resource to delete, instead of the argument |
| `-f`, `--file` | File of the resource to delete: its id is read from the file |
| `-d`, `--dry-run` | Stop after the dry run, or before the question when there is no dry run |
| `-y`, `--yes` | Delete at once, without the dry run or the question |

## Resources

Parents are the resources a resource lives in; see [Parents](#parents). Only AM resources can be listed:
the APIM Automation API has no list endpoint, so APIM `get` needs an id.

| Module | Resource | Id | Parents | List | Example |
|--------|----------|----|---------|------|---------|
| `am` | `domains` | `key` | — | yes | [examples/am/domain.yaml](examples/am/domain.yaml) |
| `am` | `identityproviders` | `key` | `domain` | yes | [examples/am/identity-provider.yaml](examples/am/identity-provider.yaml) |
| `am` | `reporters` | `key` | `domain` | yes | [examples/am/reporter.yaml](examples/am/reporter.yaml) |
| `am` | `certificates` | `key` | `domain` | yes | [examples/am/certificate.yaml](examples/am/certificate.yaml) |
| `am` | `applications` | `key` | `domain` | yes | [examples/am/application.yaml](examples/am/application.yaml) |
| `am` | `secrets` | `key` | `application`, `domain` | yes | [examples/am/secret.yaml](examples/am/secret.yaml) |
| `am` | `dataplanes` | `id` | — | yes | [examples/am/dataplane.yaml](examples/am/dataplane.yaml) |
| `apim` | `groups` | `hrid` | — | no | [examples/apim/group.yaml](examples/apim/group.yaml) |
| `apim` | `apis` | `hrid` | — | no | [examples/apim/api.yaml](examples/apim/api.yaml) |
| `apim` | `documentations` | `hrid` | `api` | no | [examples/apim/api-documentation.yaml](examples/apim/api-documentation.yaml) |
| `apim` | `links` | `hrid` | `api` | no | [examples/apim/api-link.yaml](examples/apim/api-link.yaml) |
| `apim` | `applications` | `hrid` | — | no | [examples/apim/application.yaml](examples/apim/application.yaml) |
| `apim` | `subscriptions` | `hrid` | `api` | no | [examples/apim/subscription.yaml](examples/apim/subscription.yaml) |
| `apim` | `dictionaries` | `hrid` | — | no | [examples/apim/dictionary.yaml](examples/apim/dictionary.yaml) |
| `apim` | `shared-policy-groups` | `hrid` | — | no | [examples/apim/shared-policy-group.yaml](examples/apim/shared-policy-group.yaml) |

```bash
giops am apply domains -f examples/am/domain.yaml
giops am get domains
giops am get domains example-domain -o yaml
giops am delete domains example-domain
```

## Resources with additional options

### Parents

A child resource lives in its parents: an AM identity provider, reporter, certificate or application in a
security domain; an AM client secret in an application of a domain; an APIM API documentation, link or
subscription in an API. Name each parent's id with a field in the file (`domain`, `application`, `api`), or
with the flag of the same name on `get`, `apply` and `delete`. The flag wins over the file.

```bash
giops am apply identityproviders -f examples/am/identity-provider.yaml
giops am get identityproviders --domain example-domain
giops am apply secrets -f examples/am/secret.yaml --application other-app   # into other-app, not the file's
giops am get secrets example-secret --domain example-domain --application example-app
giops apim delete subscriptions example-subscription --api example-api
```

### APIM apis and applications: `--ignore-groups`

`apply apis -g` / `--ignore-groups`, and the same on `applications`, saves the resource without its
`groups`: the groups already in APIM are left as they are. It applies to the dry run too.

```bash
giops apim apply apis -f examples/apim/api.yaml --ignore-groups
```

### APIM groups: `--ignore-members`

`apply groups -m` / `--ignore-members` saves the group without its `members`: the members already in APIM
are left as they are. It applies to the dry run too.

```bash
giops apim apply groups -f examples/apim/group.yaml --ignore-members
```

### APIM subscriptions and AM data planes: no apply dry run

Their API has no dry run. `apply` prints "Dry run is not configured", then asks before applying;
`--dry-run` stops there and saves nothing, `--yes` applies at once.

### AM client secrets: write-only value

The API never returns a secret's `value`, and needs it on every apply: re-applying the stored value does
nothing, another value is refused. So `get secrets -o yaml` cannot be applied back. A secret is immutable:
to rotate it, apply a secret with a new key, then delete the old one.

### AM check: `--print-result`

`giops am check -p` / `--print-result` also prints, as JSON, the domains the check read.
