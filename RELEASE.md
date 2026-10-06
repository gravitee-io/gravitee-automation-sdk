# Release Process

This repository uses [release-please](https://github.com/googleapis/release-please) for automated releases with independent per-module versioning.

## How It Works

1. **Commit to `main`** using [conventional commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, etc.)
2. **release-please** automatically creates or updates a release PR per module, with a changelog and version bump
3. **Merge the release PR** to cut the release — release-please creates the git tag and GitHub Release

### Version Bumps

| Commit type | Version bump |
|---|---|
| `fix:` | Patch (0.1.0 → 0.1.1) |
| `feat:` | Minor (0.1.0 → 0.2.0) |
| `feat!:` or `BREAKING CHANGE:` footer | Minor while < 1.0, Major after |

A major bump to v2+ also needs the `/vN` suffix on the module path (`go mod edit -module .../am-sdk/v2`, imports and dependents' `go.mod` updated) in the same PR — Go refuses a `v2.x.y` tag on a module path without it.

## Tag Format

Each module gets its own semver tag: `am-sdk/v0.2.0`, `common/v0.1.3`, `am-mock-server/v0.3.0`, etc. This is the standard Go multi-module convention — consumers use `go get github.com/gravitee-io/gravitee-automation-sdk/am-sdk@v0.2.0`.

## Module Types

### Libraries (`am-sdk`, `apim-sdk`, `common`)

Release artifact is the git tag + GitHub Release with changelog. No binaries — Go resolves the source from the module proxy.

### Binaries (`am-mock-server`)

Same as libraries, plus [goreleaser](https://goreleaser.com/) cross-compiles binaries (darwin/linux × amd64/arm64) and uploads them to the GitHub Release.

## Auto-bump Dependencies

When `common` or `am-sdk` is released, a GitHub Actions workflow automatically:

1. Updates `go.mod` in dependent modules (`go mod edit -require` + `go mod tidy`)
2. Opens a PR with `fix(deps): update <module> to <version>`
3. That PR, once merged, triggers release-please for the dependent modules

## Configuration

| File | Purpose |
|---|---|
| `release-please-config.json` | Per-module release config (release type, component names) |
| `.release-please-manifest.json` | Current version of each module |
| `am-mock-server/.goreleaser.yml` | Cross-compilation config for binary releases |

## First Release

All modules start at `0.0.0` in the manifest. The first `feat:` commit touching a module will produce a `v0.1.0` release PR.

**Important:** Release `common` first, then `am-sdk`, then `am-mock-server` — so that dependency versions resolve correctly on the Go module proxy.

## CI Checks (on every PR)

| Job | What it does |
|---|---|
| `generate` | Runs `task generate`, fails if generated files have uncommitted changes |
| `lint` | `go vet` + `staticcheck` + `revive` + license header check |
| `test` | `go test ./...` per module |
