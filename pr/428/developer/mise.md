---
url: /notopia/pr/428/developer/mise.md
---

# Tooling with mise

[mise](https://mise.jdx.dev) pins every CLI tool the repository needs (codegen, Helm, Go, Node,
linters, …) for both local development and CI. Versions are locked with
[`mise.lock`](https://mise.jdx.dev/dev-tools/mise-lock.html) so every machine installs the same
binaries, verified by checksum.

## Layout

Local development and CI need different tools, so the config is split with
[config environments](https://mise.jdx.dev/configuration/environments.html). Each config file
has its own lockfile next to it.

| File            | Lockfile        | Holds                                              | Loaded                           |
| --------------- | --------------- | -------------------------------------------------- | -------------------------------- |
| `mise.toml`     | `mise.lock`     | Tools and env shared by local and CI               | Always                           |
| `mise.dev.toml` | `mise.dev.lock` | Local-only tools and env (pnpm, delve, stripe, …)  | Locally, the default environment |
| `mise.ci.toml`  | `mise.ci.lock`  | CI-only tools (just, codecov, python, uv, …)       | In CI, with `MISE_ENV=ci`        |
| `.miserc.toml`  |                 | Selects `dev` by default, ignores `.tool-versions` | Read before every other file     |
| `.mise/locks/`  |                 | Dependency graphs of `npm:` and `pipx:` tools      | Committed with the lockfiles     |

`MISE_ENV=ci` replaces the `.miserc.toml` default rather than adding to it, so CI never loads
`mise.dev.toml` and local machines never load `mise.ci.toml`.

```sh
mise config ls            # mise.toml + mise.dev.toml
mise -E ci config ls      # mise.toml + mise.ci.toml
```

`.tool-versions` is not read by mise. It only feeds `actions/setup-go` and
`golangci-lint-action` in CI, so keep its `golang` version equal to `go` in `mise.toml`.

Project settings in `mise.toml` that keep the lockfiles reproducible:

* `min_version`: lockfiles written by older mise versions are rejected by newer ones (missing
  checksums), so everyone runs at least the version CI pins in `.github/actions/setup/action.yml`.
* `lockfile = true`: `mise use` and `mise upgrade` keep the lockfiles in sync.
* `[tool_config] locked = true`: installing a project tool without a lock entry fails, but your
  global tools in `~/.config/mise` keep their own policy.
* `npm.package_manager = "aube"`: the `npm:` dependency graphs do not depend on a personal global
  `npm.package_manager` setting.

## First setup

```sh
mise self-update   # if mise reports a too-old version
mise trust
mise install
```

## Adding or changing a tool

Pick the file with `--env`. Without it, `mise use` writes to the shared `mise.toml`.

```sh
mise use helm@4.3.2                  # local and CI   -> mise.toml,     mise.lock
mise use --env dev stripe@latest     # local only     -> mise.dev.toml, mise.dev.lock
mise use --env ci uv@0.13.9          # CI only        -> mise.ci.toml,  mise.ci.lock
```

To move a tool between scopes, move its line by hand, then refresh the lockfiles.

## Refreshing and bumping lockfiles

Lockfiles cover `linux-x64` (CI, WSL) and `macos-arm64`. `mise lock` keeps the platforms that are
already recorded, so pass `--platform` only to add one.

```sh
mise lock && mise -E ci lock               # refresh after editing a toml by hand
mise lock --bump && mise -E ci lock --bump # move `latest` / prefix requests to newer versions
mise install
```

`--bump` does not change exact pins such as `helm = "4.3.2"`; use `mise upgrade --bump helm` for
those. Running `mise -E ci lock` installs nothing, it only resolves and writes `mise.ci.lock`.

## Before committing

```sh
mise lock --check && mise -E ci lock --check
```

Commit the changed `mise*.toml`, `mise*.lock` and `.mise/locks/` files together. Personal
overrides go to `mise.local.toml` or `mise.dev.local.toml`, which are gitignored with their
lockfiles.

## CI

`.github/workflows/ci.yaml` sets `MISE_ENV=ci`. The setup action pins the mise version, runs
`jdx/mise-action` (which installs with `mise install --locked` because `mise.lock` exists) and
then checks that all lockfiles are up to date:

```sh
mise lock --check && mise -E dev lock --check
```

A pull request that edits a `mise*.toml` without refreshing its lockfile fails that step. Changes
to only `mise.dev.toml`, `mise.dev.lock` or `.mise/locks/mise.dev/` do not trigger CI.

## Upgrading mise

Bump `min_version` in `mise.toml` and `version` of `jdx/mise-action` in
`.github/actions/setup/action.yml` together, then refresh the lockfiles with the new version.

## Renovate

Renovate's mise manager reads `mise.toml`, `mise.dev.toml` and `mise.ci.toml` and updates the
matching lockfiles and sidecars in the same pull request. If an update arrives without its
lockfile, the CI lockfile check fails; run the refresh commands above on that branch.
