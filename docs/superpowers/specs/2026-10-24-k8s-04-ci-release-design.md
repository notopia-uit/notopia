# K8s 04 — CI and release

> Sub-project 4 of 4. See `2026-10-24-k8s-00-overview-design.md`.
> Can run in parallel with 02. The tag bump step needs 03's `clusters/homelab/apps/*/values.yaml`.

## Goal

Release every deployable service (adding the missing authorization and billing). After pushing
images, the release workflow writes the new image tags into the cluster values, so Flux deploys
immediately. Chart changes are linted, tested and version-checked. Release stays manual, shaped
so that switching to automatic on push to `main` is a trigger change.

## Current state (verified)

- `.github/workflows/ci.yaml`: one `main` job (verify + release steps gated by `SHOULD_RELEASE`)
  plus a `docs` job (Kroki sidecar). Setup steps are duplicated between them.
- Dispatch inputs: `releaseWeb`, `releaseNote`, `releaseDocument`, `releaseSearchWorker`,
  `releaseNpm`, `releaseAll`, `releaseContainerType`. There are **no inputs for authorization or
  billing**.
- `nx.json` release group `docker-go`: `note, authorization, document, search-worker, web`.
  **billing is missing.** Tag pattern `{projectName}@{version}`, version schemes
  `production: {currentDate|YYMM.DD.HHmm}`, `development: dev-{currentDate|YYMM.DD.HHmm}`.
- `paths-ignore` contains `deploy/**/*`, so chart-only changes would skip CI.
- `.mise.ci.toml` is separate from `mise.toml` (no `helm`, no `yq`).

## Decisions

- **Keep one `main` job** for verify and release (shared checkout, setup and Nx cache). This is
  an accepted trade-off: same-repo PR runs get the job's write permissions. Fork PRs always get a
  read-only token.
- **Shared setup** in a composite action, used by `main` and `docs`.
- **Tags are bumped by the release workflow** (`yq`), not by Flux image automation or Renovate.
  Renovate keeps handling third-party versions.

## Changes

### 1. Composite action `.github/actions/setup/action.yml`

Inputs: `cache` (bool), `go` (bool, default `true`), `golangci` (bool, default `false`).
Steps, as they are today: GitHub runtime, checkout (`fetch-depth: 0`, `filter: tree:0`), pnpm,
node (`.tool-versions`), Go (optional), golangci-lint install (optional), mise
(`MISE_OVERRIDE_CONFIG_FILENAMES` from the workflow env), buf cache, `nrwl/nx-set-shas`,
`pnpm install`, `go mod download` (optional).

`main` uses it with `golangci: true`. `docs` uses it with `go: true` (Kroki-only steps stay in the
job).

### 2. Paths

`paths-ignore`: `deploy/**/*` → `deploy/compose/**/*`. Chart projects (02) are Nx projects, so
`nx affected` picks up `lint`, `test` and `gen` for them.

### 3. Chart version check

`scripts/check-chart-versions.sh`, run in `main` after `Nx all`:

- for each `deploy/k8s/charts/<chart>/` changed against the base SHA (from `nx-set-shas`),
  `Chart.yaml` `version` must differ from the base and be a higher semver
- if `notopia-common` changed, every chart depending on it must also be bumped
- generated files (`README.md`, `values.schema.json`) count as changes. They only change when
  `values.yaml` does.

It fails with a list of the charts needing a bump.

### 4. Release group and inputs

- `nx.json`: rename group `docker-go` → `docker` (it also holds Node images; the tag pattern
  doesn't include the group name, so existing tags are unaffected). Projects:
  `web, note, authorization, billing, document, search-worker`.
- Each released project needs a `docker:build` target (verify authorization and billing have
  one, and add it if missing) and its `Dockerfile` (`cmd/authorization/Dockerfile`,
  `cmd/billing/Dockerfile` exist).
- New dispatch inputs: `releaseAuthorization`, `releaseBilling`.

### 5. Release steps (in `main`, gated by `SHOULD_RELEASE`)

```
Resolve release projects   → outputs RELEASE_PROJECTS (comma list)
                              workflow_dispatch: from the boolean inputs (or all, for releaseAll)
                              (future) push to main: nx show projects --affected --with-target docker:build
Docker build               → nx run-many --projects $RELEASE_PROJECTS --target docker:build
Nx release                 → nx release --projects $RELEASE_PROJECTS --dockerVersionScheme … --yes
Bump deploy tags           → for each project: version = the tag "<project>@<version>" created at HEAD
                              yq -i '.image.tag = strenv(VERSION)' deploy/k8s/clusters/homelab/apps/<project>/values.yaml
                              (skip a project with no values file and log a warning)
Commit + push              → one commit: "chore(deploy): release <p>@<v>, …", push to main
```

- The existing `--projects` flag building (`${{ … && '--projects web' }}` per input) is replaced
  by the single `RELEASE_PROJECTS` output. The npm group keeps its own flag (`--groups npm`).
- The push uses `GITHUB_TOKEN`. Its commits don't start workflows (so there's no loop), but the
  repo webhook still fires, so Flux reconciles. If `main` gains branch protection, switch to the
  bot app token (the commented `NOTOPIA_BOT` step).
- `yq` is added to `.mise.ci.toml` (pinned), not taken from the runner image.

### 6. Renovate

`.renovaterc.json`:

- Disable updates for `ghcr.io/notopia-uit/**` image tags under `deploy/k8s/**` (release-owned).
- Enable the `flux` and `helm-values` managers for `deploy/k8s/clusters/**` (Valkey and
  Meilisearch chart versions and images).

### 7. mise

`mise.toml` and `.mise.ci.toml` get the chart tools from 02 (`helm`, `helm-schema`, `helm-docs`,
`kubeconform`) and `yq`. `helm-unittest` is installed as a Helm plugin in the setup action.

## Switching to automatic release later

1. Add a `push: branches: [main]` condition to `SHOULD_RELEASE` (event `push`, ref `main`, not
   Renovate).
2. The "Resolve release projects" step uses its `nx show projects --affected` branch.

Nothing else changes.

## Testing

- Dispatch with `releaseBilling: true` on a test branch, with the push step disabled through an
  `if:` on `main`. Images are built, and the yq diff shows only `billing/values.yaml`.
- Change a chart template without bumping its version: CI fails, naming the chart.
- A docs-only change still runs `docs` with the composite action.
