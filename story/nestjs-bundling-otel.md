# Bundling the NestJS apps (document / search-worker): OpenTelemetry logs, CJS, and the externals split

> **Single consolidated story.** Supersedes and merges the two former docs
> (`js-bundling-rspack.md` — the OTel-logs / bundle-size investigation, and
> `nestjs-cjs-vs-esm-bundling.md` — why ESM output is a dead end). Read top to
> bottom. It is self-contained: a fresh session should be able to understand both
> *what* the build does and *why every alternative was rejected* without re-deriving
> any of it. Long on purpose — the reasoning is the point.

---

## 0. TL;DR — the final design (✅ shipped)

The two NestJS apps (`apps/document`, `apps/search-worker`) are bundled with **rspack to CommonJS**
(`main.cjs`, `isEsm = false`). The externals rule is:

> **Externalize everything by default; bundle ONLY the blocknote editor cluster** (the packages that
> cannot survive a native `require()`), plus the workspace libs and the tiny transpile runtimes.

Concretely, `keepBundled` (bundled) is:

```
@blocknote/*  prosemirror-*  @handlewithcare/*  @tiptap/*
yjs  y-protocols  y-prosemirror  lib0
react  react-dom  scheduler  @floating-ui/*
@bufbuild/protobuf
@notopia-uit/*           (workspace libs — they import @blocknote)
tslib  @swc/helpers      (transpile runtimes)
```

**Everything else is external** — all backend infra (`@nestjs/*`, `rxjs`, `typeorm`, `express`, `pg`,
`kafkajs`, `@grpc/*`, `@aws-sdk/*`, `@smithy/*`, `pino`, `@opentelemetry/*`, `jose`, `marked`, `zod`, …)
**and** blocknote's well-behaved dual-package leaf deps (the `unified`/`remark`/`mdast`/`hast` markdown
stack). They all load fine via native `require()`.

Results (verified):

| | before | after |
|---|---|---|
| `apps/document/dist/main.cjs` | **98 MB** dev | **~4.6 MB** |
| `apps/search-worker/dist/main.cjs` | (similar) | **~4.2 MB** |
| OTel instrumentations that patch (document) | `http`/`dns`/`net` only | `pino`, `pg`, `kafkajs`, `grpc`, `nestjs-core`, `express`, `router`, `aws-sdk` (+ `http`/`dns`/`net`) |

Why this shape and not any of the half-dozen alternatives we tried is the rest of this document.

---

## 1. The goal: emit OpenTelemetry **logs**

The apps already emitted traces and metrics, but **never logs**. The log pipeline is:

```
nestjs-pino → pino → @opentelemetry/instrumentation-pino → OTel Logs SDK → OTLP exporter
```

`@opentelemetry/instrumentation-pino` attaches a `pino.multistream` destination that forwards every log
record to the OTel Logs SDK. **It attaches that destination by patching the `pino` module via
require-in-the-middle (RITM) at `require('pino')` time.** No patch → no bridge → no logs.

### Things we ruled out (do NOT chase again)

These were investigated and proven harmless; the missing logs were *only* the pino patch:

- **Duplicate `@opentelemetry/api-logs` (0.214 vs 0.215).** Harmless. Both register under the same global
  symbol `Symbol.for('io.opentelemetry.js.api.logs')` with the same `API_BACKWARDS_COMPATIBILITY_VERSION = 1`,
  so the global LoggerProvider is shared. (`sdk-node@0.215` pulls api-logs 0.215; `instrumentation-pino@0.60.0`
  pins `^0.214.0`, which for 0.x means `<0.215`.)
- **`pino-pretty` custom `stream:`.** Does not block log-sending. `instrumentation-pino` wraps the existing
  destination via `multistream([{stream: yourPretty}, {stream: otelStream}])` — both run.
  (Source: `instrumentation-pino/build/src/instrumentation.js` ~lines 60–86.)
- **`@opentelemetry/sdk-node@0.215`.** Auto-creates the LoggerProvider from `OTEL_LOGS_EXPORTER` (default
  otlp) when no `logRecordProcessors` are passed (`sdk.js` ~lines 145–157, 265). The pipeline already exists;
  the only missing piece was the pino patch.

### The root cause in one line

rspack **bundled `pino` inline**, so there was **no runtime `require('pino')`** for RITM to intercept →
instrumentation never patched → no logs. In the original 98 MB all-bundled build, only the *core-module*
instrumentations patched (`http`, `dns`, `net`); none of `pino`/`pg`/`kafkajs`/`grpc`/`nestjs-core` did.

**The fix direction is therefore forced: the OTel-instrumented packages must be EXTERNAL** (real
`require()`s), not inlined. Everything below is the consequence of pursuing that without breaking the apps.

---

## 2. Why CJS output, not ESM (a settled dead end)

Before discussing *which* packages to externalize, settle the module format, because "just switch the bundle
to ESM" comes up and is wrong. The apps emit **`main.cjs` (`isEsm = false`) deliberately.**

### 2.1 RITM only intercepts CommonJS `require()`

The whole logs fix relies on RITM hooking `Module._load` and wrapping each subsequent `require('pino')`.
In an **ESM** bundle, external deps are pulled with `import`, which RITM cannot intercept. OTel's ESM path
needs **import-in-the-middle (IITM)**, a different mechanism hooking Node's ESM loader via
`module.register()` / `--import`.

### 2.2 A loader hook registered *from inside the graph* is always too late

ESM links **all** static imports before any module body executes (depth-first, post-order). The bootstrap is
load-bearing:

```ts
// apps/*/src/main.ts
import './otel';                 // otelSdk.start() → installs the RITM hook
import '@notopia-uit/lib/yjs';
import 'reflect-metadata';
// …then NestFactory.create(...)
```

Under **CJS** this works: `import './otel'` runs `otelSdk.start()` **synchronously first**, and the external
deps are `require()`d **later** (during Nest bootstrap) → intercepted. Under **ESM** in a single bundled
file, the external `import pino` is a top-level import **hoisted above** the inlined `import './otel'` side
effect, so `pino` is linked *before* `otelSdk.start()` runs → nothing left to patch. This is structural, not
incidental.

The only ESM-correct fix is to register the hook at the Node CLI level *before* the graph loads
(`node --import @opentelemetry/instrumentation/hook.mjs`), which means abandoning `import './otel'`-first,
rewiring the `@nx/js:node` run/serve invocation and the Dockerfile entrypoint, **and** trusting
`instrumentation-pino` to patch a CJS `pino` through the CJS→ESM interop under IITM — unproven for this stack.

### 2.3 Concrete ESM breakages already in the code

- `apps/document/src/main.ts` uses `module.hot` / `module.hot.dispose(...)`. `module` is a **CJS-only**
  global; it does not exist in ESM.
- The app `package.json` is `"type": "module"`, but the build emits `.cjs` on purpose (the `.cjs` extension
  forces CJS regardless of `type`). Flipping to `.mjs` re-exposes every issue here.
- The suppressed `Critical dependency` rspack warnings (`typeorm`, `nestjs`, `express`, `app-root-path`,
  `load-esm`) are **dynamic `require(expr)`** inside bundled CJS deps. ESM has no `require`; rspack injects
  `createRequire` shims, but dynamic-expression requires there are fragile.
- `reflect-metadata` ordering and NestJS decorator metadata (`emitDecoratorMetadata`) have extra sharp edges
  under ESM hoisting.
- rspack `experiments.outputModule` + `target: node` + node externals is far less battle-tested than CJS.

### 2.4 ESM doesn't even solve the blocknote crash

ESM's only marginal benefit would be that externalizing blocknote resolves the **import** condition (the good
ESM build) instead of the broken **require** `.cjs`. But we don't need that — we just **keep blocknote
bundled** (§4), which already fixes the crash under CJS. ESM solves nothing CJS doesn't, while costing
§2.1–§2.3.

> **Verdict:** Stay on CJS. Do not re-attempt ESM unless the OTel strategy is first redesigned around a
> `node --import` ESM loader hook and validated end-to-end for pino logs.

---

## 3. How the build is wired (the levers, and the Nx gotcha)

- Each app has `apps/<app>/rspack.config.ts` exporting a config that includes
  `new NxAppRspackPlugin({ ... })` (from `@nx/rspack/app-plugin.js`). `@nx/rspack` is **22.7.5**.
- `NxAppRspackPlugin.apply()` → `applyBaseConfig()`
  (`node_modules/.pnpm/@nx+rspack@22.7.5_*/node_modules/@nx/rspack/src/plugins/utils/apply-base-config.js`).
- **Critical gotcha #1:** `applyBaseConfig` ends with `config.externals = externals;` — a **hard
  overwrite**. Any `externals:` you set directly in the rspack config object is **discarded**. (The repo once
  had a `webpack-node-externals` call in the config; it was dead code for exactly this reason.)
- **Critical gotcha #2 — the `externalDependencies` option is too weak:**
  - default `'none'` → bundle everything (the original bug; pino inlined; 98 MB).
  - `'all'` → uses `webpack-node-externals` scanning `${workspaceRoot}/node_modules`. **Broken under pnpm
    isolated** (app deps aren't in the workspace-root `node_modules`).
  - **array** → Nx pushes a function doing an **exact `ctx.request` match** only
    (`if (externalDependencies.includes(ctx.request)) cb(null, 'commonjs '+request)`). Subpath imports like
    `@aws-sdk/client-s3` or `rxjs/operators` **won't match**, and you can't express "externalize a whole
    scope."

**How we beat both gotchas:** pass `externalDependencies: 'none'` to Nx (so it sets `config.externals = []`),
then install our **own externals function** in a trailing plugin (`ExternalizePlugin`) whose `apply()`
overwrites `compiler.options.externals` **after** Nx ran (it sits later in the `plugins` array). That gives
us a real regex/name-based externals function with full control.

`externalsPresets: { node: true }` stays set, so node builtins are externalized by the preset independently
of our function (our function also lets builtins fall through to be safe).

---

## 4. The blocknote crash — root cause (the pivot of the whole design)

The first instinct ("externalize everything so OTel can patch it") produces a tiny bundle (888 KB) but
**crashes at runtime**:

```
TypeError: h.default.extend is not a function
    at .../@blocknote/core/.../defaultBlocks (Code.extend({...}))
```

`search-worker` crashes identically (it pulls `@blocknote` via `@notopia-uit/lib-server`).

### 4.1 Root-caused with a bare-node repro (supersedes earlier guesses)

The crash is **purely blocknote's broken published CJS build under Node's `require(ESM)` interop.** It is
**NOT** caused by OTel/RITM, **NOT** by rspack, and **NOT** a `.ts`-resolution issue. Proven with plain
`node` (no bundler, no instrumentation):

| Repro | What Node loads | Result |
|---|---|---|
| `require('@blocknote/core')` (plain node) | `dist/*.cjs` (**require** condition) | 💥 `h.default.extend is not a function` |
| `import` the ESM entry (incl. `ServerBlockNoteEditor.create()`) | `dist/blocknote.js` (**import** condition) | ✅ works |

- `@blocknote/core@0.50.0` `exports["."]` = `{ import: ./dist/blocknote.js, require: ./dist/blocknote.cjs }`.
  The **`.cjs` build is broken**; the **`.js` (ESM) build is fine**. `h.default.extend` is blocknote's own
  CJS-interop artifact — a `.default` access on a dep that Node's `require(ESM)` shapes differently. The RITM
  frame in the *original* stack trace was incidental; bare `node` crashes the same way.
- So the lever is **which build loads (ESM good vs CJS broken)**, governed by the **import vs require
  condition**:
  - **Externalize + CJS bundle** → emits `require('@blocknote/core')` → Node picks the **require** condition →
    broken `.cjs` → crash. *(This is why "externalize blocknote" fails.)*
  - **Bundle (rspack)** → resolves via `mainFields: ['module','main']` → inlines the **ESM** build → works.
    It works because it pulls the ESM build, not because rspack "fixes" interop.
- **Upgrade note:** a newer `@blocknote/*` may ship a fixed `.cjs`, which would make plain externalization
  viable. Re-test the bare-node repro after any bump before relying on it.

**Conclusion: `@blocknote/*` MUST be bundled.** That single fact is the seed the whole `keepBundled` list
grows from.

### 4.2 The jsdom landmine (the opposite direction)

We use `@blocknote/server-util` → `jsdom`. **BlockNote issue #1939:** *bundling* jsdom bundles its
dynamically-loaded worker file `lib/jsdom/living/xhr/xhr-sync-worker.js`, which rspack doesn't emit →
runtime `Cannot find module './xhr-sync-worker.js'`. So **jsdom must be EXTERNAL** (it's plain CJS, safe) so
Node loads it from `node_modules` and finds the worker. Under our default-external rule this is automatic —
but it's the canonical example that the bundle/externalize decision has failures in *both* directions:

- `@blocknote/core` → **must bundle** (externalizing loads the broken `.cjs`).
- `jsdom` → **must externalize** (bundling drops a worker file).

There is no universally safe default; you maintain an exception list whichever way you point.

---

## 5. The journey: every approach we tried and why it was rejected

This is the heart of the doc. Each attempt taught one fact that constrains the final design.

### Attempt 0 — Bundle everything (the starting state)
- `externalDependencies: 'none'`. `main.cjs` ≈ **98 MB** dev.
- pino inlined → **no logs** (§1). Slow rebuilds.
- ❌ Fails the goal.

### Attempt 1 — Externalize everything (minus workspace libs)
- Bundle shrank to **888 KB** (document) / **72 KB** (search-worker); `pino`/`grpc`/`nestjs-core` patched. 🎉
- But **crashes** on the blocknote broken `.cjs` (§4.1).
- ❌ Proves: you cannot blindly externalize; blocknote (and its ilk) must bundle.
- ✅ Also proved something crucial used later: in that 888 KB build, the app ran **all the way to the
  blocknote crash with no `MODULE_NOT_FOUND`** — i.e. *externalized transitive deps resolve fine at runtime
  in this repo's run setup.* (The nx node executor / pnpm layout makes them resolvable.) Keep this in mind.

### Attempt 2 — Curated externalize **allowlist** (the first thing that worked)
- Externalize only the OTel-instrumented CJS packages:
  document `['pino','pg','kafkajs','@grpc/grpc-js']`, search-worker `['pino','kafkajs']`.
- ✅ Works: boots to DB-retry, no blocknote crash, `pino`/`pg`/`kafkajs`/`grpc` patch. Logs bridge attached.
- ⚠️ Bundle still ≈ 98 MB → slow. **But here we found the real size culprit (below).**

#### 5.a The 98 MB was mostly a **duplicate source map**, not packages
`apps/document` had BOTH `devtool: 'source-map'` **and** a redundant `new rspack.SourceMapDevToolPlugin({})`.
The plugin inlined a **64 MB base64 source-map data URI** into `main.cjs` *on top of* the external
`main.cjs.map` (35.6 MB of actual code + 64 MB inline map ≈ 99.7 MB). `search-worker` never had that plugin.
**Removing it dropped `main.cjs` from 99.7 MB → 35.6 MB** with zero behavior change. *Lesson: measure before
blaming `@aws-sdk`.* (Externalizing `@aws-sdk/*`/`@smithy/*` is a real but secondary win.)

### Attempt 3 — Invert to **default-external + `keepBundled` editor list + issuer fallback**
- Rule: externalize everything except a `keepBundled` regex list (the editor/CRDT stack), plus "bundle
  anything imported *from within* a bundled package" (an issuer-path regex) to drag transitive editor deps in.
- Bundle ≈ **18 MB**. Builds.
- ❌ A static scan of the emitted `require()`s found **41 externalized packages that don't resolve at
  runtime** (`@tiptap/*`, `mdast-util-*`, `micromark-*`, `hast-util-*`, jsdom internals, …). The one-level
  issuer regex only catches direct deps of named packages; deep transitives leak out as broken externals.
  They'd `MODULE_NOT_FOUND` the moment that editor code path runs — *invisible at boot.*

### Attempt 4 — Default-external + **`require.resolve`-based** decision (clever, still wrong)
- Idea: instead of guessing, externalize a package **only if it actually resolves** from the app
  (`createRequire(app).resolve(name)`), else bundle. "Self-validating; MODULE_NOT_FOUND impossible."
- Bundle ≈ **4 MB**. Booted.
- ❌ Two fatal flaws:
  1. **`package.json` fallback false-positives.** `resolve('@tiptap/core')` fails but
     `resolve('@tiptap/core/package.json')` succeeds → marked "resolvable" → externalized → but the real
     `require('@tiptap/core')` at runtime fails (no/blocked main export). Static check still showed **33
     unresolvable**.
  2. **Build-time vs runtime resolution don't match** under pnpm/Nx — the anchor during the build resolved a
     different set than `apps/<app>/dist/main.cjs` sees at runtime. A resolver-based heuristic is unreliable
     here. *Lesson: don't decide externalization by filesystem resolution under pnpm isolated.*

### Attempt 5 — Bundle **only `@blocknote/*`** (+ workspace libs + helpers)
- The minimalist reading of "just bundle the broken one." Bundle ≈ **978 KB** / **604 KB**. 🎉 tiny.
- ❌ Crashes with a *new* error:
  ```
  No "exports" main defined in .../@handlewithcare/prosemirror-...
  ```
  blocknote → `prosemirror-*` → `@handlewithcare/prosemirror-*`, which ships an **ESM-only `exports` map with
  no CJS main**. Externalized, Node's native `require()` of it throws. *Lesson: "bundle only blocknote" is
  impossible — its prosemirror editor cluster contains packages that can't be `require()`d natively, so they
  must bundle with it.*

### Attempt 6 — Bundle the **blocknote editor cluster**, externalize everything else ✅ (FINAL)
- `keepBundled` = blocknote + prosemirror + @handlewithcare + @tiptap + yjs/y-*/lib0 + react/react-dom/
  scheduler + @floating-ui + @bufbuild/protobuf + @notopia-uit/* + tslib/@swc/helpers (§0).
- Externals function is **simple** (no resolver, no issuer): bundle app code / aliases / builtins / the
  `keepBundled` cluster; externalize all else.
- ✅ Bundle ≈ **4.6 MB** / **4.2 MB**. No `Code.extend`, no `No "exports" main`, no `MODULE_NOT_FOUND`, both
  boot. **More** instrumentation patches than ever (because nestjs/express/aws/pg/kafka are now external):
  `pino`, `pg`, `kafkajs`, `grpc`, `nestjs-core`, `express`, `router`, `aws-sdk` + core `http`/`dns`/`net`.

Why bundling fixes both crash classes: rspack reads the package's **`exports`/`mainFields` at build time**
(it understands ESM `exports`), compiles the ESM build in, and there is **no runtime `require()` to fail** —
neither blocknote's broken `.cjs` nor `@handlewithcare`'s missing CJS main is ever hit.

---

## 6. The rule of thumb (how to classify a package)

> **Bundle** a package only if it **cannot survive a native `require()`** under CJS. Two sub-cases:
> 1. **Broken require-condition interop** — e.g. `@blocknote/core` (its `.cjs` throws `Code.extend …`).
> 2. **ESM-only / no CJS `exports` main** — e.g. `@handlewithcare/prosemirror-*`, prosemirror view layer;
>    `require()` throws `No "exports" main defined`.
>
> Plus, by extension, anything that *imports* a must-bundle package and would otherwise run that import
> natively (`@notopia-uit/*` workspace libs import `@blocknote`), and singletons that must stay one instance
> while a bundled consumer holds them (`yjs` CRDT identity; `@bufbuild/protobuf` type registry).
>
> **Externalize** everything else — it loads fine natively, keeps the bundle small, and (for the
> OTel-instrumented packages) is the *only* way RITM can patch it. This includes blocknote's own dual-package
> leaf deps (`unified`/`remark`/`mdast`/`hast`), which are well-behaved CJS.

A new editor dep that's ESM-only may occasionally need adding to `keepBundled`. The failure is **loud**
(`No "exports" main` / `Code.extend` at boot) and the fix is **one regex line**. Re-run the §8 probe after any
`@blocknote`/prosemirror bump.

---

## 7. Why a custom externals function, not `webpack-node-externals`

This came up repeatedly; the answer is decisive.

1. **It scans a directory; we must decide by name.** `webpack-node-externals` externalizes "whatever package
   folders exist in a scanned `node_modules`." Under pnpm's **isolated** linker that gives the **wrong answer
   for the single most important package**: `pino` is a *transitive* dep (we depend on `nestjs-pino`/
   `pino-http`/`pino-pretty`, not `pino`), so it isn't a flat entry in `apps/<app>/node_modules` → a dir-scan
   **bundles** it → `instrumentation-pino` can't patch → **no logs**, the exact bug we set out to fix.
   Pointing it at `.pnpm` doesn't help (the folder is `pino@9.x`, which won't match `require('pino')`).
2. **The decision axis is "editor cluster vs backend," a name-based regex set** — independent of pnpm's
   on-disk layout. `webpack-node-externals`' `allowlist` could express the keep-bundled side, but you'd write
   the *same* regex list **and** still be fighting #1. It buys nothing and costs correctness.
3. **Smaller and clearer.** The whole thing is one `keepBundled` array + a ~10-line function. No
   `modulesDir`/`additionalModuleDirs`/`importType` tuning, no dependency on store layout.

Con (honest): a custom default-external means a new ESM-only editor dep could need a `keepBundled` line later
— but `webpack-node-externals` wouldn't save you there either (you'd touch its `allowlist` the same way).

---

## 8. How to verify (the probe — no DB/Kafka needed; we check patching + no crash)

```bash
# Build + size
pnpm exec nx run document:build --no-tui && du -h apps/document/dist/main.cjs

# Patch + boot probe (routes logs to console, silences trace/metric exporters)
OTEL_LOG_LEVEL=debug OTEL_LOGS_EXPORTER=console OTEL_TRACES_EXPORTER=none OTEL_METRICS_EXPORTER=none \
  pnpm exec nx run document:run --no-tui > /tmp/doc.log 2>&1; echo "exit=$?"

grep -aoE "instrumentation-[a-z-]+ Applying" /tmp/doc.log | sort -u          # which patched
grep -aiE 'Code.extend|is not a function|No "exports" main|Cannot find module' /tmp/doc.log  # MUST be empty
grep -aiE "Starting Nest application|Unable to connect to the database" /tmp/doc.log         # booted?
```

> ⚠️ zsh gotcha: write `nx run "${app}:run"`, not `nx run $app:run`. In zsh, `$app:run` triggers the `:r`
> modifier (strip extension) and mangles `document:run` → `documentun` ("Cannot find configuration for task").

- **Good run** = no crash lines; reaches "Starting Nest application…"; for document, `pino`/`pg`/`kafkajs`/
  `grpc`/`nestjs-core`/`express`/`aws-sdk` all show `Applying`. For search-worker: `pino`/`kafkajs`/
  `nestjs-core`/`express`.
- **Non-zero exit is expected** — the app dies in the DB/Kafka connect-retry loop (no local DB/broker). That
  is not a build/instrumentation failure.
- **Static safety check (optional but valuable):** parse the emitted `require("…")` names out of `main.cjs`
  and confirm each resolves from `apps/<app>/dist/main.cjs`; a non-empty "unresolvable" list means an editor
  transitive leaked out as a broken external → add it (or its bundled importer) to `keepBundled`. This is how
  Attempts 3/4 were caught.

> Full end-to-end log *record* emission can't be observed locally (no DB → the app exits before NestJS
> flushes pino; Nest uses `bufferLogs: true`, flushed only after `NestFactory.create()` resolves +
> `app.useLogger(...)`). The instrumentation **patch firing** is the deterministic proof the bridge is in
> place; records follow against a real DB + collector.

---

## 9. Environment / constraints (all verified)

- Monorepo: **Nx + pnpm** (workspace catalog), **rspack** for the NestJS apps; `@nx/rspack` **22.7.5**.
- pnpm uses the **isolated** node-linker (default; no hoist in `.npmrc`). App direct deps are symlinked into
  `apps/<app>/node_modules`; transitive deps live nested in `node_modules/.pnpm`. **A transitive package is
  NOT a flat entry in the app's `node_modules`** — this is why §7 #1 bites and why directory/resolver
  heuristics (§5 Attempts 3–4) are unreliable.
- **Node 25.x** — executes `.ts` via type-stripping and has strict `require(ESM)` interop (the lens through
  which blocknote's `.cjs` breaks).
- `tsconfig.base.json` has `"customConditions": ["@nx/source"]` (TS-solution setup).
- Build runs from the **workspace root** cwd, but **runtime `require()` resolution is anchored at the built
  `main.cjs` location** (`apps/<app>/dist/`). Externalized deps resolve via the app's reachable
  `node_modules`. Attempt 1 verified the broad set resolves at runtime in this repo.
- Runtime image is handled: `NxAppRspackPlugin({ generatePackageJson: true })` emits a pruned
  `dist/package.json`; `prune` / `copy-workspace-modules` targets ship `node_modules` for Docker. The
  generated manifest lists the app's **direct** deps (e.g. `@blocknote/core`, `pino-http`, `@nestjs/*`); their
  transitives are installed by pnpm from that manifest.

---

## 10. The final config (annotated)

`apps/document/rspack.config.ts` and `apps/search-worker/rspack.config.ts` (same shape; search-worker's
`keepBundled` is identical — it also pulls blocknote via `@notopia-uit/lib-server`):

```ts
import { createRequire, isBuiltin } from 'module';
// …

// Bundle ONLY the blocknote editor cluster — packages that cannot survive a native
// require() (broken require-condition interop, or ESM-only / no CJS exports main),
// plus the workspace libs that import them and the tiny transpile runtimes.
// EVERYTHING ELSE is external (backend infra + blocknote's dual-package leaf deps),
// which keeps the bundle small AND lets OTel's require-in-the-middle patch pino/pg/
// kafkajs/grpc/nestjs/express at runtime.
const keepBundled: RegExp[] = [
  /^@blocknote\//,
  /^prosemirror-/,
  /^@handlewithcare\//,
  /^@tiptap\//,
  /^yjs$/, /^y-protocols(\/|$)/, /^y-prosemirror$/, /^lib0(\/|$)/,
  /^react$/, /^react-dom(\/|$)/, /^scheduler(\/|$)/, /^@floating-ui\//,
  /^@bufbuild\/protobuf/,
  /^@notopia-uit\//,
  /^tslib$/, /^@swc\/helpers/,
];

class ExternalizePlugin {
  apply(compiler: rspack.Compiler) {
    // Overwrite the externals Nx hard-assigned. Runs after NxAppRspackPlugin
    // because it sits later in the `plugins` array.
    compiler.options.externals = [
      ({ request }, callback) => {
        if (
          !request ||
          /^[./]/.test(request) ||           // relative/absolute
          request.startsWith('@/') ||         // alias
          request.startsWith('@database') ||  // alias
          request.startsWith('#/') ||         // package imports field
          isBuiltin(request)                  // node builtins (preset handles too)
        ) {
          return callback();                  // → bundle
        }
        return keepBundled.some((re) => re.test(request))
          ? callback()                                 // → bundle the editor cluster
          : callback(undefined, `commonjs ${request}`); // → externalize everything else
      },
    ];
  }
}
```

Plugin order (load-bearing): `NxAppRspackPlugin({ … externalDependencies: 'none' })` **first**, then
`new ExternalizePlugin()`, then the rest.

### 10.1 Bundling-only scaffolding still required

These exist purely because we bundle the editor cluster; keep them:

- `stub.js` — `NormalModuleReplacementPlugin(/file-type$/, …)` stubs `file-type` (ESM-only, pulled by a
  bundled dep).
- `inquire-shim.js` — `NormalModuleReplacementPlugin(/@protobufjs\/inquire/, …)` shims a dynamic
  `require(expr)` so it doesn't warn/break when bundled (`@bufbuild/protobuf` neighborhood).
- `IgnorePlugin` + the `lazyImports` set — drops optional peer deps (`mysql`, `oracledb`, `ioredis`,
  `class-validator`, etc.) that bundled CJS libs (`typeorm`, `@nestjs/*`) reference behind feature flags.
- `jsdom` is **not** in `keepBundled` (it externalizes by default) — §4.2.

### 10.2 The source-map fix (don't regress)

`apps/document` must **not** carry both `devtool: 'source-map'` and `new rspack.SourceMapDevToolPlugin({})` —
that double-emits a ~64 MB inline source map into `main.cjs` (§5.a). Use only `devtool` (external `.map`).

---

## 11. Files / facts reference

- `apps/document/rspack.config.ts`, `apps/search-worker/rspack.config.ts` — the `ExternalizePlugin` +
  `keepBundled` described in §10.
- `apps/*/src/otel.ts` — NodeSDK + `getNodeAutoInstrumentations()`; sets the enabled-instrumentations default
  (pino included).
- `apps/*/src/main.ts` — `import './otel'` first; `bufferLogs: true`; `app.useLogger(Logger)` after create.
- `apps/*/src/app.module.ts` — `LoggerModule` (nestjs-pino) with a `pino-pretty` stream; `OpenTelemetryModule`
  (nestjs-otel) for metrics.
- Nx externals logic: `@nx/rspack/src/plugins/utils/apply-base-config.js` (`config.externals = externals`
  overwrite; the `externalDependencies` branches).
- Related already-merged OTel work (don't redo): `fix(otel): connect Go traces, emit gin/grpc metrics, and
  enable web OTLP logs`; `fix(document,search-worker): externalize npm deps so OTel can instrument them`;
  `chore: move type express to dev deps in nestjs`; `chore(lib,ui,pb): use peerDependencies for host-provided
  singletons`.

---

## 12. Roads not taken / future directions

- **Proposal: stop bundling, swc transpile-only (`@nx/js:swc`).** NestJS backends are normally not bundled.
  Output your compiled `src/` (a few hundred KB), keep **all** `node_modules` external. Every package then
  loads as authored — no interop landmines, no `keepBundled` list, every instrumentation patches natively, and
  you can delete `stub.js` / `inquire-shim.js` / `IgnorePlugin`. Cost: a real build migration (touch
  `project.json` build target, Dockerfile, the TypeORM migrations path, `mise.toml`, `out-tsc`/`dist`
  assumptions). The runtime `node_modules` is already handled by `generatePackageJson` + prune. **This is the
  cleanest long-term answer** if/when someone wants to invest the migration; the current rspack design is the
  pragmatic in-place fix.
- **Upgrade `@blocknote/*` past a fixed `.cjs`.** If a future release ships a working require-condition build,
  blocknote could be externalized and `keepBundled` could shrink toward empty. Re-run the bare-node repro
  before relying on it.
- **`dependencies`/`devDependencies` split (optional, image size only).** Bundled-only packages
  (`@blocknote/*`, prosemirror, yjs, react…) could move to each app's `devDependencies` so
  `generatePackageJson` stops shipping them in the runtime image. Correctness-neutral (dead weight only);
  **externalized packages must stay in `dependencies`.** Skipped to keep the externals↔manifest coupling out
  of scope.
```
