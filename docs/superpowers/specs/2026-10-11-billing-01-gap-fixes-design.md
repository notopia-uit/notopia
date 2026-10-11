# Billing 01 — Pre-existing gap fixes

> Sub-project 1 of 4. Source of truth for all decisions: grilling session 2026-10-10/11
> (original notes `tmp/BILLING_SPEC.md`, Part 2 `[gap]` items).
> Order: 01 + 02 (parallel) → 03 → 04. Progress is tracked by the task checkboxes in each plan under `docs/superpowers/plans/`.

## Goal

Fix lifecycle bugs in note, document and search-worker that billing, auto-commit and the
retention crons depend on. No billing code here.

## Why first

- Auto-commit (sub-project 4) would loop forever on the `modified` bug.
- The revision purge and Stripe workspace-deleted handling need note deletion and workspace
  deletion events that are never emitted today.
- Notes created in the app may never persist their content.

## Findings (verified in code, 2026-10-11)

| # | Finding | Where |
| --- | --- | --- |
| F1 | No code inserts `documents` rows. `updateDataById` uses `repo.update` → 0 rows for a new note. Only `internal/notecreateseed/seed.gen.sql` inserts documents. | `apps/document/src/document/document.service.ts` |
| F2 | `updateDataById` hardcodes `modified: true`. Hocuspocus v4 fires `onStoreDocument` on **any** change, so `HocuspocusService.setModified(false)` after a commit triggers a store that sets the column back to `true`. | `document.service.ts`, `hocuspocus/hocuspocus.service.ts` |
| F3 | `IntegrationEventNoteDeleted` is declared and mapped to `events.integration.note.note.deleted`, but no `e_*` handler emits it. | `internal/note/app/integrationevent.go`, `component/eventtopic.go` |
| F4 | Nothing consumes `note.deleted`: document keeps `documents` + `revisions`, search-worker keeps the Meilisearch doc. | `apps/document/src/hocuspocus/hocuspocus.controller.ts`, `apps/search-worker/src/app.controller.ts` |
| F5 | No `workspace.deleted` integration event exists (only the internal domain event and the SSE workspace event). Billing needs it. | `internal/note/component/eventtopic.go` |

## Design

### G1 — Document upsert (F1)

`DocumentService.updateDataById(id, data)` → `upsertData(id, data, modified)`:
`INSERT … ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data, modified = EXCLUDED.modified`.
`onAuthenticate` already rejects ids that are not existing notes, so only real notes get rows.
`fetch` keeps returning `null` for a missing row (new empty doc).

### G2 — `modified` flag (F2)

Rule: Yjs metadata is the truth while a document is loaded; the DB column is the truth while it
is unloaded.

```ts
// Database extension store: mirror Yjs metadata, never hardcode true
store: async ({ documentName: id, state, document }) => {
  const modified = (document.getMap('metadata') as YDocMetadataMap).get('metadata')?.modified ?? false;
  await documentService.upsertData(id, state, modified);
},
```

```ts
// afterLoadDocument: push column into Yjs without triggering a store
afterLoadDocument: async ({ documentName, document }) => {
  const entity = await documentService.getById(documentName);
  document.transact(
    () => metadata.set('metadata', { modified: entity?.modified ?? false }),
    { source: 'local', skipStoreHooks: true } satisfies LocalTransactionOrigin,
  );
},
```

- `onChange` (client edit) keeps setting `modified: true` in metadata.
- `setModified(false)` after a commit keeps a normal origin, so a store runs and persists bytes
  plus `modified=false` together.
- `commitDocument` keeps setting the column `false` (covers commits of unloaded docs).

Hocuspocus facts used (v4.0.0, installed source + tiptap docs): store payload
`{clientsCount, document, documentName, instance, lastContext, lastTransactionOrigin}`;
`LocalTransactionOrigin { source: 'local'; skipStoreHooks?; context? }`;
`shouldSkipStoreHooks` returns true for `redis` and for `local` with `skipStoreHooks`.

### G3 — Emit `note.deleted` (F3)

`internal/note/app/e_notedeletedintegration.go`: on domain `NoteDeletedEvent`, publish
`IntegrationEventNoteDeleted{ID}` through the outbox. Same pattern as
`e_notecreatedintegration.go`; register in `app/wire.go`, `app/app.go` and
`controller/event/event.go`.

### G4 — Consume `note.deleted` in document (F4)

New Kafka handler in `apps/document` (`@EventPattern('events.integration.note.note.deleted')`):

1. Close live Hocuspocus connections for that document and unload it (no store).
2. `DELETE FROM documents WHERE id = $1` — revisions go via existing `onDelete: 'CASCADE'`.

Idempotent: deleting a missing row is a no-op.

### G5 — Consume `note.deleted` in search-worker (F4)

`@EventPattern('events.integration.note.note.deleted')` → `index.deleteDocument(id)` on the
notes index. Idempotent.

### G6 — `workspace.deleted` integration event (F5)

- `api/share/components/schemas/WorkspaceDeletedIntegrationEvent.yaml` `{ id: uuid }`,
  registered in `api/share/openapi.yaml`, `nx gen api`.
- `internal/note/app/integrationevent.go`: `IntegrationEventWorkspaceDeleted`.
- `app/e_workspacedeletedintegration.go` on domain `WorkspaceDeletedEvent`.
- Topic `events.integration.note.workspace.deleted` in `component/eventtopic.go` and the
  integration publisher mapping.
- Workspace delete stays **soft** (`deleted_at` only). No consumer yet (billing consumes it in
  sub-project 3).

## Out of scope

- Permanent deletion of soft-deleted workspaces and cascading cleanup.
- Any billing / entitlement code.

## Testing

| Item | Test |
| --- | --- |
| G1 | Unit: `upsertData` inserts when missing, updates when present. |
| G2 | Unit: store passes metadata `modified`; `afterLoadDocument` transaction uses `skipStoreHooks`. Manual: commit → column stays `false` after the follow-up store. |
| G3, G6 | Go unit test per handler: domain event in → integration event published with correct payload. |
| G4 | Unit: handler deletes row and closes connections; missing row no error. |
| G5 | Unit: handler calls `deleteDocument`. |
| E2E (manual, compose up) | Create note in web → edit → `documents` row exists. Permanently delete note → row and revisions gone, Meilisearch doc gone. Delete workspace → message on `events.integration.note.workspace.deleted` (Redpanda console). |

## Acceptance

- All tests above pass; `nx lint note document search-worker` clean.
- `apps/document/AGENTS.md` updated: store semantics, upsert, note-deleted consumer.
- `apps/search-worker/AGENTS.md` topic table updated.
