# Refactor: NestJS app building (document / search-worker) for OpenTelemetry + bundle size

> Handoff doc for a fresh session. Self-contained. Read top to bottom, then pick a proposal.

## 0. TL;DR — ✅ RESOLVED (Proposal B: curated regex externals)

- **Goal:** make the NestJS apps (`apps/document`, `apps/search-worker`) emit OTel **logs** (they already emit traces+metrics). Logs require `@opentelemetry/instrumentation-pino` to **patch** `pino`, which only works if `pino` is **not bundled** (require-in-the-middle needs a real `require('pino')`).
- **Resolution (committed):** stay on **CJS** rspack output, but replace Nx's weak `externalDependencies` *array* with a **custom regex `externals` function** set in a trailing `ExternalizePlugin` (Nx now gets `externalDependencies: 'none'` → empty externals, which the plugin overwrites). This externalizes whole heavy CJS clusters incl. subpath imports while keeping ESM/interop-fragile + singleton packages bundled. See `tasks/fix-nestjs-bundling-otel.md` and `story/nestjs-cjs-vs-esm-bundling.md` for rationale.
  - **document** externalizes: `pino`, `pg`, `kafkajs`, `@grpc/grpc-js`, `@aws-sdk/*`, `@smithy/*`. All patch at runtime (verified: `instrumentation-pino/pg/kafkajs/grpc/aws-sdk Applying`), no `@blocknote` crash, boots to DB-retry.
  - **search-worker** externalizes: `pino`, `kafkajs` (it uses no pg/grpc/aws). Both patch; boots to Kafka-connect retry.
- **The real size win was NOT aws-sdk — it was a duplicate source map.** `apps/document` had both `devtool: 'source-map'` *and* a redundant `new rspack.SourceMapDevToolPlugin({})`, which inlined a **64 MB base64 source-map data URI** into `main.cjs` on top of the external `.map`. Removing the redundant plugin (search-worker never had it) dropped `apps/document/dist/main.cjs` from **99.7 MB → 35.6 MB** (the external `main.cjs.map` is loaded only when debugging). Externalizing `@aws-sdk/*`/`@smithy/*` is a smaller additional win on top.
- **Historical failure modes (for context):**
  1. Bundle everything → huge bundle + inline-map → slow dev iteration.
  2. Externalize everything → tiny bundle (**888 KB**) but Node mis-loads ESM/interop-fragile packages (e.g. `@blocknote/core`) at runtime → crash. **That's why we keep `@blocknote/*`, `yjs`, `@hocuspocus/*`, `@bufbuild/protobuf` bundled.**
- **Optional follow-up (not done):** move bundled-only heavy packages (`@blocknote/*`, `yjs`, `@hocuspocus/*`, `@bufbuild/protobuf`, react/react-dom) to each app's `devDependencies` so `generatePackageJson` stops shipping them in the runtime image. Correctness-neutral (dead weight only); skipped to keep the externals↔dep-split coupling out of scope. **Externalized packages must stay in `dependencies`.**

---

## 1. Environment / constraints (all verified)

- Monorepo: Nx + pnpm (workspace catalog), rspack for the NestJS apps.
- `pnpm` uses the **isolated** node-linker (default; no `node-linker`/hoist in `.npmrc`). App deps live in `apps/<app>/node_modules` (symlinks into `node_modules/.pnpm`), **not** in the workspace-root `node_modules`.
- **Node 25.6.1** — natively executes `.ts` via type-stripping (relevant to the stack traces below).
- `tsconfig.base.json` has `"customConditions": ["@nx/source"]` (TS-solution setup; `tsconfig.json` has `references: true`).
- `@nx/rspack` version **22.7.5**.
- Build runs from the **workspace root** as cwd (Nx default). NOTE: `require()` resolution is based on the **file location** of the built `main.cjs` (`apps/<app>/dist/`), not cwd — externalized deps resolve via `apps/<app>/node_modules`. This was verified working (42/43 of the all-externalized deps resolved; the 1 miss was `@types/express`, types-only).
- Apps already generate a runtime manifest: `NxAppRspackPlugin({ generatePackageJson: true })` + `prune`/`copy-workspace-modules` targets. So shipping `node_modules` at runtime (Docker) is already handled.

---

## 2. How the build is wired (so you understand the levers)

- Each app has `apps/<app>/rspack.config.ts` exporting a config that includes `new NxAppRspackPlugin({ ... })` (imported from `@nx/rspack/app-plugin.js`).
- `NxAppRspackPlugin.apply()` → `applyBaseConfig()` (file: `node_modules/.pnpm/@nx+rspack@22.7.5_*/node_modules/@nx/rspack/src/plugins/utils/apply-base-config.js`).
- **Critical gotcha:** `applyBaseConfig` ends with `config.externals = externals;` — a **hard overwrite**. So any `externals:` you set in `rspack.config.ts` is **discarded**. (The repo previously had a `webpack-node-externals` call in the config; it was dead code for this reason. It has been removed.)
- The only supported lever is the plugin option **`externalDependencies`**:
  - default `'none'` → bundle everything (this was the original bug → 98 MB, pino inlined).
  - `'all'` → uses `webpack-node-externals` scanning `${workspaceRoot}/node_modules`. **Broken under pnpm isolated** because the app's deps aren't in the workspace-root `node_modules`.
  - **array** → Nx pushes a function: `if (externalDependencies.includes(ctx.request)) callback(null, 'commonjs ' + request)`. **Exact `ctx.request` match** (subpath imports like `rxjs/operators` would NOT match). Workspace libs not in the array stay bundled.
- Nx auto-bundles workspace libs (`@notopia-uit/*`) via the project graph; with the **array** branch they're simply not listed, so they stay bundled. Good (we want them bundled).

---

## 3. The original problem (logs not emitted) — root cause

NestJS apps exported traces + metrics but **never logs**.

- Logs flow: `nestjs-pino` → `pino` → `@opentelemetry/instrumentation-pino` adds a `pino.multistream` destination that forwards records to the OTel Logs SDK → OTLP exporter.
- `instrumentation-pino` (v0.60.0) attaches that destination by **patching the `pino` module via require-in-the-middle** at `require('pino')` time.
- Because rspack **bundled** `pino` inline, there was **no runtime `require('pino')`** to intercept → instrumentation never patched → no log bridge → no logs.
- Verified: in the 98 MB all-bundled build, only **core-module** instrumentations patched (`http`, `dns`, `net`); none of `pino`/`pg`/`kafkajs`/`grpc`/`nestjs-core` did.

### Ruled out (do NOT chase these again)
- **Duplicate `@opentelemetry/api-logs` (0.214 vs 0.215):** harmless. Both register under the same global symbol `Symbol.for('io.opentelemetry.js.api.logs')` with the same `API_BACKWARDS_COMPATIBILITY_VERSION = 1`, so the global LoggerProvider IS shared. (`sdk-node@0.215` pulls api-logs 0.215; `instrumentation-pino@0.60.0` pins `^0.214.0` which for 0.x means `<0.215`.)
- **`pino-pretty` custom `stream:`** does NOT block log-sending. `instrumentation-pino` wraps the existing destination via `multistream([{stream: yourPretty}, {stream: otelStream}])` — both run. (Source: `instrumentation-pino/build/src/instrumentation.js` ~lines 60-86.)
- **`@opentelemetry/sdk-node@0.215`** auto-creates the LoggerProvider from `OTEL_LOGS_EXPORTER` (default otlp) when no `logRecordProcessors` are passed (`sdk.js` ~lines 145-157, 265). So the pipeline exists; the only missing piece was the pino patch.

---

## 4. The fix direction and the new problem

To let instrumentation patch, we externalize packages. Set `externalDependencies` (array) on `NxAppRspackPlugin`.

### Attempt 1 — externalize ALL package.json deps (minus `@notopia-uit/*`)
- Result: bundle shrank to **888 KB** (document) / **72 KB** (search-worker). `pino`, `grpc`, `nestjs-core` patched.
- BUT at runtime (`nx run document`) it **crashes**:

```
ESM loader error: TypeError: h.default.extend is not a function
    at Object.<anonymous> (.../@blocknote/core/src/blocks/defaultBlocks.ts:141:10)   <-- Code.extend({...})
    at Module._compile (node:internal/modules/cjs/loader:1811:14)
    ...
    at Module.patchedRequire (.../require-in-the-middle/index.js:209:27)
```

`nx run search-worker` crashes the same way (it pulls `@blocknote` via `@notopia-uit/lib-server`).

### Exact problem (ROOT-CAUSED via isolated repro — supersedes earlier guesses)

The crash is **purely blocknote's broken published CJS build under Node 25's `require(ESM)` interop.** It is
**NOT** caused by OTel / require-in-the-middle, **NOT** by rspack, and **NOT** a `.ts`-resolution issue. Proven
with bare `node` (no bundler, no instrumentation) — see `tmp/document/bn-test.cjs` / `bn-test-esm.mjs`:

| Repro | What Node loads | Result |
|---|---|---|
| `require('@blocknote/core')` (plain node, **no RITM/rspack/OTel**) | `dist/*.cjs` (**require** condition) | 💥 `h.default.extend is not a function` |
| `import` the ESM entry (incl. `@blocknote/server-util` + `ServerBlockNoteEditor.create()`) | `dist/blocknote.js` (**import** condition) | ✅ works |

- `@blocknote/core@0.50.0` `exports["."]` = `{ import: ./dist/blocknote.js, require: ./dist/blocknote.cjs }`.
  The **`.cjs` build is broken**; the **`.js` (ESM) build is fine**. The `h.default.extend` is blocknote's own
  CJS-interop artifact (a `.default` access on a dep that Node 25's `require(ESM)` shapes differently). The
  `require-in-the-middle` frame in the *original* stack was **incidental** — bare `node` crashes identically.
- Therefore the lever is **which build loads (ESM good vs CJS broken)**, governed by the **import vs require
  condition**:
  - **Externalize + CJS bundle** → emits `require('@blocknote/core')` → Node picks the **require** condition →
    broken `.cjs` → crash. ← *this is why "externalize + cjs failed".*
  - **Bundle (rspack)** → resolves via `mainFields: ['module','main']` → inlines the **ESM** build → works. It
    works because it pulls the ESM build, not because rspack "fixes interop".
  - **Next.js `serverExternalPackages`** (BlockNote's official server guidance) externalizes too, but loads via
    the **import** condition → ESM build → works. We can't copy that under CJS output (require → `.cjs`).
- **jsdom landmine (we use `@blocknote/server-util` → jsdom):** BlockNote issue #1939 — **bundling** server-util
  bundles jsdom, whose dynamically-loaded `lib/jsdom/living/xhr/xhr-sync-worker.js` worker file isn't emitted →
  runtime `Cannot find module './xhr-sync-worker.js'`. So even though bundling blocknote works, **`jsdom` must
  be externalized** (plain CJS, safe) so Node loads it from `node_modules` and finds the worker file.
- Upgrade note: a newer `@blocknote/*` may ship a fixed `.cjs` build, which would make plain externalization
  viable — re-test `tmp/document/bn-test.cjs` after any bump before relying on it.

### Attempt 2 (CURRENT, uncommitted working tree) — curated allowlist of instrumented CJS packages only
- `apps/document/rspack.config.ts`: `const externalDependencies = ['pino', 'pg', 'kafkajs', '@grpc/grpc-js'];`
- `apps/search-worker/rspack.config.ts`: `const externalDependencies = ['pino', 'kafkajs'];`
- Result: **works** — `nx run document` boots through to DB-retry (no blocknote crash), and instrumentation patches `pino`, `pg`, `kafkajs`, `grpc` (+ core `http/dns/net`). Logs bridge is attached.
- Downside: everything else (incl. the giant `@aws-sdk/*`) is bundled → `main.cjs` ≈ **98 MB** dev → slow rebuilds. This is what the user wants to avoid.

> Note: full end-to-end log *record* emission couldn't be observed locally because there's no DB running — the app crashes in the DB-retry loop before NestJS finishes bootstrap and flushes pino (Nest uses `bufferLogs: true`; logs flush only after `NestFactory.create()` resolves, i.e. `app.useLogger(logger)` in `main.ts`). The instrumentation **patch firing** is the deterministic proof the bridge is in place; record emission follows once it runs against a real DB + collector.

---

## 5. How to reproduce / verify (no DB needed — we only check patching)

```bash
# From repo root. Routes logs to console, disables trace/metric exporters to cut noise.
OTEL_LOG_LEVEL=debug OTEL_LOGS_EXPORTER=console OTEL_TRACES_EXPORTER=none OTEL_METRICS_EXPORTER=none \
  pnpm exec nx run document:run --no-tui > /tmp/doc.log 2>&1; echo "exit=$?"

du -h apps/document/dist/main.cjs                       # bundle size
grep -aoE "instrumentation-[a-z-]+ Applying" /tmp/doc.log | sort -u   # which patched
grep -aiE "Code.extend|is not a function" /tmp/doc.log  # blocknote crash? (empty = good)
grep -aiE "Starting Nest application|Unable to connect to the database" /tmp/doc.log  # booted?
```

- **Good run** = no `Code.extend` error, reaches "Starting Nest application…" + DB retry, and `instrumentation-pino Applying` appears.
- The process exits non-zero on DB connection failure — that's expected (no DB). It is NOT a build/instrumentation failure.

---

## 6. Proposals

### Proposal A — STOP bundling these NestJS services (RECOMMENDED)

NestJS backends are normally not bundled. Build with **swc transpile-only** (per-file), keep **all** `node_modules` external.

**Why it's the best fix:**
- Fast builds; output is only your compiled `src/` (a few hundred KB).
- All instrumentation patches natively (every package is a real `require`), zero allowlist to maintain.
- No ESM/CJS interop landmines — Node loads each package as authored (blocknote, yjs, etc. just work).
- Lets you DELETE the bundling-only workarounds in `rspack.config.ts`: the `IgnorePlugin`/`lazyImports` list, the `file-type` stub (`stub.js`), and the `@protobufjs/inquire` shim (`inquire-shim.js`) — these exist only because of bundling.
- Runtime `node_modules` already handled (`generatePackageJson` + `prune`/`copy-workspace-modules`).

**Implementation sketch (verify details against installed Nx 22.7.5):**
- Replace the rspack build target with `@nx/js:swc` (the repo already has `@swc/core`, `@swc-node/register`, `unplugin-swc`). Output CommonJS to `apps/<app>/dist`, preserve module structure (no bundling).
- Keep the run via `@nx/js:node` (executor unchanged) pointing at the transpiled entry.
- Ensure `reflect-metadata` is imported first (already is, in `main.ts`), and decorators/`emitDecoratorMetadata` are enabled in the swc config (NestJS needs `legacyDecorator` + `decoratorMetadata` — same as the existing `builtin:swc-loader` options in `rspack.config.ts`).
- Remove `rspack.config.ts`, `stub.js`, `inquire-shim.js`, and the `webpack-node-externals` devDep once migrated.
- Watch out for: TypeORM CLI / migrations build path (`tsconfig.database.json`, `database/`), the `mise.toml`, and any `out-tsc`/`dist` path assumptions in `project.json` and `Dockerfile`.
- **Verify** with the Section 5 probe: instrumentation patches, no crash, then a real run (DB+collector) shows logs landing.

**Cost:** a real one-time build migration; touch `project.json` build target + Dockerfile. Do it on a branch.

### Proposal B — Stay on rspack, expand the curated allowlist

Keep bundling but also externalize the **safe heavyweight CJS clusters** to shrink the bundle, while bundling the ESM-fragile ones.

- Externalize (CJS, ship dist, safe): the instrumented set + `@aws-sdk/client-s3`, `@aws-sdk/s3-request-presigner`, `typeorm`, `express`, `ws`, `@grpc/proto-loader`, `protobufjs`, and the **NestJS cluster together** (`@nestjs/common`, `@nestjs/core`, `@nestjs/microservices`, `@nestjs/platform-express`, `@nestjs/config`, `@nestjs/typeorm`) **plus `rxjs`** (must be one instance with nestjs — `instanceof Observable`). `reflect-metadata` is global-side-effect so safe either way.
- Keep bundled (ESM/interop-fragile/source): `@blocknote/*`, `@hocuspocus/*`, `yjs`, `@bufbuild/protobuf`, `nestjs-pino`, `pino-http`, `pino-pretty`, `@opentelemetry/*`, `jose`, `marked`, `zod`, `react`, `react-dom`.
- **Risk:** it's a hand-maintained list and you must test that each externalized cluster doesn't interop-break (same class of bug as blocknote). Bundle stays larger than Proposal A.
- `@aws-sdk/*` is the single biggest contributor to the 98 MB — externalizing just that + the instrumented set may already be a big, low-risk win if Proposal A is too much for now.

### Proposal C — Leave as-is (current working tree)

- 4-package allowlist (Section 4, Attempt 2). Works; logs+traces patch; bundle ~98 MB dev (slow). Lowest effort. Just commit it.

---

## 7. Immediate next step for the fresh session

1. `main` is broken (`externalDependencies` = all deps → blocknote crash). Either:
   - commit the current working-tree 4-pkg allowlist (Proposal C) to unbreak `main` immediately, OR
   - implement Proposal A/B on a branch and merge that.
2. Recommended: do **Proposal A** on a branch, verify with Section 5, fall back to **Proposal C** if blocked.
3. Always re-run the Section 5 probe after changes: confirm `instrumentation-pino Applying` patches AND no `Code.extend` crash, for BOTH `document` and `search-worker`.

---

## 8. Files / facts reference

- `apps/document/rspack.config.ts`, `apps/search-worker/rspack.config.ts` — build configs; `externalDependencies` passed to `NxAppRspackPlugin`. Currently the curated 4-pkg / 2-pkg allowlist (uncommitted).
- `apps/document/src/otel.ts`, `apps/search-worker/src/otel.ts` — NodeSDK + `getNodeAutoInstrumentations()`; sets `OTEL_NODE_ENABLED_INSTRUMENTATIONS` default. (`pino` is in the enabled list.)
- `apps/document/src/main.ts` — imports `./otel` first, then app; `bufferLogs: true`, `app.useLogger(Logger)` after create.
- `apps/*/src/app.module.ts` — `LoggerModule` (nestjs-pino) with `pino-pretty` stream; `OpenTelemetryModule` (nestjs-otel) for metrics.
- Nx externals logic: `@nx/rspack/src/plugins/utils/apply-base-config.js` (`config.externals = externals` overwrite; `externalDependencies` branches).
- Already-merged related OTel work on `main` (do not redo):
  - `fix(otel): connect Go traces, emit gin/grpc metrics, and enable web OTLP logs` (Go propagator via `autoprop` + explicit providers to otelgin/otelgrpc; web `instrumentation.ts` gets `logRecordProcessors`).
  - `fix(document,search-worker): externalize npm deps so OTel can instrument them` (introduced `externalDependencies`; this is the commit that currently over-externalizes → blocknote crash).
  - `chore: move type express to dev deps in nestjs`.
  - `chore(lib,ui,pb): use peerDependencies for host-provided singletons` (ui/pb peer-dep hygiene; may be on a branch `chore/workspace-lib-peer-deps`).

## 9. Key commands

```bash
# build only (size check)
pnpm exec nx run document:build --no-tui && du -h apps/document/dist/main.cjs

# the patch/boot probe (Section 5)
OTEL_LOG_LEVEL=debug OTEL_LOGS_EXPORTER=console OTEL_TRACES_EXPORTER=none OTEL_METRICS_EXPORTER=none \
  pnpm exec nx run document:run --no-tui 2>&1 | grep -aoE "instrumentation-[a-z-]+ Applying" | sort -u

# lint after changes
pnpm exec nx lint document && pnpm exec nx lint search-worker
```
