# Note Service

Workspaces, folders, notes, links, tags, trash, publishing. DDD + CQRS.
Entrypoint: `cmd/note/`. HTTP spec: `api/note/`. gRPC: `proto/note/`.

## Layers

| Dir | Contains | Rules |
| --- | --- | --- |
| `domain/` | Aggregates (`workspace`, `folder`, `note`), repo interfaces, `UnitOfWork` | No `app`/`infra` imports. Pure |
| `app/` | One file per use case + outbound service interfaces | Depends only on `domain` interfaces |
| `infra/` | Adapters implementing `domain`/`app` interfaces | postgres/kafka/grpc/meili live here |
| `controller/` | `http/` (gin), `grpc/`, `event/`, `health/` | Thin — translate, don't decide |
| `errs/` | Typed errors | Mapped to HTTP status in `controller/http/err.go` |

## Aggregate convention

Unexported fields; `NewX(...)` validates for creation / `UnmarshalX(...)` skips
validation for DB rehydration. Mutators take trailing `userID string` and
call `addEvent(...)`; `PopEvents()` drains. Repo `Save` publishes drained
events. See `domain/note.go`.

## `app/` file naming

| Prefix | Kind | Interface |
| --- | --- | --- |
| `cmd_*.go` | Write use case | `commonhandler.Cmd[T]` |
| `q_*.go` | Read model query — bypasses aggregates, hits `pgreadmodel` | `commonhandler.Query[Q,R]` |
| `e_*.go` | Event handler (domain event → integration event, or inbound) | — |
| `svc_*.go` | Outbound service interface (implemented in `infra/service/`) | — |

Writes use `uow.Execute(ctx, func(r domain.RepoRegistry) error {...})`, which owns
transaction. Don't open transactions in handlers.

## Adding a use case

1. Domain method + event in `domain/` if state changes shape
2. `app/cmd_foo.go` — struct, `NewFooHandler`, `var ProvideFooHandler = NewFooHandler`,
   `type FooCmd commonhandler.Cmd[Foo]`, `var _ FooCmd = (*FooHandler)(nil)`
3. Register in `app/wire.go` and `app/app.go` (`Server` struct aggregates handlers)
4. Path yaml in `api/note/paths/` + entry in `api/note/openapi.yaml`, then `nx gen api`
5. Implement generated method in `controller/http/`, map in `mapin.go`/`mapout.go`
6. `nx wire note` if providers changed

## Persistence (`infra/persistence/`)

- `pgmigration/*.sql` — goose. New migration = new numbered file; never edit applied one
- `pgsqlc/` — sqlc. Write `.sql`; `.sql.go` generated alongside
- `pgrepo/` — aggregate repos (write side); `pgreadmodel/` — denormalized reads (`q_*`)

## Events

`infra/outbox/` is Watermill transactional outbox: integration events written in
same transaction as domain change, then forwarded to Redpanda. Thus
`IntegrationPublisher.Publish` safe inside `uow.Execute`.

Published integration topics: `note.created`, `note.updated`, `note.deleted`,
`workspace.deleted`. Every permanent delete (folder, note, trash) deletes affected
notes through note repo first, so each emits own `NoteDeletedEvent`.

`e_documentcommitted.go` is inbound side — consumes document service commit event
and updates note size, tags and outgoing links.

## Entitlement gates (billing)

Billing calls (`app/svc_billing.go`) uncached; unreachable billing → 503
`billingServiceUnavailable`. 402 carries `PaymentRequiredError` with `requiredPlan` (cheapest
tier unlocking it); 409 `billingPayerMustTransfer`. Helpers in `app/billinggate.go`.

- `members` limit (`UpdateWorkspaceMembers`): blocks growth only; over-limit workspace
  after downgrade can still remove members or change roles.
- Payer guard: payer cannot be demoted/removed/leave while billing says
  `requires_payer` (409); allowed for lifetime or cancelling subscriptions.
- `UpdateWorkspaceMembers` checks owner/manage permission first, so 403 wins over 402/409.
- `workspace_graph` feature → 402 when missing; note graph `depth` clamped to
  `note_graph_max_depth` limit (no error).
- TODO: gate `published_notes` when publishing implemented (`cmd_publishnote.go` and
  `cmd_publishworkspace.go` are stubs; limit lives in billing).

## Crons

`controller/cron/` registers `note:purge-expired-trash` (`PurgeExpiredTrashHandler`), env
`NOTOPIA_NOTE_CRON_PURGE_EXPIRED_TRASH_{NAME,SCHEDULE,TIMEOUT,ENABLED}`, default `0 3 * * *`,
runs with `time.Local`, locked via `pkgcron.Locker`. Per-workspace cutoff is plan's
`trash_retention_days`, never younger than 24h (`purgeMinTrashAge`); unlimited retention
and workspaces missing from entitlements result skipped (warn). Billing down → batch
skipped. Runs as service-account user.

## Redis

`infra/common` owns single Redis client, shared by SSE hub (`workspaceevent`,
Redis streams) and cron locker. Billing changes arrive as event (`e_billingchanged.go`)
and publish to workspace SSE channel as `WorkspaceBillingChangedEvent`
`{workspaceId, plan}`.

## Gotchas

- Trash read-model filters compare `trashed_by::text = sqlc.narg('trashed_by')::text` (cast
  needed for nullable arg); keep when editing queries.
- `cmd_createworkspace.go` calls `authorizationSvc.CreateWorkspaceWithOwner` *inside*
  `uow.Execute` — known cross-service side effect can diverge if commit fails
  (see TODO there). Don't copy into new code; use outbox.
- `q_workspacesearchtoken.go` mints scoped Meilisearch tenant tokens — search scoping
  decided here, not in `search-worker`.