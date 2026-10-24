# K8s 01 — App readiness

> Sub-project 1 of 4. See `2026-10-24-k8s-00-overview-design.md` for shared contracts.
> No dependencies. Blocks 02 and 03.

## Goal

Make every service runnable as Kubernetes workloads and hook Jobs: explicit `serve`,
`migrate` and `seed` commands, non-zero exit on failure, per-command config, health endpoints
on every service, OTLP on short-lived processes, and gateway claims that Istio and Traefik
can both map.

## Current state (verified)

- `cmd/{note,authorization,billing}/main.go` log the error and `return`, so **a failure exits 0**.
- `note` and `billing` run `goose.Up` inside `Server.Run` (`internal/note/server.go:62`,
  `internal/billing/server.go:59`). The startup probe already gates on `IsMigrationDone`
  (`internal/note/controller/health/health.go:38`, same in billing).
- `internal/billingseed/seed.go:64` also runs migrations itself.
- `authorization` has no goose migrations. The Casbin gorm-adapter creates `casbin_rule` itself.
- `cmd/authorizationseed` and `cmd/billingseed` are separate binaries with their own env prefixes
  (`NOTOPIA_AUTHORIZATIONSEED_*`, `NOTOPIA_BILLINGSEED_*`) and duplicated config structs.
- `commonconfig.Redis` has only `Addr` (`pkg/common/config/redis.go:3`), with no password or DB.
- `document`: `synchronize: false`, the `database/migrations/` folder is empty (dev uses `db:sync`),
  and `datasource.ts` globs `migrations/*.ts`. **There is no health endpoint.**
- `search-worker`: the seed runs via `tsx`, with dev fallbacks for host and master key
  (`notopiauit`), a hard-coded key UID `00000000-0000-4000-8000-000000000001`, and it prints the
  key value. **There is no health endpoint** (`NOTOPIA_SEARCH_WORKER_HEALTH_CHECK_PORT=28083` exists
  in `.env` but is unused).
- `web` hard-codes `ws://${API_URL}/document/ws/document`
  (`apps/web/app/(workspace)/workspace/[workspaceId]/note/[noteId]/page.tsx:16`).
- The Authentik `app-web` provider maps `openid`, `profile`, `email` and `picture`. `groups` comes
  from the default `profile` mapping. No `roles` claim exists. Traefik `headerMap` maps
  `X-Forwarded-Roles: roles`, which is therefore always empty.

## 1. Go CLI (cobra)

### `pkg/common/cli` (new)

Shared command scaffolding, used by every Go service. It imports nothing from `internal/`.

```go
// Injector matches the signature of a wire injector.
type Injector[T any] func(ctx context.Context) (T, func(), error)

type Runner   interface{ Run(ctx context.Context) error }
type Seeder   interface{ Run(ctx context.Context) error }
type Migrator interface {
    Up(ctx context.Context) error
    Down(ctx context.Context) error    // one step
    Status(ctx context.Context) error  // logs applied and pending versions
}

func NewRoot(name metadata.ServiceName, logLevelEnv string, cmds ...*cobra.Command) *cobra.Command
func NewServeCommand(inject Injector[Runner]) *cobra.Command
func NewMigrateCommand(inject Injector[Migrator]) *cobra.Command // migrate up|down|status
func NewSeedCommand(inject Injector[Seeder]) *cobra.Command
func Execute(root *cobra.Command)                                // never returns
```

- `NewRoot`: `SilenceUsage: true` and `SilenceErrors: true`. `PersistentPreRunE` calls
  `logging.FirstStart(logLevelEnv)`. The root has no `RunE`: running without a subcommand prints
  help and exits non-zero.
- `Execute`: `signal.NotifyContext(SIGINT, SIGTERM)` → `root.ExecuteContext(ctx)`. On error it
  logs with `slog.Error` and calls `os.Exit(1)`. The injector `cleanup` (which flushes OTel
  providers) always runs **before** `Execute` exits, because each command defers cleanup inside
  its own `RunE`.
- `migrate` and `seed` wrap the work in one root span (`cli.migrate.up`, `cli.seed`) using the
  global tracer that the injector set up through `otel.Global`.

### Per-service layout

```
cmd/note/
  main.go        func main() { cli.Execute(cli.NewRoot(ServiceName, "NOTOPIA_NOTE_LOG_LEVEL",
                     cli.NewServeCommand(InitializeServer),
                     cli.NewMigrateCommand(InitializeMigrator))) }
  wire.go        InitializeServer, InitializeMigrator
  wire_gen.go    (generated: nx wire note)
  metadata.go    unchanged
```

| Service       | Commands                   | Injectors                                |
| ------------- | -------------------------- | ---------------------------------------- |
| note          | `serve`, `migrate`         | `InitializeServer`, `InitializeMigrator` |
| billing       | `serve`, `migrate`, `seed` | `+ InitializeSeeder`                     |
| authorization | `serve`, `seed`            | `InitializeServer`, `InitializeSeeder`   |

- `internal/{svc}/wire.go` exposes `ServeProviderSet` (today's `ProviderSet`, renamed),
  `MigrateProviderSet` and `SeedProviderSet`. Each contains only what its command needs:
  migrate needs config, logging, otel and persistence, with no Gin, gRPC, Kafka or Redis.
- `internal/{note,billing}/infra/persistence`: a new `Migrator` (wraps `*goose.Provider`, and
  implements `cli.Migrator`, with `var _ cli.Migrator = (*Migrator)(nil)`). `Server.Run` no longer
  calls `RunMigrations`, and `Pg.RunMigrations` is removed. `IsMigrationDone` stays for the startup
  probe.
- `Server` already satisfies `cli.Runner` (`Run(ctx) error`). Add the `var _` assertion.
- Dockerfiles: `ENTRYPOINT ["./note"]` + `CMD ["serve"]`. nx `run` and `debug` targets pass `serve`.

### Seeds move into their services

- `internal/authorizationseed` → `internal/authorization/seed`
- `internal/billingseed` → `internal/billing/seed`. Its `RunMigrations` call is removed.
- `cmd/authorizationseed`, `cmd/billingseed` and their nx projects are deleted.
  `internal/*fixtures` and `cmd/*fixtures` stay (dev only, never released).
- nx targets on the service projects: `migrate` (`go run ./cmd/{svc} migrate up`) and `seed`.
  `justfile`: `seed-authorization` → `nx run authorization:seed`; `seed-billing` →
  `nx run billing:migrate && nx run billing:seed`.
- Existing goose nx targets (`goose:create`, `goose:validate`, …) stay. They are authoring tools.

## 2. Config per command

- `pkg/common/config`: a generic loader replaces the copy-pasted `NewConfig` bodies:

  ```go
  func Load[T any](validate *validator.Validate, v *viper.Viper, envPrefix, fileName string) (*T, error)
  ```

  It keeps the same order as today: env prefix, key replacer, optional config file, `defaults.Set`,
  unmarshal, log, validate.

- Each service config package defines one root struct per command, composed from section structs,
  all with the **same env prefix** (`notopia_{svc}`):

  ```go
  // internal/billing/config
  type StripeAPI struct { SecretKey commonconfig.Secret `…` }
  type Stripe    struct { StripeAPI `mapstructure:",squash"`; WebhookSecret commonconfig.Secret `…` }

  type ServeConfig   struct { General; Log; Server; Database; Kafka; Redis; Identity; Services; Stripe; Billing; Cron }
  type MigrateConfig struct { General; Log; Database }
  type SeedConfig    struct { General; Log; Database; Identity; Stripe StripeAPI; Billing }
  ```

  `billingseed.Catalog` merges into `config.Billing` (same fields: `currency`, `web_base_url`).

- The section providers wire wants (`*commonconfig.SQL`, `*config.Server`, …) are provided from
  each root, so downstream constructors don't change.
- The `NOTOPIA_BILLINGSEED_*` and `NOTOPIA_AUTHORIZATIONSEED_*` env vars become `NOTOPIA_BILLING_*`
  and `NOTOPIA_AUTHORIZATION_*`. Update the `.env` files.

### Redis

```go
type Redis struct {
    Addr     string              `default:"-" … validate:"required"`
    Username string              `default:"-" … validate:""`
    Password commonconfig.Secret `default:"-" … validate:""`
    DB       int                 `default:"0" … validate:"min=0"`
}
func (r *Redis) Options() *redis.Options
```

note and billing build their clients from `Options()`. Node services already take a URL, which
carries the password and DB.

## 3. Document

- **Initial migration**: run `nx run document:db:generate` against an empty database, which
  produces `database/migrations/<ts>-Init.ts`.
- `database/migrations/index.ts` exports `migrations = [Init…]` as an explicit array. Both
  `database/datasource.ts` (CLI) and the runtime migrator use it, so there's one source and no glob,
  which bundling would break.
- `src/migrate.ts` (new rspack entry, giving `dist/migrate.cjs`):
  1. `import './otel'` first
  2. build a `DataSource` from `createDatasourceOptions` + `migrations`
  3. `runMigrations({ transaction: 'each' })`, logging each applied migration with the pino logger
  4. `destroy()`, then `await otelSdk.shutdown()`, then exit `0`. Any error is logged, the SDK is
     shut down, and the process exits `1`.
- rspack: `entry: { main: './src/main.ts', migrate: './src/migrate.ts' }`. `[name].cjs` already
  produces `main.cjs` and `migrate.cjs`. The Dockerfile copies `dist/` unchanged.
- **Health**: `@nestjs/terminus`, on the main port (8082):
  - `GET /document/health/live`: process up
  - `GET /document/health/ready`: Postgres ping, plus Redis ping when `SCALING_ENABLED=true`
  - `GET /document/health/startup`: no pending migrations (`dataSource.showMigrations()`)
- `apps/document/AGENTS.md`: replace the "no migration files" note. Production now uses
  migrations, and `db:sync` stays dev only.

## 4. Search-worker

- **Seed**: `seed/seed.ts` becomes a rspack entry (`dist/seed.cjs`), loaded with `./otel` and
  flushed on exit like `migrate.cjs`.
  - `NOTOPIA_SEARCH_WORKER_MEILI_HOST` and `…_MEILI_API_KEY` are **required**, with no fallbacks.
  - The key UID comes from `NOTOPIA_SEARCH_WORKER_MEILI_NOTE_SEARCH_KEY_UID` (the current UID is
    the dev `.env` value). It must equal note's `NOTOPIA_NOTE_MEILISEARCH_NOTE_SEARCH_KEY_UID`.
  - The key **value** is printed only when `NODE_ENV !== 'production'`. In production it is read
    once with `GET /keys/{uid}` (runbook in 03).
  - `package.json` `seed` keeps working for dev through `tsx`.
- **Health**: an HTTP listener on `NOTOPIA_SEARCH_WORKER_HEALTH_CHECK_PORT` (28083), with
  `GET /search-worker/health/{live,ready}`. Ready checks the Kafka consumer is connected and
  Meilisearch is reachable. It runs as a Nest hybrid app: HTTP for health plus the existing
  microservice.

## 5. Web

- Replace the template with `process.env.DOCUMENT_WS_URL` (full URL:
  `ws://api.notopia.localhost/document/ws/document` locally, `wss://notopia-api…/document/ws/document`
  in the cluster). Declare it in `global.d.ts`, and add it to `.env`.
- Unused `NEXT_PUBLIC_*` entries in `apps/web/.env` are removed (nothing reads them).

## 6. Gateway claims (Authentik and Traefik)

### Authentik scope mapping

New entry in `deploy/compose/infrastructure/authentik-blueprints/app-web.yaml`, modeled on the
builtin `blueprints/system/providers-oauth2.yaml`:

```yaml
- model: authentik_providers_oauth2.scopemapping
  identifiers:
    scope_name: notopia
  attrs:
    name: 'Notopia: gateway header claims'
    scope_name: notopia
    description: 'String claims mapped to X-Forwarded-* headers by the gateway'
    expression: |
      import json
      return {
          "groups_json": json.dumps([group.name for group in request.user.groups.all()]),
      }
```

- Add the mapping to the `app-web` provider's `property_mappings`
  (`!Find [authentik_providers_oauth2.scopemapping, [scope_name, notopia]]`), and add `notopia` to
  the web's requested `scopes` (`apps/web/lib/auth.ts:24`).
- `request.user.groups` is the accessor in Authentik `main`. Verify it on **both** versions in
  use: local `2026.8.5` and homelab `2026.5.6`. If one needs `ak_groups`, use whichever both
  support.
- Production is configured by hand. The same mapping is listed in the 03 checklist.

### Traefik (local)

- `jwt-middleware.yaml`: `headerMap` uses `X-Forwarded-Groups: groups_json`. Drop the
  `X-Forwarded-Roles` mapping. Add `removeMissingHeaders: true`.
- New middleware `jwt-admin`: the same issuers and headerMap, plus
  `require: { groups: notopia-admins }` (the default `Identity.AdminGroup`).
- `dynamic/billing.yaml`: a new router `billing-admin`
  (`PathPrefix(/billing/admin)`, middlewares `cors`, `jwt-admin`). The `billing` rule excludes
  that prefix.

## 7. Docs to update

- `pkg/AGENTS.md`: `common/cli`, the `Redis` fields, `Load[T]`, and `GatewayUserAuth` (the headers
  come from Traefik locally and Istio in the cluster; `Groups` comes from `groups_json`).
- `deploy/compose/AGENTS.md`: the `groups_json` mapping and `jwt-admin`.
- `CLAUDE.md` structure: remove `cmd/authorizationseed`, `cmd/billingseed`,
  `internal/authorizationseed` and `internal/billingseed`.
- `internal/authorization/AGENTS.md` and `internal/billing/AGENTS.md`: the seed location and commands.

## Testing

- `pkg/common/cli`: a unit test that a failing `RunE` makes `Execute` exit 1. Use a subprocess
  test (`os/exec` re-running the test binary), not a mock of `os.Exit`.
- `commonconfig.Load[T]`: a table test where required fields of a smaller struct don't fail when
  sections of another command are missing.
- Existing seed tests move along with their packages.
- Manual: `go run ./cmd/billing migrate status`, `seed`, `serve`. Then `node dist/migrate.cjs`
  against a fresh database. Then `curl` each health endpoint.
