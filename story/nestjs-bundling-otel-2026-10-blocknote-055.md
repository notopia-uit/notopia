# Follow-up (2026-10): blocknote 0.55, transitive bundling, and a container that actually boots

> Follow-up to [`nestjs-bundling-otel.md`](./nestjs-bundling-otel.md). Read that first for the _why_ of
> the CJS + `ExternalizePlugin` + `keepBundled` design. This doc only records what changed on top of it
> and what was verified.

---

## 0. TL;DR

| Question                             | Answer                                                                                                                                                     |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Did blocknote fix its broken `.cjs`? | **No.** `@blocknote/core@0.55.0` still throws `te.default.extend is not a function` under plain `require()`. The `keepBundled` hack is **still required.** |
| Upgraded?                            | `@blocknote/*` 0.50.0 → **0.55.0**; `@nestjs/*` 11.1.x → **11.2.7** (single version, enforced by `overrides`).                                             |
| Was the shipped design correct?      | **Not in Docker.** It only worked under `nx run` (the `@nx/js:node` executor). A plain `node main.cjs`, and the Docker image, crashed. Fixed.              |
| NestJS 12?                           | **Not yet** — see §6.                                                                                                                                      |

---

## 1. Re-test of the blocknote CJS build (§4.1 of the old story)

Bare-node repro in a scratch dir with `@blocknote/core@0.55.0` + `@blocknote/server-util@0.55.0`:

```
require('@blocknote/core')        → 💥 te.default.extend is not a function
require('@blocknote/server-util') → 💥 te.default.extend is not a function
exports["."] = { import: ./dist/blocknote.js, require: ./dist/blocknote.cjs }   (unchanged)
```

No upstream issue/PR/release note mentions fixing it (searched `TypeCellOS/BlockNote` issues, PRs, and
0.51–0.55 release notes). `@handlewithcare/prosemirror-inputrules` is still `0.1.4` (ESM-only). So
`@blocknote/*` + the prosemirror cluster **must stay bundled**.

---

## 2. Blocknote 0.50 → 0.55 migration

| Release | Change                                                                                                                                                                     | What we did                                                                                                                                                                                                                                                                                                    |
| ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 0.51    | Markdown ↔ HTML rewritten; ~100 deps (the `unified`/`remark`/`mdast`/`hast` stack) removed from core                                                                       | Nothing — `blocksToMarkdownLossy` (search-worker) and the seed markdown parsing still typecheck/test green                                                                                                                                                                                                     |
| 0.52    | **Yjs decoupled**: `collaboration` option must be wrapped in `withCollaboration()` from `@blocknote/core/yjs`; `yjs`/`y-prosemirror`/`y-protocols` became _optional_ peers | Wrapped options in `packages/ui/src/components/editor-core.tsx` and `apps/test-editor/src/app/app.tsx`. `CollaborationUser` moved `@blocknote/core/extensions` → `@blocknote/core/yjs` (`packages/lib/src/block-note/collab-user.ts`). Verified pnpm still links `yjs`/`y-prosemirror`/`y-protocols` into core |
| 0.53    | `@blocknote/shadcn` uses base-ui instead of radix                                                                                                                          | No code change needed; `web` + `test-editor` build                                                                                                                                                                                                                                                             |

Note: core also ships a `./y` entry for **Yjs 14** (`@y/y`, `@y/prosemirror`). We are on Yjs 13 / Hocuspocus,
so we use `./yjs`.

---

## 3. The latent bug: transitive deps of bundled packages were external

### 3.1 Symptom

The old static check (§8 of the old story) was run from the _workspace_, where pnpm + the Nx executor make
almost anything resolvable. Run against the built file directly:

```
node apps/document/dist/main.cjs   →  Error: Cannot find module 'orderedmap'
```

and in a simulated/real Docker install, 15 externals were unresolvable:

```
orderedmap rope-sequence w3c-keyname tabbable fast-equals fast-deep-equal use-sync-external-store
emoji-mart @emoji-mart/data async-mutex crossws kleur  (+ subpaths)
```

### 3.2 Cause

`keepBundled` matches by **request name**. `prosemirror-model` is bundled, but its own
`require('orderedmap')` is a different request that matches nothing → externalized. Under pnpm isolated,
`orderedmap` only exists in `.pnpm/prosemirror-model@x/node_modules/`, never in the app's top-level
`node_modules` (nor in the pruned Docker install, which only hoists the app's **direct** deps). This was
true on 0.50 as well — not caused by the upgrade.

### 3.3 Fix (`ExternalizePlugin`, both apps)

```ts
const keepExternal: RegExp[] = [/^jsdom(\/|$)/];
// ...
if (keepExternal.some((re) => re.test(request))) return callback(undefined, `commonjs ${request}`);
return contextInfo?.issuer?.includes('/node_modules/') || keepBundled.some((re) => re.test(request))
  ? callback()
  : callback(undefined, `commonjs ${request}`);
```

**"If the importer is already-bundled third-party code, bundle what it imports too."** Because only the
`keepBundled` cluster's node_modules code is ever _in_ the bundle, this gives the full transitive closure
of that cluster — not the one-level issuer regex that failed in old Attempt 3. Workspace libs
(`@notopia-uit/*`) resolve to `packages/…` (symlinks resolved), so their backend imports (`@nestjs/*`, …)
stay external and OTel-patchable.

`jsdom` is the one exception (old §4.2 — bundling drops `xhr-sync-worker.js`), so it is forced external and
added as a **direct dependency** of both apps so the pruned install puts it on the top-level path.

### 3.4 What got bundled additionally (measured from the source maps, old rule vs new rule, both on 0.55)

|                                       | `document`                                                                                                                                                                                                | `search-worker`                                                                   |
| ------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| `main.cjs`                            | 4.8 MB → **5.5 MB**                                                                                                                                                                                       | 4.3 MB → **4.9 MB**                                                               |
| node_modules packages in bundle       | 43 → 57                                                                                                                                                                                                   | 38 → 49                                                                           |
| **Newly bundled**                     | `orderedmap` `rope-sequence` `w3c-keyname` `tabbable` `fast-equals` `fast-deep-equal` `use-sync-external-store` `emoji-mart` `preact` `@parcel/runtime-js` `@swc/helpers` `async-mutex` `crossws` `kleur` | same minus `async-mutex` `crossws` `kleur` (those come from `@hocuspocus/server`) |
| **No longer bundled** (reverse)       | none                                                                                                                                                                                                      | none                                                                              |
| Externals (`require()` in `main.cjs`) | 38 → **22**, 0 unresolvable                                                                                                                                                                               | 25 → **12**, 0 unresolvable                                                       |

Build time (`nx run document:build --skip-nx-cache`, 2 runs each): old ≈ 11.9 / 12.8 s, new ≈ 13.7 /
11.6 s — **within noise.** All newly bundled packages are small editor/view leaf deps.

No OTel-instrumented package moved into the bundle.

---

## 4. Docker / Nx fixes (all pre-existing, found while verifying)

| Problem                                                                                                                                                                                                                                                                                                                                       | Fix                                                                                                                                                                           |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `docker:build` depended on `build` only (`prune` commented out: "It got nx error"). `dist/pnpm-lock.yaml` references `file:./workspace_modules/...` which was never copied → `pnpm install --frozen-lockfile` failed. **The image did not build.**                                                                                            | Re-enabled `"prune"` in `docker:build.dependsOn` (both `project.json`). The old Nx error does not reproduce on Nx 23.2.1.                                                     |
| `CMD ["node", "main.js"]` but the build emits `main.cjs`                                                                                                                                                                                                                                                                                      | `CMD ["node", "main.cjs"]` (both Dockerfiles)                                                                                                                                 |
| Pruned lockfile mixed `@nestjs/core@11.2.3` with `@nestjs/common@11.1.21` → `Cannot find module '@nestjs/common/decorators/http/sse-signal.decorator'`; after bumping, `@openapitools/openapi-generator-cli` (pins `@nestjs/common@11.1.17`) leaked a second copy → `Nest can't resolve dependencies of the TypeOrmCoreModule (…, ModuleRef)` | Catalog `@nestjs/*` → `^11.2.7`, plus `overrides: { "@nestjs/common": "catalog:", "@nestjs/core": "catalog:" }` in `pnpm-workspace.yaml` → exactly one version workspace-wide |
| `uuid` imported by `apps/document/src/document/document.service.ts` but not declared (phantom dep via the root `package.json`)                                                                                                                                                                                                                | Added `uuid` to `apps/document` deps. Root `package.json` and `apps/search-worker` (only `seed/transform.ts` uses it) moved `uuid` to `devDependencies`                       |

### 4.1 `.proto` files at runtime

> Deep dive: [`nestjs-grpc-proto-loader.md`](./nestjs-grpc-proto-loader.md). No `.proto` files are needed,
> but the `@grpc/proto-loader` _package_ is required eagerly by Nest.

Not needed. The gRPC clients (`apps/document/src/{note,authorization}/*.module.ts`) pass the ts-proto
generated `packageDefinition` (`NoteServiceService`, `AuthorizationServiceService` from `@notopia-uit/pb`)
directly — no `protoPath`, so `@grpc/proto-loader` is never used and no `.proto` file must be shipped.
If someone switches to `protoPath`, the `.proto` files must be copied into the image (`/proto/` is in
`.dockerignore`).

---

## 5. Verification (what "works" means now)

1. `nx run-many -t typecheck test -p packages/ui packages/lib packages/lib-server document search-worker test-editor` ✅
2. `nx run-many -t build -p web test-editor` ✅, `nx lint` ✅
3. `nx run-many -t docker:build -p document search-worker` ✅
4. Static resolve check **inside the container** (every `require("x")` in `/app/main.cjs` must resolve): 0 unresolvable for both.
5. Boot probe in the container (`--env-file apps/<app>/.env`, `OTEL_LOG_LEVEL=debug`, no local DB/Kafka):
   - `document`: reaches `ECONNREFUSED 127.0.0.1:5434` (Postgres) — expected; patched `grpc http kafkajs pg pino`.
   - `search-worker`: `Starting Nest application…` → Kafka connect retry — expected; patched `kafkajs pino`.
   - The patched set equals `OTEL_NODE_ENABLED_INSTRUMENTATIONS`' default in `apps/*/src/otel.ts`
     (`grpc,http,kafkajs,pino,pg,runtime-node`). The richer list in the old story came from the local
     env enabling more.

**The probe that matters is now the container one**, not `nx run`, because the Nx executor masks
resolution failures.

```bash
cat > tmp/document/resolve-check.cjs <<'EOF'
const fs=require("fs"),{createRequire,isBuiltin}=require("module");
const f="/app/main.cjs";const src=fs.readFileSync(f,"utf8");
const names=new Set([...src.matchAll(/require\("([^"./][^"]*)"\)/g)].map(m=>m[1]).filter(n=>!isBuiltin(n)));
const r=createRequire(f);const bad=[...names].filter(n=>{try{r.resolve(n);return false}catch{return true}});
console.log("externals:",names.size,"unresolvable:",bad.join(", ")||"none");
EOF
nx run-many -t docker:build -p document search-worker
for img in document search-worker; do docker run --rm -v $PWD/tmp/document/resolve-check.cjs:/check.cjs:ro $img node /check.cjs; done
```

---

## 6. NestJS 12 — not now

12.1.2 is out. Ecosystem peers mostly allow it (`@nestjs/config@12`, `@nestjs/typeorm@12`,
`nestjs-otel@8.1`, `nestjs-pino@5`), but:

- **All Nest 12 packages ship as ESM** (`"type": "module"`). Our CJS bundle would load them via
  `require(esm)`. That works on Node 25, but require-in-the-middle can't usefully patch a frozen ESM
  namespace, which breaks the premise of old §2. `nestjs-core`/`express` instrumentation is not in our
  default enabled list today, so the immediate loss is small. Still, it needs re-validating, and is
  probably the push toward the swc transpile-only / `--import` hook direction (old §12).
- `@nx/nest@23.2.1` peers `@nestjs/* <12`.
- `@openapitools/openapi-generator-cli` pins `@nestjs/common` 11.x internally. Our new `overrides` would
  force it onto 12. That needs testing, or scoping the override.
- The `typescript-nestjs-server` generator output (`packages/api-document-nestjs-server`) has not been
  checked against 12.

Treat it as its own migration, with the container probe in §5 as the acceptance test.
