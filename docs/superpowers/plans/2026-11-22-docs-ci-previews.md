# Docs CI Previews Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Validate affected documentation on pull requests and publish safe, working previews while retaining the existing two-job CI workflow.

**Architecture:** The existing `main` job detects whether the combined `page` project is affected and supplies an output to the existing `docs` job. The docs job builds with Kroki only when needed, publishes previews for trusted same-repository PRs and production from main, and serializes writes to `gh-pages`.

**Tech Stack:** GitHub Actions, Nx, `nx-github-pages` 1.3.1, VitePress, Vite, Kroki, pnpm.

**Spec:** Agreed conversation: retain two jobs, docs waits for main, no Kroki for unaffected changes, affected PR validation, trusted PR previews, production deployment from main, manual rebuild override. Preview cleanup is included as a follow-up using the existing two jobs.

## Global Constraints

- Work in the main worktree and preserve any unrelated user changes.
- Integrate after PR #185 or base implementation on its `nx-github-pages` 1.3.1 dependency changes; do not duplicate or overwrite the Renovate PR.
- Keep exactly two jobs in `.github/workflows/ci.yaml`: `main` and `docs`. No separate affected-detection or publishing job.
- Never use `pull_request_target` to run PR code or give fork PR code a write token.
- Preserve `workflow_dispatch` controls, existing diagram caching, and production `/notopia/` URLs.
- Do not commit unless the user separately requests it.

## Review Focus

- Fork PRs and Dependabot-style restricted tokens: builds work without write credentials; deployment is skipped.
- Production and preview builds with identical source: Nx cannot reuse output built for a different base path.
- Parallel PR, main and cleanup runs: all writes to `gh-pages` serialize, without cancelling an in-flight publisher.
- Production redeployment: existing previews survive and old production files do not remain unexpectedly.
- API docs, diagrams, search and deep links: all assets resolve under the correct preview prefix.

## File Map

- `.github/workflows/ci.yaml`: affected detection, PR validation, trusted publishing, serialization and cleanup scheduling.
- `page/project.json`: non-cacheable deployment and preview/cleanup executor configuration.
- `docs/.vitepress/config.mts`, `docs/project.json`: configurable VitePress/diagram base and Nx cache inputs.
- `apps/api-web/vite.config.mts`, `apps/api-web/project.json`: configurable API base and Nx cache inputs.
- `docs/.vitepress/*.spec.ts` and any appropriate workflow test script: regression coverage following existing patterns.

### Task 1: Make combined page builds base-aware

**Files:** Modify `docs/.vitepress/config.mts`, `docs/project.json`, `apps/api-web/vite.config.mts`, `apps/api-web/project.json`; extend existing docs config tests.

**Interfaces:** `DOCS_BASE` is the full site base with leading and trailing slash, default `/notopia/`; API base is `${DOCS_BASE}api/`. Preview uses `/notopia/pr/<PR_NUMBER>/`. Both build targets include `{ "env": "DOCS_BASE" }` in their Nx inputs without dropping existing inputs.

- [ ] Add focused tests asserting production and PR 123 base values for VitePress, diagram URLs, API assets and the Scalar API link. Use existing Vitest patterns; extract a small shared base resolver only if necessary for testability.
- [ ] Run the focused tests and confirm the preview assertions fail against the current hard-coded paths.
- [ ] Configure VitePress `base`, diagram `publicPath` and dead-link handling from `DOCS_BASE`; preserve `/notopia/` as the default. Remove or update the production target's existing `DOCS_BASE` override so it cannot override the preview value.
- [ ] Configure Vite's API base from the same environment variable. Ensure the environment survives Nx's dependency builds and that the relevant `build` targets hash it; verify the inferred API-web build configuration before editing its inputs.
- [ ] Run the focused tests and `nx build page --configuration production` with Kroki available, once with `DOCS_BASE=/notopia/` and once with `DOCS_BASE=/notopia/pr/123/`; confirm generated HTML, diagram references, API JS assets, search and links use the matching prefix. Serve the built output at each prefix and request a nested docs page and `/api/`.
- [ ] Repeat both builds with Nx cache enabled in alternating order; verify the prefix changes each time and no stale output is restored.

### Task 2: Validate affected docs in the existing two jobs

**Files:** Modify `.github/workflows/ci.yaml`.

**Interfaces:** `needs.main.outputs.page-affected` is `true` or `false` for PR, main and manual runs; manual `nxAllCmd=run-many` forces the docs job. The docs job runs only after successful `main`.

- [ ] Add a workflow-condition regression check covering: affected PR runs docs; unaffected PR skips docs; affected main runs docs; unaffected main skips docs; manual `run-many` runs docs. Assert a failed `main` does not deploy.
- [ ] Run the check against the existing workflow; confirm the affected PR case fails.
- [ ] Run `nx show projects --affected --with-target deployPage --json` after setup for PRs as well as main. Parse the JSON and set the existing job output; keep the Nx base/head values supplied by `.github/actions/setup`.
- [ ] Update the docs job condition for PRs, main and the manual override; preserve `main`-only production publishing and the existing diagram cache and controls. Build on fork and restricted-token PRs without attempting a deploy. Avoid duplicate publishing on branch `push` events.
- [ ] Run the workflow-condition regression check and a YAML/workflow validation tool available in the repository. Confirm affected detection includes docs, API spec/API-web and `page` configuration changes using representative base/head commits; document any Nx graph gap and correct its dependency/inputs.

### Task 3: Publish previews and preserve production

**Files:** Modify `page/project.json` and `.github/workflows/ci.yaml`.

**Interfaces:** `page:deployPage` is not cached and has preview enabled, URL `https://notopia-uit.github.io/notopia`, prefix `pr/{prNumber}`, and `syncWithBaseBranch: true`; `GH_TOKEN` is available only to the trusted publishing step. Docs job concurrency uses one shared non-cancelling group for all `gh-pages` writers.

- [ ] Add regression checks for deployment conditions: same-repository non-Dependabot PR publishes a preview; fork and Dependabot PRs validate only; `main` push publishes production; manual main rebuild publishes production; branch pushes do not publish. Assert PR number and base prefix correspond.
- [ ] Run the checks and confirm preview publishing is not currently configured.
- [ ] Set `cache: false` on `deployPage`, add the plugin's preview and sync options, and configure the docs workflow's explicit build and deploy steps. Use `/notopia/pr/${{ github.event.pull_request.number }}/` for trusted PR builds and `/notopia/` for production; pass the same `DOCS_BASE` through build dependencies and deployment.
- [ ] Grant the docs job `contents: write` and `pull-requests: write` only as required by publishing; gate publication on a same-repository PR and non-Dependabot actor, or a main push/manual run. Ensure checkout and build steps do not receive a write token for untrusted PRs; account for the workflow's current global `GITHUB_TOKEN` environment and checkout credential persistence.
- [ ] Add a single `gh-pages` job-level concurrency group with `cancel-in-progress: false`. Ensure production, preview and the cleanup in Task 4 share it. Review other workflows for writers to the same branch.
- [ ] In a disposable repository or isolated test remote, deploy production, deploy two previews, redeploy production and check both preview directories remain. Remove a production source page and check it no longer appears after redeploy. `nx-github-pages` 1.3.1 performs a force push after merging; if its merge strategy fails either preservation or stale-file removal, implement a safe clone/update/push strategy before enabling production publishing.
- [ ] On a test PR, verify preview comment URL, rendered nested page, diagrams, search and `/api/`; verify production remains intact. Exercise an untrusted or restricted-token PR and confirm build succeeds with no publish attempt.

### Task 4: Clean up previews without adding a CI job

**Files:** Modify `page/project.json`, `.github/workflows/ci.yaml`.

**Interfaces:** Add `page:cleanup-previews` using `nx-github-pages:cleanup-preview`, prefix `pr`, `maxAgeDays: 7`, and comment updates enabled. Scheduled cleanup uses the same `gh-pages` concurrency group; no PR code is checked out.

- [ ] Add a scheduled trigger and, if desired, a manual cleanup selector to the existing workflow. Route the schedule through the existing jobs and run cleanup in `docs` on the default branch, skipping ordinary Nx validation/deploy steps on cleanup-only runs. Retain the existing manual CI controls.
- [ ] Document that the plugin's age-based policy can remove previews for still-open PRs; a new PR update recreates them. Verify whether the executor uses file or commit timestamps and that the selected behavior matches this policy before enabling the schedule.
- [ ] In a disposable remote, create old and recent preview directories and corresponding test PR comments. Run cleanup and confirm old previews are removed, recent previews and production are intact, and affected comments are updated. Confirm cleanup does not race with a deploy.
- [ ] Validate workflow syntax and review the event/permissions matrix, including schedule, fork PR, Dependabot PR, same-repo PR, main push and manual dispatch. Record any GitHub-only checks that could not be executed locally.
