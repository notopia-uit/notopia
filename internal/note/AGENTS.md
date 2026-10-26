# Note Service

Workspaces, folders, notes, links, tags, trash, publishing. DDD + CQRS.
Entrypoint: `cmd/note/`. HTTP spec: `api/note/`. gRPC: `proto/note/`.

## Layers

| Dir | Contains | Rules |
| --- | --- | --- |
| `domain/` | Aggregates (`workspace`, `folder`, `note`), repo interfaces, `UnitOfWork` | No imports from `app`/`infra`. Pure |
| `app/` | One file per use case + outbound service interfaces | Depends on `domain` interfaces only |
| `infra/` | Adapters implementing `domain`/`app` interfaces | Where postgres/kafka/grpc/meili live |
| `controller/` | `http/` (gin), `grpc/`, `event/`, `health/` | Thin — translate, don't decide |
| `errs/` | Typed errors | Mapped to HTTP status in `controller/http/err.go` |

## Aggregate convention

Unexported fields; a `NewX(...)` (validates, for creation) / `UnmarshalX(...)` (no
validation, for rehydration from DB) pair; mutators take a trailing `userID string` and
call `addEvent(...)`; `PopEvents()` drains. The repo `Save` is what publishes drained
events. See `domain/note.go`.

## `app/` file naming

| Prefix | Kind | Interface |
| --- | --- | --- |
| `cmd_*.go` | Write use case | `commonhandler.Cmd[T]` |
| `q_*.go` | Read model query — bypasses aggregates, hits `pgreadmodel` | `commonhandler.Query[Q,R]` |
| `e_*.go` | Event handler (domain event → integration event, or inbound) | — |
| `svc_*.go` | Outbound service interface (implemented in `infra/service/`) | — |

Writes go through `uow.Execute(ctx, func(r domain.RepoRegistry) error {...})`, which owns
the transaction. Do not open transactions in handlers.

## Adding a use case

1. Domain method + event in `domain/` if state changes shape
2. `app/cmd_foo.go` — struct, `NewFooHandler`, `var ProvideFooHandler = NewFooHandler`,
   `type FooCmd commonhandler.Cmd[Foo]`, `var _ FooCmd = (*FooHandler)(nil)`
3. Register in `app/wire.go` and `app/app.go` (the `Server` struct aggregating handlers)
4. Path yaml in `api/note/paths/` + entry in `api/note/openapi.yaml`, then `nx gen api`
5. Implement the generated method in `controller/http/`, map in `mapin.go`/`mapout.go`
6. `nx wire note` if providers changed

## Persistence (`infra/persistence/`)

- `pgmigration/*.sql` — goose. New migration = new numbered file, never edit an applied one
- `pgsqlc/` — sqlc. Write the `.sql`, the `.sql.go` is generated alongside it
- `pgrepo/` — aggregate repos (write side); `pgreadmodel/` — denormalized reads (`q_*`)

## Events

`infra/outbox/` is a Watermill transactional outbox: integration events are written in the
same transaction as the domain change, then forwarded to Redpanda. This is why
`IntegrationPublisher.Publish` is safe to call inside `uow.Execute`.

Published integration topics: `note.created`, `note.updated`, `note.deleted`,
`workspace.deleted`. Every permanent delete (folder, note, trash) deletes the affected
notes through the note repo first, so each one emits its own `NoteDeletedEvent`.

`e_documentcommitted.go` is the inbound side — consumes the document service's commit event
and updates note size, tags and outgoing links.

## Entitlement gates (billing)

Billing calls (`app/svc_billing.go`) are uncached; unreachable billing → 503
`billingServiceUnavailable`. 402 carries `PaymentRequiredError` with `requiredPlan` (cheapest
tier that unlocks it); 409 `billingPayerMustTransfer`. Helpers in `app/billinggate.go`.

- `members` limit (`UpdateWorkspaceMembers`): blocks growth only, so an over-limit workspace
  after a downgrade can still remove members or change roles.
- Payer guard: the payer may not be demoted/removed/leave while billing says
  `requires_payer` (409); allowed for lifetime or cancelling subscriptions.
- `UpdateWorkspaceMembers` checks owner/manage permission first, so 403 wins over 402/409.
- `workspace_graph` feature → 402 when missing; note graph `depth` is clamped to the
  `note_graph_max_depth` limit (no error).
- TODO: gate `published_notes` when publishing is implemented (`cmd_publishnote.go` and
  `cmd_publishworkspace.go` are stubs; the limit lives in billing).

## Crons

`controller/cron/` registers `note:purge-expired-trash` (`PurgeExpiredTrashHandler`), env
`NOTOPIA_NOTE_CRON_PURGE_EXPIRED_TRASH_{NAME,SCHEDULE,TIMEOUT,ENABLED}`, default `0 3 * * *`,
runs with `time.Local`, locked via `pkgcron.Locker`. Per workspace the cutoff is the plan's
`trash_retention_days`, but never younger than 24h (`purgeMinTrashAge`); unlimited retention
and workspaces missing from the entitlements result are skipped (warn). Billing down → batch
skipped. Runs as the service-account user.

## Redis

`infra/common` owns the single Redis client, shared by the SSE hub (`workspaceevent`,
Redis streams) and the cron locker. Billing changes arrive as an event (`e_billingchanged.go`)
and are published to the workspace SSE channel as `WorkspaceBillingChangedEvent`
`{workspaceId, plan}`.

## Gotchas

- Trash read-model filters compare `trashed_by::text = sqlc.narg('trashed_by')::text` (the
  cast is needed for the nullable arg); keep it when editing those queries.
- `cmd_createworkspace.go` calls `authorizationSvc.CreateWorkspaceWithOwner` *inside*
  `uow.Execute` — a known cross-service side effect that can diverge if the commit fails
  (see the TODO there). Don't copy this pattern into new code; use the outbox.
- `q_workspacesearchtoken.go` mints scoped Meilisearch tenant tokens — search scoping is
  decided here, not in `search-worker`.
