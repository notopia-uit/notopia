---
authorship:
  - ai
---

# NestJS 12 and Native ESM Migration Plan

**Research snapshot: 2026-11-21. Status: preparation, not an implemented migration.**

**Goal:** Manually upgrade NestJS after a compatible stable Nx release, then migrate document, AI,
and search-worker to native ESM without losing required telemetry or breaking deployment entrypoints.

**Scope:** A deferred implementation plan and supporting research. No package, application,
infrastructure, or build configuration changes have been applied.

**Execution:** Follow the [task checklists](#8-manual-execution-order-and-acceptance-gates) after
satisfying the [release gate](#2-what-is-actually-waiting-on-nx).
Use `executing-plans` for inline work or `subagent-driven-development` if explicitly chosen.

This plan builds on [NestJS bundling and OTel](../../src/stories/nestjs-bundling-otel.md) and
[the BlockNote 0.55/container investigation](../../src/stories/nestjs-bundling-otel-2026-10-blocknote-055.md).
Those stories explain why we currently bundle the editor cluster but externalize backend
dependencies. Their historical recommendation to keep CommonJS is not a permanent prohibition on ESM.

## 1. The decision in plain language

**We can move `document`, `ai`, and `search-worker` to native ESM. OpenTelemetry does not forbid it.**
What cannot move unchanged is our current telemetry startup arrangement.

Think of auto-instrumentation as attaching a recorder to a library when Node loads it:

```text
Current CJS bundle:
  start OTel → require backend libraries → OTel attaches → start Nest

Naively changed ESM bundle:
  load external imports → run bundled OTel setup → too late to attach reliably

Intended ESM startup:
  install ESM hook + start OTel outside the app → load app imports → start Nest
```

There are **two separate requirements**:

1. **Load order:** telemetry must start before the application imports the libraries it instruments.
   Moving `import './otel'` to the first source line does not guarantee this after ESM bundling.
2. **Loader coverage:** CommonJS and ESM have different loading mechanisms. Our CJS interception
   alone does not cover every ESM import. OTel provides an ESM hook for that.

There is also a **package-specific support limit**, not an ESM-wide limit:
`@opentelemetry/instrumentation-nestjs-core@0.72.1` declares support for Nest **`>=4.0.0 <12`**.
An ESM hook does not expand that supported range. Our default instrumentation lists do not enable
`nestjs-core`, so this does not by itself block the upgrade, but environments enabling it need an
explicit decision rather than silently accepting missing Nest-specific spans.

The destination is native ESM, in two separately verifiable stages:

- **Stage A:** manually upgrade Nest and related packages after a compatible stable Nx release.
  Keep CJS output if needed as an intermediate state; validate telemetry even in that state.
- **Stage B:** move all three services to ESM with a separate telemetry preload, correct externals,
  and coordinated Nx/container/deployment commands. Switch only after runtime acceptance checks.

A successful build is not evidence that logs and traces still arrive.

## 2. What is actually waiting on Nx?

[nrwl/nx#36937](https://github.com/nrwl/nx/pull/36937), “support NestJS 12,” merged on
2026-11-04. At this snapshot, both this workspace and npm's stable `nx`/`@nx/nest` release are
**24.0.6**. Published `@nx/nest@24.0.6` still peers:

```text
@nestjs/common >=10.0.0 <12.0.0
@nestjs/core   >=10.0.0 <12.0.0
```

The upstream migration is registered at **`24.1.0-beta.0`**. Treat **24.1.x as the expected release
line**, not as proof that a particular stable version is already available.

Before starting, verify that the selected stable release includes the PR and its `@nx/nest` peers
accept Nest 12. Upgrade `nx` and every direct `@nx/*` package together to that exact release:
`docker`, `js`, `nest`, `next`, `node`, `oxlint`, `react`, `rspack`, `vite`, `vitest`, and `web`.
Do not use a prerelease merely to unblock this plan.

### Read Nx's migration; do not run it

The references at the supplied commit are:

- [Migration guide](https://github.com/nrwl/nx/blob/aee9993e238f96c372a6b99eafce2ba8c706ce17/packages/nest/src/migrations/update-24-1-0/upgrade-to-nest-12.md).
- [Deterministic migration implementation](https://github.com/nrwl/nx/blob/aee9993e238f96c372a6b99eafce2ba8c706ce17/packages/nest/src/migrations/update-24-1-0/upgrade-to-nest-12.ts).
- [Paired source-review instructions](https://github.com/nrwl/nx/blob/aee9993e238f96c372a6b99eafce2ba8c706ce17/packages/nest/src/migrations/update-24-1-0/ai-instructions-for-nest-12.md).
- [Migration registration](https://github.com/nrwl/nx/blob/aee9993e238f96c372a6b99eafce2ba8c706ce17/packages/nest/migrations.json).

The implementation bumps declared Nest packages/add-ons, ensures TypeScript 6 for schematics, and
handles Jest's ESM requirements. Its minimum core target is `^12.0.1`, because 12.0.0 published
incorrect Nest 11 peers. It does **not** convert our Rspack configs or application output to ESM.
Its paired instructions explicitly say to retain the build setup and `@nx/js:node` targets.

Its automatic package scan starts from root `package.json`. Our manual inventory must also include
`pnpm-workspace.yaml`, app manifests, shared-package peers, generated-package peers, and tooling.
Do not assume an upstream package list is a complete inventory of this monorepo.

We will **not run Nx's internal Nest migration, its agentic step, or `nest upgrade`**. Apply reviewed
changes manually. Nx's new-project generator defaults and Nest CLI's new Rspack default do not
replace this repository's existing Nx/Rspack build pipeline.

## 3. Package targets and Renovate work

These are concrete versions available at the research snapshot, not an instruction to install
whatever `latest` means months later. The “current” column distinguishes catalog declarations
from resolved versions where that matters. Re-check peers and release notes at execution time.

### Required framework alignment

The lockfile currently resolves `@nestjs/common` and `core` to **11.2.9**, while
`microservices`, `platform-express`, and `testing` resolve to **11.2.11** under the `^11.2.7`
declarations. The upgrade should align the resolved core-family versions, not just their ranges.

| Package(s)                                                               | Current catalog declaration          | Proposed catalog target               | Reason / PR                                                                                                                                              |
| ------------------------------------------------------------------------ | ------------------------------------ | ------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `@nestjs/common`, `core`, `microservices`, `platform-express`, `testing` | `^11.2.7`                            | `^12.6.0` for each                    | Keep core-family packages aligned. [#350](https://github.com/notopia-uit/notopia/pull/350) currently resolves 12.6.0.                                    |
| `@nestjs/platform-socket.io`, `platform-ws`, `websockets`                | `^11.2.7`                            | `^12.6.0` for each                    | Align existing catalog entries; do not add unused adapters to apps. Also covered by #350.                                                                |
| `@nestjs/schematics`                                                     | `^11.1.0`                            | `^12.0.16`                            | Independently versioned; requires TypeScript 6. #350.                                                                                                    |
| `@nestjs/config`                                                         | `^4.0.4`                             | `^12.1.1`                             | Nest-12-compatible add-on; review config usage. #350.                                                                                                    |
| `@nestjs/typeorm`                                                        | `^11.0.3`                            | `^12.0.3`                             | v11's Nest peers exclude 12. Coordinate [#340](https://github.com/notopia-uit/notopia/pull/340), whose range is currently `^12.0.0`.                     |
| `@nestjs/terminus`                                                       | `^12.1.0`                            | `^12.2.0`                             | Already on major 12 and accepts Nest 11/12. [#396](https://github.com/notopia-uit/notopia/pull/396) is a maintenance update, not the Nest-major blocker. |
| `@nestjs/schedule`                                                       | `12.0.4`                             | Keep `12.0.4`                         | Published peers already accept Nest 11/12.                                                                                                               |
| `@nestjs/locks`                                                          | `0.0.6`                              | Keep `0.0.6`                          | Published peers already accept Nest 11/12. Check runtime locks and shutdown.                                                                             |
| `typescript`                                                             | `npm:@typescript/typescript6@^6.0.2` | Keep the TypeScript 6 alias initially | Already meets schematics' major requirement. Do not replace it blindly with Nx's example `~6.0.3`.                                                       |
| `@typescript/native`                                                     | `npm:typescript@^7.0.2`              | No change required by this migration  | Separate tool alias; not the `typescript` peer used by schematics.                                                                                       |

Do not set every package beginning with `@nestjs/` to the same version number: core packages,
schematics, and add-ons have independent releases.

### Database, testing, editor, and observability

| Package(s)                                                    | Current declaration / resolution      | Proposed action                                                          | Relationship                                                                                                                                                                                                                                |
| ------------------------------------------------------------- | ------------------------------------- | ------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `typeorm`                                                     | `^1.0.0`, already locked to **1.1.3** | Make baseline explicit with `^1.1.3`; retain 1.1.3 initially             | This is **not** a pending 0.3 → 1 migration. `@nestjs/typeorm@12.0.3` peers `^0.3.0 \|\| ^1.0.0-dev`, which admits stable 1.1.3.                                                                                                            |
| `typeorm-extension`                                           | `^3.9.0`, locked to 3.9.0             | `^4.6.0`, subject to CLI/fixture checks                                  | [#346](https://github.com/notopia-uit/notopia/pull/346) proposes `^4.0.0`. v3.9.0 peers `~0.3.0`, already mismatching our TypeORM; v4.6.0 peers `^1.1.3`. This is a related existing mismatch, not a Nest runtime dependency.               |
| `nestjs-pino`                                                 | `^5.3.1`, locked to 5.3.2             | Keep initially                                                           | Both 5.3.1 and 5.3.2 accept Nest 12. Retain Pino 10 / pino-http 11.                                                                                                                                                                         |
| `nestjs-otel`                                                 | `^8.1.1`                              | Keep initially                                                           | Accepts Nest `>=11 <13`; this does **not** prove OTel auto-instrumentation supports Nest internals.                                                                                                                                         |
| `@suites/di.nestjs`, `@suites/unit`, `@suites/doubles.vitest` | `~3.1.0`                              | Optional coordinated update to `~4.0.3`, `~4.1.3`, `~4.1.3` respectively | [#375](https://github.com/notopia-uit/notopia/pull/375). Adapter 3.1.0 already peers Nest `>=8`; not automatically required for Nest 12.                                                                                                    |
| All five catalog `@blocknote/*` entries                       | `0.55.0`                              | `0.57.0` together, separately validated                                  | [#238](https://github.com/notopia-uit/notopia/pull/238). ESM-only distribution plus a server DOM API change; not just a version bump.                                                                                                       |
| `yjs`                                                         | `^13.6.30`, locked to 13.6.33         | Raise declaration to `^13.6.33` with BlockNote                           | Matches BlockNote 0.57's peer floor. Stay on Yjs 13 with Hocuspocus; do not switch to `@y/y` 14.                                                                                                                                            |
| OTel API / auto-instrumentations / SDK                        | `~1.10.0` / `~0.83.0` / `0.301.2`     | No blind major bump                                                      | Review [#209](https://github.com/notopia-uit/notopia/pull/209) and [#354](https://github.com/notopia-uit/notopia/pull/354) separately, including matching core/resource/SDK peers. An SDK upgrade does not automatically repair load order. |
| `@openapitools/openapi-generator-cli`                         | `2.42.2`                              | Keep only after generation smoke test                                    | Internally pins `@nestjs/common` and `core` to 11.2.3 and `@nestjs/axios` to 4.0.1. Our global overrides currently force its Nest versions too.                                                                                             |

The overrides for `@nestjs/common` and `@nestjs/core` point to `catalog:`. Updating those catalog
entries changes the forced version workspace-wide, including code-generation tooling.
Preserve a single compatible Nest identity in the application graph, but do not mistake an override
for compatibility proof. If the generator breaks on forced Nest 12, investigate an updated tool or
a narrowly scoped application/tooling dependency boundary. Do not casually delete the overrides and
reintroduce the duplicate `ModuleRef` problem from the previous investigation.

[#350](https://github.com/notopia-uit/notopia/pull/350) alone leaves `@nestjs/typeorm` on 11; it is
therefore not a complete manual migration. Likewise, keep unrelated AI SDK major updates
[#387](https://github.com/notopia-uit/notopia/pull/387),
[#388](https://github.com/notopia-uit/notopia/pull/388), and
[#389](https://github.com/notopia-uit/notopia/pull/389) outside this work. BlockNote currently relies
on our AI SDK v6 line.

### Runtime and test-tool floors

- `mise.toml` and all three service Dockerfiles use **Node 26.13.1**. CI's shared setup reads the
  selected mise configuration via `node-version-file`; re-check the actual CI override at execution.
- Nest's documented CJS interoperability floor is Node 20.19 / 22.12, but that is **not this
  workspace's full toolchain floor**. Schematics 12.0.16 requires
  `^22.22.3 || ^24.15.0 || >=26.0.0`; Node 26.13.1 satisfies it.
- Existing Nest tests use Vitest. Do not add Jest merely because Nx's migration discusses it.
  If a Jest-tested Nest project is introduced, Nx's guide requires Node 24.9+, Jest 30.5+, and
  `--experimental-vm-modules`.
- Dockerfiles currently install pnpm **10.33.2**, while the workspace uses **12.14.0**.
  Revalidate frozen installation of the pruned lockfile; do not assume the workspace's successful
  pnpm install proves the container's install works.

## 4. The actual failure in the Nest 12 Renovate CI

The supplied [Main job](https://github.com/notopia-uit/notopia/actions/runs/44860430468/job/136712305428?pr=350)
ran against commit `c1f17f6de69475c30286a4e5a51a9fbd50af25b9` on 2026-11-21. Its reported failed task
was **`search-worker:typecheck`**, not a Rspack build:

```text
src/readiness.spec.ts:66:5 - error TS2353:
Object literal may only specify known properties,
and 'consumer' does not exist in type 'Promise<unknown>'.
```

The test currently uses:

```ts
vi.spyOn(server, 'createClient').mockReturnValue({
  consumer: () => consumer,
  producer: () => producer,
});
```

The upstream change explains the diagnostic:

| API                             | Nest 11.2.11                  | Nest 12.6.0                                         |
| ------------------------------- | ----------------------------- | --------------------------------------------------- |
| `ServerKafka.createClient<T>()` | Returns `T` synchronously     | Returns `Promise<T>`; dynamically imports `kafkajs` |
| `ServerKafka.listen()`          | Assigns `this.createClient()` | Awaits `this.createClient()`                        |

During implementation, change that mock to **`mockResolvedValue({ ... })`**, rather than adding a cast
or disabling typechecking. Re-run the readiness test and typecheck under the upgraded dependency
graph. Also review `TrackedKafkaServer` in `apps/search-worker/src/kafka.server.ts`: its retry,
consumer listener, and close overrides must retain their intended behavior with the new base class.

That job reports all three services' production builds and tests successful, and `ai:typecheck`
and `document:typecheck` successful. Tests transpiling successfully do not substitute for static
typechecking. Neither those successes nor this one diagnostic prove production startup, telemetry,
or a combined BlockNote/TypeORM upgrade is correct. This plan does not claim the mock change has
already been applied or tested.

Sources: [Nest 11 Kafka server](https://github.com/nestjs/nest/blob/v11.2.11/packages/microservices/server/server-kafka.ts)
and [Nest 12 Kafka server](https://github.com/nestjs/nest/blob/v12.6.0/packages/microservices/server/server-kafka.ts).

## 5. Nest 12 source-change audit

This inventory covers application/shared source on the current branch; it is not a claim that every
future dependency combination is verified.

| Change                                      | Repository finding                                                                                                                                                                                                                  | Manual upgrade action                                                                                                                                                                        |
| ------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Generic `ArgumentMetadata` / refined pipes  | No hand-written `ArgumentMetadata` or `PipeTransform` implementations found in the audited app/shared source. Generated headers decorators mention pipe support.                                                                    | Do not invent pipe edits. Typecheck the apps and generated API package; fix actual diagnostics.                                                                                              |
| Lifecycle hierarchy ordering                | gRPC wrappers assign service stubs in `onModuleInit`; app modules also initialize/log configuration, and document connects its Kafka client. `CronRegistrar` starts jobs in `onModuleInit`.                                         | Exercise startup and an early cron tick; ensure downstream stubs/connections are ready. Do not reorder unrelated hooks speculatively.                                                        |
| Shutdown hierarchy ordering                 | Document gateway closes connections in `onModuleDestroy`; Hocuspocus waits for pending stores in `beforeApplicationShutdown`, before TypeORM closes in `onApplicationShutdown`. Shared locks quit Redis in `onApplicationShutdown`. | Test SIGTERM with unsaved edits and a running locked job. Preserve final database stores and lock cleanup. Test AI's usage teardown and search tombstone/Kafka cleanup too.                  |
| `@Optional()` inheritance                   | No `@Optional()` use found in audited app/shared source.                                                                                                                                                                            | No blanket constructor rewrite. Check third-party injectable subclasses if DI errors appear.                                                                                                 |
| Config Standard Schema changes              | All three services use `registerAs(Symbol(...), ...)` with Zod `.parse()` inside factories, not Joi `validationSchema`/`validationOptions`.                                                                                         | Keep the factories and token conventions; verify invalid config fails startup. No Joi migration is needed.                                                                                   |
| Terminus legacy API removal                 | Already on Terminus 12. Search health uses `HealthIndicatorService`; no legacy `extends HealthIndicator`/`HealthCheckError` uses found in audited app/shared source.                                                                | Verify HTTP liveness/readiness and dependency failure/recovery; no automatic legacy-indicator rewrite.                                                                                       |
| NATS / GraphQL migrations                   | Not declared by the three apps; Kafka and gRPC are the relevant transports.                                                                                                                                                         | Do not install NATS or GraphQL replacements. An optional dependency name in Rspack's ignore set does not mean we use that transport.                                                         |
| `ConsoleLogger` structured params           | Production logging uses `nestjs-pino`, but framework bootstrap/fallback logging can still differ.                                                                                                                                   | Inspect bootstrap/error log shape and correlation; do not globally rewrite structured logger calls.                                                                                          |
| Express error mapping and graceful shutdown | HTTP gateway-user guards/error filters, AI SSE, and document's WebSocket upgrade share runtime behavior with the adapter.                                                                                                           | Verify HTTP error envelopes, auth status codes, streaming disconnects, and graceful draining. Do not change public contracts merely to adopt new opt-in Nest features.                       |
| Generated Nest server                       | `packages/api-document-nestjs-server` is CommonJS, exposes source via `@nx/source`, and has catalog Nest peers. Workspace packages are currently bundled into apps.                                                                 | Run API generation and inspect the diff; typecheck/build the generated package and document handlers. Do not hand-edit generated files or change API contracts for a build-format migration. |

**Upstream-instruction correction:** Nx's pinned AI instructions broadly describe `@Optional()` as no
longer inherited and suggest editing subclasses without constructors. The current Nest migration guide
clarifies that **12.1+ preserves inherited optional markers when the subclass does not declare its own
constructor**. The breaking case still relevant at our 12.6.0 target is a redeclared constructor without
its own optional marker. Follow the target version, not an obsolete blanket rewrite.

Native observability (`@nestjs/observe`), Standard Schema route decorators, route-specificity policies,
and new exception/error-code features are opt-in. They are not required here, and the new observability
SDK is not an automatic drop-in replacement for our OTel exporters and instrumentation.

## 6. What stays custom in Nx and Rspack?

The build is not owned by Nest CLI. `nx.json` registers **`@nx/rspack/plugin`**, which infers build
targets from each `rspack.config.ts`. Each app's `project.json` adds build options/configurations;
`run`/`debug` explicitly use **`@nx/js:node`**. Pruning uses
`@nx/js:prune-lockfile` and `@nx/js:copy-workspace-modules`. Docker targets depend on `prune` and `build`.

There is no need to replace these with new Nest CLI defaults. Inspect the resolved configuration with
`nx show project document --json` (and the other apps) after upgrading Nx; `project.json` alone is
not the entire target definition.

All three configs currently:

- Set `isEsm = false`, emit `main.cjs`, and use CommonJS externals.
- Pass `externalDependencies: 'none'` to `NxAppRspackPlugin`, then install `ExternalizePlugin`
  **after** it because Nx overwrites `externals`.
- Bundle the BlockNote/editor/Yjs/React/protobuf/workspace cluster, including dependencies imported
  by already-bundled third-party code.
- Force `jsdom` external to preserve runtime worker-file lookup.
- Keep backend packages external for instrumentation.
- Preserve SWC legacy decorators and decorator metadata, plus names during minification.

`document` also bundles `migrate.cjs`; `search-worker` bundles `seed.cjs`.
Document's allowlist additionally includes `@hocuspocus/*`. Do not flatten intentional differences
between the configs while migrating.

### Stage A: minimal changes, not “no telemetry work”

Keep output format, custom externals policy, pruning, and Nx node targets initially.
But Nest 12 itself now imports backend packages through ESM, even when our application emits CJS.

[OTel's ESM guidance](https://github.com/open-telemetry/opentelemetry-js/blob/main/doc/esm-support.md)
explicitly uses **Nest 12's Express adapter** as an example: without the ESM hook, Express imported
from that adapter may not be instrumented and HTTP spans can lose `http.route`.
The Kafka source comparison above also shows the move from `require('kafkajs')` to `import('kafkajs')`.
Check Kafka/gRPC/Pino/PG paths rather than assuming a `.cjs` suffix preserves their instrumentation.

If Stage A needs the ESM hook/preload to retain required telemetry, add that startup arrangement
before accepting Stage A. Do not postpone known telemetry loss until Stage B.

### Stage B: coordinated native ESM, not one Boolean

| Area                 | Required preparation                                                                                                                                                                                                                                                                     |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Output               | Emit `.mjs`, ESM chunks/runtime, and a separate telemetry preload. Re-check final Rspack options after Nx's plugin runs; use valid module chunk/loading settings, not the old CJS `array-push` assumptions.                                                                              |
| Externals            | Current callbacks return **`commonjs ${request}`**. Changing `output.module` alone leaves those requires. Use native module externals for packages intended to load through ESM, retaining deliberate CJS interop only where needed. Keep instrumented packages external.                |
| Bootstrap            | Preload telemetry before the app graph; register the OTel ESM hook. Do not initialize the SDK both in the preload and again inside `main`. Keep `reflect-metadata` available before decorated classes evaluate.                                                                          |
| Source               | All three `src/main.ts` files reference **`module.hot`**. Remove the unused CJS HMR branch or use a supported ESM-aware mechanism; HMR is currently disabled in the configs. Audit other CJS globals/dynamic requires.                                                                   |
| TypeScript / SWC     | Apps already use ESM source syntax; document uses `module: esnext` and `moduleResolution: bundler`. Keep decorator transforms. Do not apply the Nest CLI's NodeNext/extension rewrite wholesale to a bundled app. Separate unbundled native-Node entrypoints need Node-valid specifiers. |
| Nx                   | Ensure `run` and `debug` pass the loader/preload flags to the **child** Node process. Verify that `@nx/js:node` resolves the `.mjs` output correctly; if not, use an explicit Node run command with equivalent watch/debug semantics. Keep build/prune target relationships.             |
| Runtime dependencies | Declare any preload/hook packages as direct production dependencies of the apps using them. The hook imports must survive generated manifests, pruned lockfiles, and `--prod` installs.                                                                                                  |
| Docker               | Update all three Dockerfile commands and copy/build the preload alongside application output. Keep startup configuration consistent with Nx, not Docker-only.                                                                                                                            |
| Kubernetes           | Deployment templates for **all three services override the image command** with `node main.cjs`. Update them and their chart tests; changing Docker `CMD` alone is insufficient.                                                                                                         |
| Auxiliary commands   | Update document's migration hook (`job-migrate.yaml`, `migrate.cjs`) and search's seed hook (`job-seed.yaml`, `seed.cjs`), including hook tests. Audit compose/just/scripts and database CLI/fixtures commands.                                                                          |
| Shutdown             | Keep one SDK per process and flush it reliably. Document's migration runner explicitly receives `otelSdk`; preserve its shutdown contract without creating a second SDK in the preload.                                                                                                  |

The current upstream OTel documented startup shape is:

```sh
node --experimental-loader=@opentelemetry/instrumentation/hook.mjs \
  --import ./telemetry.mjs \
  ./main.mjs
```

This illustrates the **proposed emitted files**, not a command that works in today's checkout.
`--import` preloads our SDK initialization; the loader flag enables ESM interception. Merely passing
`--import @opentelemetry/instrumentation/hook.mjs` is not equivalent to registering that file as a
loader. OTel also discusses a `node:module.register()` direction; use the mechanism supported by the
chosen published OTel/Node versions and verify it in the real image.

Search's `seed/otel.ts` deliberately starts an SDK **without auto-instrumentation** to avoid
capturing credentials. Keep that smaller CLI policy; do not reuse the app's full telemetry preload
indiscriminately for seed commands.

### What does OTel support today?

The [instrumentation framework](https://github.com/open-telemetry/opentelemetry-js/tree/main/experimental/packages/opentelemetry-instrumentation)
documents ESM hooks but still labels its ESM instrumentation experimental. This is a supported path
to evaluate, not a blanket promise for every instrumentation/library/version combination.

Published package/source checks at this snapshot establish:

- **Pino instrumentation 0.71.0** supports Pino `>=5.14.0 <11` and explicitly handles an ESM module
  namespace/default export. Our Pino 10 line is in range. Prove both log sending and trace correlation.
- **Nest core instrumentation 0.72.1** excludes Nest 12. A loader cannot fix its version gate or
  guarantee compatibility with Nest's changed internal exports.
- **`nestjs-otel@8.1.1`** accepts Nest 12, but is a different integration from
  `@opentelemetry/instrumentation-nestjs-core`. Do not confuse their compatibility claims.
- Current defaults are `grpc,http,kafkajs,pino,pg,runtime-node` for document,
  `kafkajs,pino,runtime-node` for search, and `pino,runtime-node` for AI. Environment overrides may
  enable more. Neither `nestjs-core` nor `express` is in those defaults.

Therefore the ESM gate is **retaining the telemetry we require**, not “wait until all dependencies
become ESM,” and not “every OTel plugin must support Nest 12.” If an environment needs Nest-specific
auto-spans, wait for a compatible instrumentation release or explicitly design an alternative.
Do not widen its supported-version constant locally and call it supported.

Sources: [Pino instrumentation](https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/instrumentation-pino)
and [Nest instrumentation source](https://github.com/open-telemetry/opentelemetry-js-contrib/blob/main/packages/instrumentation-nestjs-core/src/instrumentation.ts).

## 7. BlockNote 0.57: related, but independently verifiable

[#238](https://github.com/notopia-uit/notopia/pull/238) upgrades 0.55 → 0.57 and confirms the upstream
ESM-only distribution change. **An ESM dependency does not require an ESM application bundle**:
Rspack can compile it into CJS. Therefore #238 does not itself force Stage B.

The old “broken published BlockNote `.cjs`” explanation ceases to describe 0.57, because that
distribution is gone. That does **not** mean all `keepBundled` rules can immediately disappear.
The policy also preserves Yjs/protobuf identity, bundles workspace source, and avoids undeclared
transitive runtime imports in pruned pnpm installations.

Start with the existing cluster policy, then remove individual exceptions only after proving they
are unnecessary. In particular:

- BlockNote 0.57's server-util requires an explicit DOM `Document` for DOM-dependent operations
  unless globals are already available. Audit `BlocknoteEditorService`,
  `AppService.handleDocumentCommitted`, document fixtures, and their tests. Markdown/HTML conversion
  must run successfully in Node, not just typecheck. Follow its new `@blocknote/server-util/dom` API
  and release notes when choosing DOM setup/lifetime.
- `jsdom` remains external initially. A changed DOM API does not prove worker-file packaging issues
  have vanished, nor justify deleting the direct dependency before runtime checks.
- Re-test both commit conversion and search markdown conversion using our custom schema, plus
  document fixture transforms and the browser editor. Keep BlockNote packages at one version.
- Review 0.56/0.57 editor UI changes separately; an ESM backend migration must not silently bundle
  unrelated version-history or emoji-picker behavior changes into “build cleanup.”

An eventual transpile-only/no-bundler design is another option described in the earlier investigation.
It would change source resolution, package publishing, and deployment substantially. It is not
required to achieve native ESM here; retain Rspack while changing startup and output deliberately.

## 8. Manual execution order and acceptance gates

### Task 1: Confirm the release gate and capture a baseline

**Files to inspect:** root `package.json`, `pnpm-workspace.yaml`, `pnpm-lock.yaml`, `mise.toml`,
`.github/actions/setup/action.yml`, `nx.json`, and each service's Dockerfile and `src/otel.ts`.

- [ ] Confirm the stable Nx release gate in §2. Record exact chosen package versions and PR revisions.
- [ ] Record each environment's enabled instrumentation and capture a known-good baseline of logs,
      spans, readiness, and shutdown. Include production overrides, not just repository defaults.
- [ ] Review the relevant Renovate diffs together; do not merge all open major upgrades as one batch.
- [ ] Use an isolated upgrade branch. Keep Stage A and Stage B changes separable for diagnosis/rollback.

**Gate:** A published compatible stable Nx version and a recorded telemetry baseline; otherwise stop.

### Task 2: Stage A — framework and source compatibility

**Files to change:** root `package.json`, `pnpm-workspace.yaml`, `pnpm-lock.yaml`, and
`apps/search-worker/src/readiness.spec.ts`. Change other files from the §5 audit only where evidence
requires it; review `apps/search-worker/src/kafka.server.ts` without assuming its overrides need edits.

- [ ] Update root Nx dependencies together and the **required framework alignment** targets from §3
      manually, including `@nestjs/typeorm` and the Terminus maintenance update. Keep schedule, locks,
      TypeScript, Pino, and `nestjs-otel` on the stated baseline. Make the existing TypeORM 1.1.3
      resolution explicit and coordinate `typeorm-extension` 4.6.0 with its own CLI/fixture checks.
      **Exclude BlockNote/Yjs (Task 4), optional Suites 4, OTel major updates, and AI SDK majors.**
      Inspect all affected manifests/peers and keep the application Nest graph coherent.
- [ ] Run `pnpm install`, inspect peer warnings and lockfile changes, then verify
      `pnpm install --frozen-lockfile`. Use `pnpm why -r @nestjs/common @nestjs/core typeorm` to inspect
      actual graph identities. Resolve peers rather than merely suppressing warnings.
- [ ] Confirm the Kafka mock diagnostic under Nest 12, change only that mock to `mockResolvedValue`,
      and run `nx typecheck search-worker --skip-nx-cache` plus
      `nx test search-worker --run --skip-nx-cache`. Expect no TS2353 and passing readiness recovery tests.
- [ ] Apply any additional source changes justified by §5. Test code generation
      with `nx gen api`; inspect generated diffs rather than hand-editing generated files.
- [ ] Run the checks below under the upgraded graph; build pruned images and run them directly.
- [ ] Validate CJS/ESM loader coverage. Add the supported hook/preload arrangement if needed to retain
      required telemetry. Do not accept a telemetry regression as the cost of a green typecheck.

**Gate:** Passing static/build checks, coherent dependency graph, working runtime workflows, and
required baseline telemetry retained. Commit the Stage A changes separately.

### Task 3: Stage B — native ESM and operational entrypoints

**Files to change:** `apps/{document,ai,search-worker}/rspack.config.ts`, `src/main.ts`, `src/otel.ts`,
`project.json`, `package.json`, and `Dockerfile`; add separate telemetry preload entries as needed.
Also update `apps/document/src/migrate.ts`, `apps/search-worker/seed/{main,otel}.ts`, deployment
templates under `deploy/k8s/charts/{document,ai,search-worker}/templates/`, their `Chart.yaml` versions,
and corresponding chart tests. Inspect the resolved Nx targets before choosing explicit child-process
runtime flags. Follow `deploy/k8s/AGENTS.md` for chart generation and release rules.

- [ ] Implement the §6 output/externals/preload changes for all three services, retaining the editor
      cluster policy initially. Choose one startup contract and apply it to Nx, Docker, and Kubernetes.
- [ ] Keep separate telemetry policies for app, document migration, and search seed. Reconcile SDK
      ownership and SIGTERM cleanup, including calls currently importing `otelSdk`.
- [ ] Bump the SemVer `Chart.yaml` version for each changed chart before release: Flux's
      ChartVersion reconciliation depends on it. Regenerate chart schema/docs through the chart's
      `gen` target where needed; never hand-edit generated chart READMEs or schemas.
- [ ] Run the static/build checks and every applicable runtime acceptance probe below.
- [ ] Only simplify bundling exceptions after successful container probes. Native ESM is the goal;
      “zero custom config” is not an acceptance criterion.

**Gate:** All three apps and operational CLIs run from their real deployment commands with expected
logs/spans, readiness, and shutdown. Commit Stage B separately from Stage A.

### Task 4: Adopt BlockNote 0.57 as a separate dependency/API change

**Files to change:** `pnpm-workspace.yaml`, `pnpm-lock.yaml`, the server editor construction sites in
§7, and related DOM/editor/fixture tests. Review UI changes separately if their APIs are affected.

- [ ] Upgrade the five catalog BlockNote entries together and align the Yjs peer floor.
- [ ] Add/adapt DOM setup only for operations that need it; test conversion, fixtures, and custom schema.
- [ ] Run backend and frontend checks from §7 and below; verify pruned images and live edit/commit/index.
- [ ] Record which bundling exceptions are still necessary, without speculative removals.

**Ordering:** Task 4 can run before or after Task 3. Stage B need not wait for #238 if ESM works with
0.55; #238 need not wait for Stage B if the CJS-bundled path passes.

**Gate:** Verified Node DOM operations, browser editor behavior, and end-to-end commit/index flow.

### Static/build checks

Run these during the actual migration, not as proof supplied by this preparation plan:

```sh
nx run-many -t typecheck test lint -p document ai search-worker packages/lib packages/lib-server packages/pb packages/api-document-nestjs-server --skip-nx-cache
nx run-many -t build -p document ai search-worker --configuration=production --skip-nx-cache
nx run-many -t docker:build -p document ai search-worker --skip-nx-cache
nx run-many -t lint test -p document-chart ai-chart search-worker-chart --skip-nx-cache
```

Also exercise generated API build/generation, document database CLI/fixtures, and frontend
`web`/`test-editor` typechecks/builds when adopting #238. Ensure applicable project targets exist
in the selected Nx release before relying on a run-many summary.

### Runtime acceptance: the checks that actually permit the switch

| Probe                                                   | Required outcome                                                                                                                                                                                                                         |
| ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Plain Node and pruned containers, not only Nx execution | No missing packages/subpaths, CJS globals, duplicate Yjs/protobuf instances, dropped worker assets, or preload resolution failures. ESM output imports are exercised by actual startup, not the older CJS `require()`-regex check alone. |
| Real telemetry collector                                | Pino log **sending** and trace-ID correlation both work. Exercise document HTTP/gRPC/PG/Kafka and search Kafka with the enabled instrumentation baseline. Check the AI defaults without claiming gRPC spans it never enabled.            |
| Additional instrumentation overrides                    | If Express is enabled, verify route attribution. If Nest core instrumentation is requested, treat its unsupported Nest 12 range as an unresolved gate, not an invisible pass.                                                            |
| Document workflow                                       | Authorized WebSocket edit, flush, commit, revision snapshot, commit event, and search indexing succeed with the custom BlockNote schema.                                                                                                 |
| Health / Kafka recovery                                 | Liveness stays up while dependencies retry; readiness reports failures and recovers on consumer group join/heartbeat and dependency recovery.                                                                                            |
| AI streaming                                            | Authorized generation streams, errors preserve the API envelope/stream rules, client cancellation releases leases, and no extra SDK initialization duplicates exports.                                                                   |
| Shutdown                                                | SIGTERM flushes logs/spans, pending document stores finish before DB teardown, jobs stop safely, Redis/Kafka clients close, and the process exits within the deployment grace period.                                                    |
| Operational CLIs                                        | Document migration completes and flushes telemetry; search seed remains idempotent and does not gain credential-capturing auto-instrumentation.                                                                                          |
| Kubernetes                                              | Rendered deployment and hook commands point to files actually present in the image and contain required startup flags/preloads. Chart tests pass.                                                                                        |

## 9. Rollback and what this plan does not prove

- If Stage B fails, revert its output/startup/deployment changes together and retain Stage A only
  if Nest 12's required telemetry already passes on CJS. CJS does not magically restore instrumentation
  of packages now loaded through Nest's ESM dependencies.
- If Stage A fails, restore its dependency declarations, lockfile, and associated source changes as
  a coherent set; redeploy the previous known-good images and commands.
- Keep database schema changes out of this framework/build migration. Do not use destructive
  schema sync or an unreviewed database downgrade as rollback.
- Do not merge unsupported peer overrides, disable typechecks, or accept missing logs merely to
  make an upgrade PR green.

This preparation verifies repository/source facts, published package metadata, and the cited CI
diagnostic. It does **not** claim a Nest 12/BlockNote 0.57/ESM image has been built or runtime-tested.
At execution time, update this plan with the selected Nx release and actual probe results.

## References

- [NestJS 11 → 12 migration guide](https://docs.nestjs.com/migration-guide)
  ([source](https://github.com/nestjs/docs.nestjs.com/blob/master/content/migration.md)).
- [Nx Nest 12 support PR](https://github.com/nrwl/nx/pull/36937); pinned migration sources in §2.
- [OTel ESM startup and Nest 12 dependency example](https://github.com/open-telemetry/opentelemetry-js/blob/main/doc/esm-support.md).
- [BlockNote 0.57 release notes](https://github.com/TypeCellOS/BlockNote/blob/v0.57.0/CHANGELOG.md)
  and [ESM-only upstream PR](https://github.com/TypeCellOS/BlockNote/pull/3118).
- [Original bundling/OTel investigation](../../src/stories/nestjs-bundling-otel.md),
  [0.55/container follow-up](../../src/stories/nestjs-bundling-otel-2026-10-blocknote-055.md), and
  [gRPC proto-loader investigation](../../src/stories/nestjs-grpc-proto-loader.md).
- Published metadata: [Nest TypeORM 12.0.3](https://registry.npmjs.org/@nestjs/typeorm/12.0.3),
  [TypeORM extension 4.6.0](https://registry.npmjs.org/typeorm-extension/4.6.0),
  [schematics 12.0.16](https://registry.npmjs.org/@nestjs/schematics/12.0.16),
  [Pino instrumentation 0.71.0](https://registry.npmjs.org/@opentelemetry/instrumentation-pino/0.71.0),
  and [Nest instrumentation 0.72.1](https://registry.npmjs.org/@opentelemetry/instrumentation-nestjs-core/0.72.1).
