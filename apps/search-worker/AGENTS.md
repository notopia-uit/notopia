# Search Worker (nx: `search-worker`)

Pure Kafka consumer → Meilisearch. No HTTP API. Small enough that `src/` is flat:
`app.controller.ts` (event handlers), `app.service.ts` (indexing), config, otel, errors.

## Topics consumed (`app.controller.ts`)

| Topic | Effect |
| --- | --- |
| `events.integration.note.note.created` | Insert note metadata (name, workspace) |
| `events.integration.note.note.updated` | Update name, folder, trashed state |
| `events.integration.note.note.deleted` | Delete from `notes` |
| `events.integration.document.document.committed` | Convert BlockNote blocks → markdown, index as content |

Content is only indexed on **commit**, not on edit — an uncommitted document is invisible
to search. See `apps/document/AGENTS.md`.

## Things worth knowing

- Payload types come from `@notopia-uit/api-share-gen`; add new event shapes in
  `api/share/components/schemas/`, not here.
- Upsert semantics — a partial document (`model.ts`: `Partial<NoteSearch> & { id }`) is a
  valid update, which is why the note-created and document-committed handlers can each
  write disjoint fields of the same record.
- The frontend never queries Meilisearch with the master key; it uses scoped tenant tokens
  minted by `internal/note` (`q_workspacesearchtoken.go`). Index-level filterable attributes
  must line up with the filters those tokens embed.
- `kafka-max-retry.exception-filter.ts` bounds retries — a handler that throws is retried,
  then dropped. Don't rely on a throw to reprocess later.
- `seed/` re-indexes from scratch against a running Meilisearch.
- Deletion tombstones (`tombstone.ts`): topics are separate, so a late `note.updated` /
  `document.committed` can arrive after `note.deleted` and upsert the note back.
  `note.deleted` marks a tombstone before deleting; upserts skip tombstoned ids and re-check
  after writing (deleting again if a tombstone appeared meanwhile). With
  `NOTOPIA_SEARCH_WORKER_REDIS_URL` set, tombstones live in Redis (`SET EX`), required when
  running more than one instance. Unset, an in-memory `lru-cache` is used (single instance,
  a warning is logged). TTL `..._TOMBSTONE_TTL_SECONDS` (default 86400) must exceed the worst
  consumer lag plus retries; `..._TOMBSTONE_MAX` caps the in-memory entries.
- Every Meilisearch write awaits `waitTask()`; a failed task throws `MeiliTaskError` so the
  Kafka retry path sees it.
- `app.service.spec.ts` covers the tombstone guards and task failure with a fake Meilisearch
  (no Meilisearch testcontainer for JS). `tmp/search-worker/tombstone.it.ts` is a manual
  check against the real local Meilisearch + Redis.
