# Docs Publishing and REST Reference Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish one cacheable docs artifact containing the VitePress site, split REST reference, and existing Scalar application.

**Architecture:** Nx builds the joined API spec and Scalar before VitePress; VitePress assembles the final artifact in `buildEnd`. The joined OpenAPI spec drives dynamic endpoint pages and their Markdown, while the general LLM plugin owns the site-wide indexes.

**Tech Stack:** Nx, pnpm, VitePress 2 alpha, Scalar/Vite React, `vitepress-openapi`, `vitepress-plugin-llms`, Pagefind, GitHub Actions.

**Spec:** [Docs publishing and REST reference](../specs/2026-11-28-docs-api-reference-design.md)

## Upstream references

- [`vitepress-openapi` repository](https://github.com/enzonotario/vitepress-openapi) — [dynamic operation pages](https://vitepress-openapi.vercel.app/pages/by-operation), [pages by tag](https://vitepress-openapi.vercel.app/pages/by-tag), and [Markdown & `llms.txt`](https://github.com/enzonotario/vitepress-openapi/blob/main/docs/pages/markdown.md), including the note on running both LLM generators. The repo currently pins `0.2.21`; check behavior against that installed version.
- [`vitepress-plugin-llms` repository](https://github.com/okineadev/vitepress-plugin-llms) — [setup and output examples](https://github.com/okineadev/vitepress-plugin-llms#usage) and [configuration types](https://github.com/okineadev/vitepress-plugin-llms/blob/main/src/types.d.ts), including `ignoreFilesPerOutput`. Verify the configuration against the planned `1.15.0` release rather than assuming upstream `main` matches it.
- [`vitepress-sidebar` repository](https://github.com/jooy2/vitepress-sidebar) and [VitePress native sidebar config](https://vitepress.dev/reference/default-theme-sidebar#multiple-sidebars) — use alongside the installed `withSidebar` behavior and the composition check in Task 2.

## Global Constraints

- VitePress owns `/api/` and `/api/index.html`; Scalar publishes at `/api/openapi-scalar/`.
- `DOCS_BASE` defaults to `/notopia/`; PR deployments use `/notopia/pr/{prNumber}/`. Cache keys distinguish the bases.
- No generated files in `docs/src/public`, no nested Nx invocations from VitePress, and no Scalar dependency for `nx dev docs`.
- No OpenAPI contract/codegen changes, hand-maintained endpoint list, or standalone schema catalog.
- Retain the default sidebar outside the REST reference and existing gRPC, Pagefind, diagram, authorship, and standalone Scalar **dev** behavior. Standalone preview serves the built artifact at its new published base.
- Prefer existing hooks and native configuration over custom glue. Reuse one spec-derived tag list for the REST overview and sidebar. Keep sidebar composition in the existing config; no new sidebar plugin, theme override, or parallel navigation tree.

## Review Focus

- Missing Scalar `index.html` must fail the build instead of publishing a broken nav link (Task 1).
- A valid operation listed under multiple tags has one canonical URL and remains navigable from each tag (Task 2).
- A raw Markdown request and a Vite module request to the same endpoint path must not get the same response (Task 3).
- A cache restore after a build using a different base must return the complete, correctly prefixed artifact (Task 4).
- Moving the Nx project must not change publishing permissions, PR guards, or cleanup policy (Task 4).

---

### Task 1: Produce the combined docs artifact

**Files:** `docs/project.json`, `docs/.vitepress/config.mts`, `docs/.vitepress/base.spec.ts`, new `docs/.vitepress/scalar-output.ts` and `scalar-output.spec.ts`, `apps/api-web/project.json`, `apps/api-web/vite.config.mts`, `docs/src/api/index.md`.

**Interfaces:** `copyScalarOutput(sourceDir: string, siteOutDir: string): Promise<void>` copies to `siteOutDir/api/openapi-scalar/`. VitePress `buildEnd` receives `siteConfig.outDir`. `api-web:build` produces `apps/api-web/dist/`; `docs:build` produces `docs/.vitepress/dist/`.

- [ ] Add `copies Scalar below the VitePress API landing page` and `rejects a missing Scalar entry point`: in a temporary fixture, assert nested Scalar HTML and assets are copied, the preexisting `api/index.html` is unchanged, and a missing source entry rejects with a message naming `api-web:build`. Update `base.spec.ts` to assert the production and PR Scalar nav, dead-link entry and asset bases; also pin the existing standalone dev base.
- [ ] Run `pnpm --dir docs exec vitest run` and confirm the updated assertions fail before implementing.
- [ ] Implement `copyScalarOutput`, wire it into VitePress `buildEnd`, and update the landing page and navbar links. Resolve the source path from the config module, not the current working directory. Keep the dead-link exception scoped to `${base}api/openapi-scalar/index.html`. In `apps/api-web/vite.config.mts`, use `${DOCS_BASE || '/notopia/'}api/openapi-scalar/` for build/preview while retaining the existing base for standalone dev.
- [ ] Wire explicit Nx dependencies: `api-web:build` depends on `api:join`; `docs:build` depends on both, and `docs` has a project dependency on `api-web`. Use `dependentTasksOutputFiles` for the joined spec and Scalar built files where consumed; keep both build caches sensitive to `DOCS_BASE` and relevant dependency versions. Include VitePress helper files and dynamic route loaders in docs named inputs. `docs:dev` orders only `api:join`; after editing the API spec, restart `nx dev docs` to rebuild the bundle and refresh routes and cached LLM Markdown. `docs:preview` depends on `docs:build`.
- [ ] Inspect `nx show project docs --json`, `nx show project api-web --json`, and `nx build docs --graph=stdout`; the graph must enforce `api:join` before `api-web:build` before `docs:build`. In a disposable checkout with the joined bundle and Scalar output absent, run the docs build with `--skipNxCache` to prove the ordering.
- [ ] Run the docs Vitest suite, then `DOCS_BASE=/notopia/ nx build docs --configuration production` with Kroki available. Assert both `api/index.html` and `api/openapi-scalar/index.html` exist in the docs output and the latter's asset URLs use `/notopia/api/openapi-scalar/`. Run `nx run-many -t lint -p docs api-web`.
- [ ] Commit as `build(docs): include Scalar in the VitePress artifact`.

### Task 2: Generate endpoint pages and tag navigation

**Files:** `docs/.vitepress/config.mts`, new `docs/.vitepress/rest-reference.ts` and `rest-reference.spec.ts`, `docs/src/api/rest.md`, new `docs/src/api/rest/operations/[operationId].md` and `[operationId].paths.ts` in the same directory.

**Interfaces:** `getRestReference(spec: typeof notopiaSpec): RestReference` takes a spec of the shape exported by `@notopia-uit/api/openapi`. It returns `{ operations: RestOperation[]; groups: { tag: string; operations: RestOperation[] }[] }`. Each `RestOperation` has `operationId`, `title`, `method`, `path`, and `href` strings; `href` is base-free `/api/rest/operations/<id>`. Export `REST_OPERATION_ROUTE = 'api/rest/operations/{operationId}'` for the Markdown generator.

- [ ] Add `groups one operation under each declared tag without duplicating routes`: assign `listPlans` to two tags in a fixture, assert one operation, two sidebar entries with the same href, and `GET` plus the spec path. Assert the real spec maps `createNote` to `/api/rest/operations/createNote` and the `Note` group. Add fixtures proving duplicate, missing and non-route-safe IDs are rejected before a route is emitted.
- [ ] Run `pnpm --dir docs exec vitest run` and confirm the new suite fails until the helper exists.
- [ ] Validate IDs by iterating raw HTTP method entries before using `vitepress-openapi` path metadata; its operation lookup can collapse duplicates. Report the method and path on a missing ID, and the offending ID on a duplicate or unsafe ID. Preserve spec tag order and provide a fallback group for untagged operations.
- [ ] Build the route loader from the canonical operation list. In `transformPageData`, apply dynamic titles without removing authorship validation. Render `OAOperation` using the joined spec and route ID and preserve custom server selection.
- [ ] Compose the sidebar **after** `withSidebar`; the installed plugin discards a sidebar passed into it. VitePress looks up the sidebar using `page.relativePath`, so scope the same tag groups to `/api/rest.md` for the overview and `/api/rest/` for nested endpoints. Preserve the generated general sidebar as the `/` fallback and exclude the raw operation templates from its scan. Test the resolved sidebar on `/api/rest`, `/api/rest/operations/createNote`, `/api/proto`, and one non-API page: only the REST pages get the tag groups, no link contains `[operationId]`, and endpoint links include the current site base when rendered. Use the final sidebar configuration as the general LLM plugin's navigation input.
- [ ] Run the docs Vitest suite and build. Inspect `/api/rest` for `Note` and `Plan` group links; inspect `createNote.html` and `listPlans.html` for endpoint paths and real schema/example content present in the spec. Serve the output and check Pagefind returns `createNote` for `/note/notes`; confirm `/api/proto` still works.
- [ ] Commit as `feat(docs): split REST reference into tagged endpoint pages`.

### Task 3: Add LLM Markdown without file collisions

**Files:** `pnpm-workspace.yaml`, `pnpm-lock.yaml`, `docs/package.json`, `docs/.vitepress/config.mts`, new `docs/.vitepress/rest-llms.spec.ts`.

**Interfaces:** Shared OpenAPI generator options use the joined spec, `REST_OPERATION_ROUTE`, `introductionRoute: 'api/rest'`, and `llmsTxt: false`; pass the same options to `generateLlmsFiles` and `llmsDevServer`.

- [ ] Add `writes OpenAPI Markdown without replacing general indexes`: seed a temporary output with sentinel `llms.txt` and `llms-full.txt`, call the generator with site pages for `api/rest.md` and `api/rest/operations/createNote.md` under `/notopia/pr/123/`, and assert both sentinels are unchanged, the overview links to `/notopia/pr/123/api/rest/operations/createNote.md`, and the operation includes `/note/notes` without `<OAOperation`.
- [ ] Run the docs Vitest suite and confirm the new test fails before wiring the shared options.
- [ ] Add `vitepress-plugin-llms` `1.15.0` to the pnpm catalog and to `docs` dev dependencies. Check the installed version’s hook order, ignore-pattern path format, and dev middleware behavior before wiring it up. Use `ignoreFiles` for the endpoint template and route subtree. Use `ignoreFilesPerOutput` to keep `api/rest.md` in `llms.txt` but prevent the general plugin from writing its per-page Markdown or copying its component source into `llms-full.txt`. Configure OpenAPI middleware before `llmstxt()` and invoke its generator from the existing `buildEnd`.
- [ ] Check the final Pagefind index after all hooks run. Pagefind wraps and awaits the configured `buildEnd`, so Scalar is already present when it scans. If Scalar's shell is indexed, exclude it using supported Pagefind settings while retaining REST operation results.
- [ ] Run the docs Vitest suite and build; assert root `llms.txt` links to the REST overview Markdown under the current base, that overview links to the operation Markdown, and `llms-full.txt` has neither raw operation markup nor internal working documents. In `nx dev docs` without Scalar, request `/notopia/api/rest/operations/createNote.md` as `text/markdown`, then request the same path with `?t=123` as Vite's module request; expect Markdown only on the first request and verify the endpoint page still loads client-side.
- [ ] Commit as `feat(docs): publish general and OpenAPI Markdown for LLMs`.

### Task 4: Move GitHub Pages publishing to `docs` and remove `page`

**Files:** `.github/workflows/ci.yaml`, `.github/workflows/docs-cleanup.yaml`, `docs/project.json`, `page/.gitignore`, `page/package.json`, `page/project.json`, `pnpm-workspace.yaml`, `pnpm-lock.yaml`. Do not rewrite historical specs or plans.

**Interfaces:** `docs:deployPage` publishes `docs/.vitepress/dist`; `docs:cleanup-previews` retains `pr`, `maxAgeDays: 7` and comment updates. Production still invokes `scripts/publish-docs.sh`.

- [ ] Capture the current CI build, deployment and cleanup rules as the comparison baseline. Query the docs-only, `apps/api-web`-only and OpenAPI-spec-only scenarios with `nx show projects --affected --with-target deployPage --files='<changed-file>' --json`; record the old target and use the same scenarios after the migration.
- [ ] Move `deployPage` and `cleanup-previews` to `docs/project.json` as non-cacheable targets; use `directory: "docs/.vitepress/dist"` for the preview deploy and keep `dependsOn: ["build"]`, `syncWithBaseBranch: true`, `https://notopia-uit.github.io/notopia`, `pr/{prNumber}`, and preview comments.
- [ ] Refactor the docs path through **`.github/workflows/ci.yaml`** using this mapping. Keep `DOCS_BASE`, Kroki and diagram caching, cache reset and bypass inputs, checkout settings, existing job and publish guards, token scopes, and `gh-pages` concurrency unchanged.

  | Workflow location                               | Change                                                                                                                                                                                                           |
  | ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
  | `jobs.main.outputs` and affected-detection step | Rename `page-affected` to `docs-affected` in both the output key and step ID. Keep `nx show projects --affected --with-target deployPage --json` and set `affected=true` only if its JSON array contains `docs`. |
  | `jobs.main` → Nx all                            | Change `--exclude=docs,page` to `--exclude=docs`; the dedicated docs job builds the publishable artifact using its own `DOCS_BASE`.                                                                              |
  | `jobs.docs.if`                                  | Read `needs.main.outputs.docs-affected`; keep the manual `run-many` override.                                                                                                                                    |
  | `jobs.docs` → Build Docs                        | Run `pnpm exec nx build docs` with the existing `nx-cache-action` wrapper, production configuration, stream output, and optional `--skipNxCache`.                                                                |
  | `jobs.docs` → Publish Preview                   | Run `pnpm exec nx run docs:deployPage --configuration production`; retain the trusted same-repo PR guard and `PR_NUMBER`.                                                                                        |
  | `jobs.docs` → Publish Production                | Pass `docs/.vitepress/dist` instead of `page/dist` to `scripts/publish-docs.sh`; retain the main-branch guard and existing authentication.                                                                       |

- [ ] Update `.github/workflows/docs-cleanup.yaml` to call `docs:cleanup-previews`. Remove the tracked `page` project files and workspace entry, refresh the lockfile with pnpm, and leave ignored local directories alone.
- [ ] Run `actionlint .github/workflows/ci.yaml .github/workflows/docs-cleanup.yaml` and rerun the affected queries for `docs/src/api/index.md`, `apps/api-web/src/app/app.tsx`, and `api/note/openapi.yaml`; each must return `docs`. Run the same query against an unrelated service implementation file (for example, `apps/ai/src/main.ts`); it must not return `docs`. Confirm `nx show projects --with-target deployPage --json` lists `docs`, not `page`. Search active workflows and workspace config for stale `page-affected`, `page:`, `page/dist`, and `nx build page` references. Do not invoke publishing or cleanup against the remote.
- [ ] Compare resolved `docs:deployPage` and `docs:cleanup-previews` in `nx show project docs --json` with the captured `page` targets: executor, build prerequisite, cache setting, preview and cleanup options must match; only the deployment directory and owning project change.
- [ ] In a disposable checkout, build with `DOCS_BASE=/notopia/`, then with `DOCS_BASE=/notopia/pr/123/` under the same cache. Preview and check `/api/`, `/api/openapi-scalar/` and one of its JS assets, `/api/rest`, an endpoint and its `.md`, and `/llms.txt` at each prefix. Also check `nx dev api-web` still serves the original dev base, while standalone preview serves the new built base and its assets.
- [ ] In that checkout, remove only `docs/.vitepress/dist` and rebuild with the original production base. Require a `docs:build` cache hit and confirm it restored VitePress HTML, Scalar HTML and JS, and endpoint and site-wide LLM files at production URLs. Change one Scalar source file and check `api-web:build` and `docs:build` invalidate; restore that test change, then change one OpenAPI source so it alters the joined spec and check `api:join`, `api-web:build`, and `docs:build` invalidate. Restore only the test edits.
- [ ] Commit as `ci(docs): publish VitePress artifact directly`.
