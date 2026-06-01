# Why we bundle the NestJS apps as CJS, not ESM (document / search-worker)

> Conclusion doc. Companion to [`js-bundling-rspack.md`](./js-bundling-rspack.md). Read that first for the
> full OTel-logs background. This file records *why ESM output is a dead end* so nobody re-tries it.

## TL;DR

The NestJS apps (`apps/document`, `apps/search-worker`) are bundled with rspack to **CommonJS** (`main.cjs`,
`isEsm = false`). Switching the bundle to **ESM** does **not** solve the original problem (OTel can't patch
inlined deps). It makes it *worse*: the OpenTelemetry instrumentation stops firing entirely, and several
codebase constructs break. CJS is the correct, deliberate choice here.

## The decision in one sentence

OTel instrumentation in these apps works via **require-in-the-middle (RITM)**, which only intercepts CommonJS
`require()`. An ESM bundle has no `require` for external deps, so the patches never apply — and the fix
(register an ESM loader hook before the module graph loads) is incompatible with our `import './otel'`-first
bootstrap and unproven for `instrumentation-pino`.

## Background: what we actually need

- Goal: emit OTel **logs** from the NestJS apps. Logs require `@opentelemetry/instrumentation-pino` to **patch
  `pino` at runtime**. Patching only works if `pino` is **not bundled** (it must be a real, interceptable module
  load). That's why we externalize the instrumented packages (`pino`, `pg`, `kafkajs`, `@grpc/grpc-js`).
- The patch mechanism is **require-in-the-middle**: `otelSdk.start()` hooks `Module._load`, and every
  subsequent `require('pino')` / `require('pg')` / … is intercepted and wrapped.
- Bootstrap order is load-bearing (`apps/*/src/main.ts`):

  ```ts
  import './otel';            // otelSdk.start() → installs the RITM hook
  import '@notopia-uit/lib/yjs';
  import 'reflect-metadata';
  // …then NestFactory.create(...)
  ```

  Under CJS this works because the RITM hook is installed **synchronously first**, and the external deps are
  `require()`d **later** (during Nest bootstrap) → they get intercepted.

## Why ESM breaks it

### 1. RITM doesn't see ESM imports — you'd need import-in-the-middle (IITM)

In an ESM bundle, external deps are pulled with `import`, not `require`. RITM cannot intercept `import`. OTel's
ESM path uses **import-in-the-middle (IITM)**, which hooks Node's **ESM loader** via `module.register()` /
`--import`.

### 2. A loader hook registered from inside the graph is always too late

ESM resolves and **links all static imports before any module body executes** (depth-first, post-order
evaluation). A loader hook only affects imports resolved *after* it is registered. So by the time `otel.ts`'s
body runs `otelSdk.start()`, `pino` / `pg` / `@grpc/grpc-js` are **already linked** → nothing left to patch.

In a **bundled** ESM app this is structural, not incidental: the whole app is one file, and the external
`import pino` is a top-level import of that file, **hoisted above** the inlined `import './otel'` side effect.
The SDK necessarily starts *after* the externals are linked. Patching from within the graph is impossible.

The only ESM-correct fix is to register the hook **before the graph loads** at the Node CLI level
(`node --import @opentelemetry/instrumentation/hook.mjs`), which means:
- abandoning the `import './otel'`-first pattern,
- changing the `@nx/js:node` run/serve invocation and the Dockerfile entrypoint,
- and trusting `instrumentation-pino` to patch a **CJS `pino` through the CJS→ESM interop** under IITM — which
  is unproven for this stack. RITM patching a real `require('pino')` is the deterministic thing the original
  investigation relied on; IITM here is a gamble.

### 3. Concrete ESM breakages already present in the code

- `apps/document/src/main.ts` uses `module.hot` / `module.hot.dispose(...)`. `module` is a **CJS-only** global;
  it does not exist in ESM.
- The app `package.json` is `"type": "module"`, but the build emits `.cjs` on purpose (the `.cjs` extension
  forces CJS regardless of `type`). Flipping to `.mjs` re-exposes every issue here.
- The suppressed `Critical dependency` rspack warnings (`typeorm`, `nestjs`, `express`, `app-root-path`,
  `load-esm`) are **dynamic `require(expr)`** inside bundled CJS deps. ESM has no `require`; rspack injects
  `createRequire` shims, but dynamic-expression requires in those packages are fragile.
- `reflect-metadata` ordering and NestJS decorator metadata (`emitDecoratorMetadata`) have additional sharp
  edges under ESM hoisting.
- rspack `experiments.outputModule` + `target: node` + node externals is far less battle-tested than the CJS
  path.

### 4. ESM doesn't even solve the blocknote crash

The blocknote interop crash (see `js-bundling-rspack.md` §4) only happens when blocknote is **externalized** and
Node loads its broken `dist/blocknote.cjs`. ESM's *only* marginal benefit would be: externalize blocknote and
have `import` resolve the `import` condition → `dist/blocknote.js` (ESM build) instead. But we don't need that —
we simply **keep blocknote bundled**, which fixes the crash under CJS already. So ESM solves nothing here that
CJS doesn't, while costing us §1–§3.

## Conclusion

Stay on **CJS** output for the NestJS apps. It preserves the RITM patching model that actually fires, keeps the
`import './otel'`-first bootstrap valid, and avoids the `module.hot` / dynamic-`require` / decorator-ordering
landmines. The remaining work is purely about **which packages to externalize vs bundle under CJS** (see
[`../tasks/fix-nestjs-bundling-otel.md`](../tasks/fix-nestjs-bundling-otel.md)) — not about the module format.

**Do not re-attempt ESM bundling for these apps** unless the OTel instrumentation strategy is first redesigned
around a `node --import` ESM loader hook and validated end-to-end for `pino` logs.
