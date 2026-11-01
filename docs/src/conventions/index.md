---
title: Conventions
order: 1
---

# Conventions

Go services follow DDD with CQRS. The full rules are in
[`docs/ai/GO_CLEAN_ARCH.md`](https://github.com/notopia-uit/notopia/blob/main/docs/ai/GO_CLEAN_ARCH.md);
this page is a summary. The Note service (`internal/note`) is the reference implementation.

- **Layers**: `controller -> app -> domain <- infra`. `domain` is stdlib only, `controller`
  imports `app` only (never `domain`), wiring with google/wire is the only place that joins
  layers.
- **CQRS**: commands (`cmd_*.go`) return only an error and go through `UnitOfWork`, aggregate
  and repo; queries (`q_*.go`) return app models only and use a read-model port that bypasses
  aggregates. Write and read SQL are separate files.
- **Models**: domain models have unexported fields with `NewX` / `UnmarshalX`; app models never
  embed domain types. Enums are `uint8` iota types with explicit mappers at every boundary.
- **sqlc**: named args only, dynamic filters with `-- :if @param`.
- **Events**: topics are constants in `component/eventtopic.go`, payloads come from `api/share`,
  integration events go through the transactional outbox.
- **Contracts**: change `api/` or `proto/` first, then `nx gen api` / `nx gen proto`.
