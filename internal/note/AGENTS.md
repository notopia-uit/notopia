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

`e_documentcommitted.go` is the inbound side — consumes the document service's commit event
and updates note size, tags and outgoing links.

## Gotchas

- `cmd_createworkspace.go` calls `authorizationSvc.CreateWorkspaceWithOwner` *inside*
  `uow.Execute` — a known cross-service side effect that can diverge if the commit fails
  (see the TODO there). Don't copy this pattern into new code; use the outbox.
- `q_workspacesearchtoken.go` mints scoped Meilisearch tenant tokens — search scoping is
  decided here, not in `search-worker`.
