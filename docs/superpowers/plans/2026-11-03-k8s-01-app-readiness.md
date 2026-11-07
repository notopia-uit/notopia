# Kubernetes App Readiness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make services runnable and observable as Kubernetes workloads and hook Jobs.

**Architecture:** Separate Go serving, migration and seeding into independently injected commands; build standalone Node migration and seed entries. Expose dependency-aware health and shared gateway claims.

**Tech Stack:** Go, Cobra, Wire, goose, NestJS, TypeORM, Rspack, Authentik, Traefik, OpenTelemetry.

**Spec:** `docs/superpowers/specs/2026-10-24-k8s-01-app-readiness-design.md` and `docs/superpowers/specs/2026-10-24-k8s-00-overview-design.md`.

## Global Constraints

- No automatic migration in servers; Jobs fail nonzero and flush telemetry. Use service env prefixes and regenerate Wire via Nx.
- Preserve dev fixtures, goose authoring targets, and the existing document soft-deletion cleanup warning. Use expand/contract migrations.
- Never log production Meilisearch key material; `groups_json` is a string claim.

## Review Focus

- Failed command or injector: nonzero exit and cleanup once (Task 1).
- Missing unrelated credentials: migration still loads its own config (Tasks 2–3).
- Database, pending migrations or scaling Redis failure: correct health status (Task 4).
- Missing seed credentials or failure: nonzero and no secret leak (Task 5).
- Forged identity or non-admin JWT on admin routes: stripped or rejected (Task 7).

## File Structure

`pkg/common/cli/` owns command lifecycle; `pkg/common/config/` owns loading and Redis. Service config, Wire, persistence and seed packages own dependencies. `cmd/` owns entrypoints. Document and search-worker own Node entries and health. Compose owns local identity mappings; web owns its websocket URL.

### Task 1: Shared Go command lifecycle

**Files:** Create `pkg/common/cli/cli.go`, `pkg/common/cli/cli_test.go`; modify `go.mod`, `go.sum` if needed.

**Interfaces:** Produce `Injector[T] func(context.Context) (T, func(), error)`, `Runner`, `Seeder`, `Migrator`, `NewRoot`, `NewServeCommand`, `NewMigrateCommand`, `NewSeedCommand`, `Execute`, with signatures in spec §1.

- [ ] Add subprocess tests: failed command exits 1; missing subcommand exits nonzero; initialized failing command runs cleanup once. Assert migrate and seed produce their specified root spans.
- [ ] Run `go test ./pkg/common/cli -v`; expect failure before implementation.
- [ ] Implement Cobra lifecycle, SIGINT/SIGTERM context, deferred injector cleanup and spans per spec §1. Run `go mod tidy` if dependencies changed.
- [ ] Run `go test ./pkg/common/cli -v`; expect PASS.
- [ ] Commit: `git add pkg/common/cli go.mod go.sum && git commit -m "feat(cli): add service command lifecycle"`.

### Task 2: Command-specific config and Redis

**Files:** Modify `pkg/common/config/`, `internal/{note,billing,authorization}/config/`, seed config, relevant `.env` files and clients; test beside each loader.

**Interfaces:** Produce `commonconfig.Load[T any](*validator.Validate, *viper.Viper, string, string) (*T, error)`, `(*Redis).Options() *redis.Options`, and service-specific `ServeConfig`, `MigrateConfig`, `SeedConfig` as applicable.

- [ ] Add table tests proving env overrides optional files/defaults, migrate needs database but not Kafka/Redis/Stripe, seed needs Stripe API key but not webhook secret, and Redis options preserve username/password/DB.
- [ ] Run `go test ./pkg/common/config/... ./internal/note/config/... ./internal/billing/config/... ./internal/authorization/config/...`; expect new tests to fail.
- [ ] Implement loader in spec order, composed config roots and section providers. Merge billing seed catalog, migrate old seed prefixes and use Redis `Options()` in clients.
- [ ] Run the same tests and `nx lint note`, `nx lint billing`, `nx lint authorization`; expect PASS.
- [ ] Commit affected config, clients and env with `feat(config): split service command configuration`.

### Task 3: Go services, migrations and seeds

**Files:** Modify `cmd/{note,billing,authorization}/{main.go,wire.go,project.json,Dockerfile}`, `internal/{note,billing,authorization}/wire.go`, servers and persistence; move `internal/authorizationseed` to `internal/authorization/seed` and `internal/billingseed` to `internal/billing/seed`; remove only their obsolete cmd projects; update `justfile`, relevant AGENTS files and `CLAUDE.md`.

**Interfaces:** Consume Tasks 1–2. Produce `InitializeServer`, `InitializeMigrator` for note/billing, `InitializeSeeder` for billing/authorization, and corresponding service provider sets; injector types use Task 1 interfaces.

- [ ] Move existing seed tests; add tests for goose up/down/status, no migration during server or seed startup, and migration injection without unrelated config.
- [ ] Run `go test ./internal/note/... ./internal/billing/... ./internal/authorization/...`; expect the new tests to fail.
- [ ] Implement goose migrators and provider sets, retain `IsMigrationDone`, add explicit commands, update Docker ENTRYPOINT/CMD and Nx run/preview/debug/migrate/seed targets. Regenerate with `nx run note:gen:wire`, `nx run billing:gen:wire`, `nx run authorization:gen:wire`.
- [ ] Run the Go tests and each service's `nx lint`; against configured dependencies run billing migrate status, seed twice and serve. Expect PASS and idempotent seed.
- [ ] Commit affected command, service and documentation paths with `feat(services): separate serving migrations and seeds`.

### Task 4: Document migration and health

**Files:** Modify `apps/document/database/datasource.ts`, `apps/document/rspack.config.*`, `apps/document/src/`, `apps/document/package.json`, `pnpm-lock.yaml`, `apps/document/AGENTS.md`; create explicit migration array, generated Init migration, `apps/document/src/migrate.ts` and focused tests.

**Interfaces:** Export `migrations`; build `dist/migrate.cjs`; expose `/document/health/{live,ready,startup}` on port 8082.

- [ ] Test liveness, readiness failure on Postgres or enabled-scaling Redis failure, startup failure with pending migrations, and migration failure logging, shutdown and nonzero exit.
- [ ] Run `nx test document`; expect failure.
- [ ] Generate Init against empty disposable Postgres with `nx run document:db:generate`. Share its explicit array between CLI and runtime. Add OTel-first migration entry with per-migration logs, `transaction: 'each'`, datasource destruction and SDK shutdown on all paths. Add Terminus health and named Rspack entry.
- [ ] Run `pnpm install`, `nx test document`, `nx build document`, `nx lint document`; execute bundled migration twice against fresh Postgres and curl health endpoints. Expect PASS and no pending migration on rerun. Retain existing-data cleanup warning in docs.
- [ ] Commit with `feat(document): add migrations and health checks`.

### Task 5: Search worker seed and health

**Files:** Modify `apps/search-worker/seed/seed.ts`, `apps/search-worker/src/`, Rspack config, `.env`, `apps/search-worker/AGENTS.md`; add focused tests.

**Interfaces:** Build `dist/seed.cjs`; expose `/search-worker/health/{live,ready}` on `NOTOPIA_SEARCH_WORKER_HEALTH_CHECK_PORT` (28083). Require `NOTOPIA_SEARCH_WORKER_MEILI_HOST`, `_MEILI_API_KEY`, `_MEILI_NOTE_SEARCH_KEY_UID`.

- [ ] Test missing host/key/UID fails, production output excludes key, repeated seed preserves UID, readiness fails on Kafka disconnect or unreachable Meilisearch, and liveness is independent.
- [ ] Run `nx test search-worker`; expect failure.
- [ ] Bundle traced seed with shutdown on success/failure, preserve tsx dev seed, remove fallbacks and set UID equal to note's UID. Add Nest hybrid HTTP listener tracking actual Kafka connection state.
- [ ] Run `nx test search-worker`, `nx build search-worker`, `nx lint search-worker`; exercise bundled seed and curl both health endpoints with dependencies available/unavailable. Expect PASS.
- [ ] Commit with `feat(search-worker): bundle seed and expose health`.

### Task 6: Websocket runtime configuration

**Files:** Modify `apps/web/app/(workspace)/workspace/[workspaceId]/note/[noteId]/page.tsx`, `apps/web/global.d.ts`, `apps/web/.env`; add a focused test.

**Interfaces:** Consume `DOCUMENT_WS_URL` as a complete URL.

- [ ] Test configured `wss://example.test/document/ws/document` is passed unchanged; run the focused test and expect failure.
- [ ] Replace hard-coded URL, declare env type, configure local `ws://api.notopia.localhost/document/ws/document`, remove only unused `NEXT_PUBLIC_*` entries.
- [ ] Run `nx test web`, `nx lint web`, `nx build web`; expect PASS. Verify runtime evaluation is not inadvertently baked into client code.
- [ ] Commit with `feat(web): configure document websocket URL`.

### Task 7: Gateway claim contract

**Files:** Modify `deploy/compose/infrastructure/authentik-blueprints/app-web.yaml`, Traefik `dynamic/jwt-middleware.yaml` and `dynamic/billing.yaml`, `apps/web/lib/auth.ts`, `deploy/compose/AGENTS.md`, `pkg/AGENTS.md`.

**Interfaces:** `notopia` OAuth scope emits JSON-string `groups_json`; Traefik maps it to `X-Forwarded-Groups`, removes absent forwarded headers, and `jwt-admin` requires `groups: notopia-admins`.

- [ ] Add integration assertions: forged headers without JWT cannot access protected/admin paths; non-admin JWT gets 403 on admin; admin succeeds; normal JWT receives its actual groups. Run against local gateway and expect failure.
- [ ] Add scope mapping and web scope; verify group accessor on Authentik 2026.8.5 and 2026.5.6. Add exclusive admin router and middleware, drop roles mapping and enable `removeMissingHeaders`.
- [ ] Run gateway tests against local compose and `nx lint web`; inspect issued tokens on both versions. Expect PASS; mark either unavailable version as an explicit deployment blocker.
- [ ] Commit with `feat(auth): standardize gateway group claims`.
