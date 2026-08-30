# Search Worker (nx: `search-worker`)

Pure Kafka consumer → Meilisearch. No HTTP API. Small enough that `src/` is flat:
`app.controller.ts` (event handlers), `app.service.ts` (indexing), config, otel, errors.

## Topics consumed (`app.controller.ts`)

| Topic | Effect |
| --- | --- |
| `events.integration.note.note.created` | Insert note metadata (name, workspace) |
| `events.integration.note.note.updated` | Update name, folder, trashed state |
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
