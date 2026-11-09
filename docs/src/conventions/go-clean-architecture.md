# Go Clean Architecture + CQRS

Rules for every Go service (`internal/{service}`). Reference implementation: `internal/note`.

## Layers and dependency direction

```
controller ─► app ─► domain ◄─ infra
```

| Layer         | Owns                                                              | May import                     |
| ------------- | ----------------------------------------------------------------- | ------------------------------ |
| `domain/`     | Aggregates, value objects, domain events, repo + UoW interfaces   | stdlib only                    |
| `app/`        | Commands, queries, app models, ports (read models, outbound svcs) | `domain`                       |
| `infra/`      | Adapters: postgres, kafka, grpc clients, payment providers        | `domain`, `app`                |
| `controller/` | http, grpc, event, cron: translate and delegate, never decide     | `app` only, **never `domain`** |
| `component/`  | Cross-cutting config: event topics (all of them), validator       | as needed                      |

`domain` and `app` never import `infra` or `controller`. Wiring (google/wire) is the only place
that joins them.

## CQRS split

|                  | Command (write)                              | Query (read)                                       |
| ---------------- | -------------------------------------------- | -------------------------------------------------- |
| App file         | `cmd_*.go`, `commonhandler.Cmd[T]`           | `q_*.go`, `commonhandler.Query[Q, R]`              |
| Returns          | **`error` only**, no result, no domain model | **app model only**, never a domain model           |
| Goes through     | `domain.UnitOfWork` → aggregate → repo       | an `app` read-model port, bypasses aggregates      |
| Port impl        | `infra/persistence/pgrepo`                   | `infra/persistence/pgreadmodel`                    |
| SQL (`pgsqlc/`)  | `<aggregate>.sql`, locks via `for_update`    | `<aggregate>_read.sql`, `Read…` names, never locks |
| Transaction      | `uow.Execute` owns it, never in a handler    | none                                               |

- Write and read SQL are separate files and separate queries even when they look identical.
- A command needing an id for its response gets it from the controller (`ID` input), and the
  controller then answers with a query. Commands never "return" through a field on the struct.
- Something that must hand back a value (e.g. a provider session URL) is modeled as a query.
- Lock-aware lookups use one query with `FOR UPDATE -- :if @for_update` (sqlc dynamic filter),
  not a `Get`/`GetForUpdate` pair.

## Models

- **Domain model**: unexported fields, `NewX` (validates, emits events) / `UnmarshalX` (rehydrate,
  no validation). Mutators call `addEvent`; the repo `Save` drains events. Domain enums are
  `uint8` `iota` types with `IsValid()` and `String()`.
- **App model** (`app/model.go`): plain exported structs and its own enums. It **never embeds or
  aliases a domain type** and may diverge from the aggregate. It is the only shape that leaves
  `app` (query results, command inputs).
- Domain ↔ app conversion lives in `app/mapper.go` only, with explicit switches per enum. Handlers
  map domain → app on the way out and app → domain on the way in.
- A read-model port may return domain aggregates when the handler must run domain rules on them
  (e.g. resolving entitlements). They are converted before the handler returns.

## Mapping at every boundary

Each boundary has its own types and an explicit mapper; never reuse a wire type across layers.

| Boundary                    | Mapper                                                            |
| --------------------------- | ----------------------------------------------------------------- |
| HTTP/gRPC request → command | `controller/*/mapin.go`: wire → **app** types                     |
| query result → response     | `controller/*/mapout.go`: **app** → wire types                    |
| SQL row ↔ domain            | `infra/persistence/pgmapper`, shared by `pgrepo` and `pgreadmodel` |
| SQL enum / jsonb            | sqlc `overrides` (typed columns), `pgmapper` for enums            |
| domain/app → integration event | `infra` publisher, payload types from `api/share`              |

## sqlc

- Use `sqlc.arg('name')` / `sqlc.narg('name')`, never `$1`. Param names match column names.
- Runtime-optional filters use the dynamic-filter plugin (`-- :if @param`), see `note_read.sql`.
- Generated `pgsqlc` package is shared by `pgrepo` and `pgreadmodel`; neither exposes sqlc types
  outside `infra`.

## Events and topics

- All topics (published and consumed) are constants in `component/eventtopic.go`; never inline a
  topic string. Payload types are defined once in `api/share`.
- Domain events are drained in repo `Save`; integration events go through the transactional outbox.
- Notifications that must not fail the request are published after `uow.Execute` returns.

## Checklist for a new use case

1. Domain method + event (writes only).
2. `app/model.go` types it needs, `app/mapper.go` conversions.
3. `cmd_*.go` or `q_*.go` (+ read-model port method for queries), register in `app.go` and `wire.go`.
4. SQL in the matching write or `_read` file, then `nx run {service}:gen:sqlc`.
5. API spec first (`api/`, `proto/`), then `nx gen api` / `nx gen proto`; controller mapin/mapout.
6. `nx wire {service}`, `nx lint {service}`.
