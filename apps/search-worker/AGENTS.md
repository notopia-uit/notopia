# Search Worker (nx: `search-worker`)

Kafka consumer → Meilisearch, with HTTP health listener. Flat `src/`:
`app.controller.ts` (event handlers), `app.service.ts` (indexing), config, otel, errors.

## Topics consumed (`app.controller.ts`)

| Topic | Effect |
| --- | --- |
| `events.integration.note.note.created` | Insert note metadata (name, workspace) |
| `events.integration.note.note.updated` | Update name, folder, trashed state |
| `events.integration.note.note.deleted` | Delete from `notes` |
| `events.integration.document.document.committed` | Convert BlockNote blocks → markdown, index as content |

Content indexed only on **commit**, not edit — uncommitted documents invisible
to search. See `apps/document/AGENTS.md`.

## Things worth knowing

- Payload types from `@notopia-uit/api-share-gen`; add new event shapes in
  `api/share/components/schemas/`, not here.
- Upsert: partial document (`model.ts`: `Partial<NoteSearch> & { id }`) is a
  valid update; note-created and document-committed handlers can each
  write disjoint fields of same record.
- Frontend never queries Meilisearch with master key; uses scoped tenant tokens
  minted by `internal/note` (`q_workspacesearchtoken.go`). Index-level filterable attributes
  must match filters embedded in tokens.
- `kafka-max-retry.exception-filter.ts` bounds retries — throwing handler retried,
  then dropped. Don't rely on throw to reprocess later.
- `seed/` (`nx run search-worker:seed`, or `node dist/seed.cjs`) ensures Meilisearch index and settings exist; idempotent, required in every environment. Host, master key and note search key UID required. UID must match `NOTOPIA_NOTE_MEILISEARCH_NOTE_SEARCH_KEY_UID`; newly created key values printed only outside production. Seed emits root span without auto-instrumentation to avoid capturing credentials.
- `/search-worker/health/live` and `/search-worker/health/ready` listen on `NOTOPIA_SEARCH_WORKER_HEALTH_CHECK_PORT` (default 28083), independent of Kafka startup. Readiness checks Meilisearch and consumer group join/heartbeat; goes down after 15 seconds without heartbeat or on consumer failures, recovers on group join or heartbeat.
- `fixtures/` (`nx run search-worker:fixtures`, after `fixtures:transform`) loads dev-only fake documents into index.
- Deletion tombstones (`tombstone.ts`): topics separate, so late `note.updated` /
  `document.committed` can arrive after `note.deleted` and upsert note back.
  `note.deleted` marks tombstone before deleting; upserts skip tombstoned ids and re-check
  after writing (delete again if tombstone appeared meanwhile). With
  `NOTOPIA_SEARCH_WORKER_REDIS_URL` set, tombstones live in Redis (`SET EX`), required for
  multiple instances. Unset, uses in-memory `lru-cache` (single instance,
  logs warning). TTL `..._TOMBSTONE_TTL_SECONDS` (default 86400) must exceed worst
  consumer lag plus retries; `..._TOMBSTONE_MAX` caps in-memory entries.
- Every Meilisearch write awaits `waitTask()`; failed task throws `MeiliTaskError` so
  Kafka retry path sees it.
- `app.service.spec.ts` covers tombstone guards and task failure with fake Meilisearch
  (no Meilisearch testcontainer for JS). `tmp/search-worker/tombstone.it.ts` manually
  checks against real local Meilisearch + Redis.