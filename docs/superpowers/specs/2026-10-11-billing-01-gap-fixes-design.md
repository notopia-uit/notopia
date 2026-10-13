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
| F6 | `NoteCreatedDomainToIntegrationEventHandler` is wired but never registered in `controller/event/event.go`, so `events.integration.note.note.created` is never published (never was, per git history). | `internal/note/controller/event/event.go` |
| F7 | Permanently deleting a folder (`cmd_permanentlydeleteworkspaceitems.go`, `cmd_emptytrash.go`) deletes child notes by SQL cascade, so those notes emit no `NoteDeletedEvent`. | `internal/note/app/cmd_permanentlydeleteworkspaceitems.go`, `cmd_emptytrash.go` |

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

1. Add the id to an in-memory `deletedDocumentIds` set; the store callback skips ids in it
   (closing the last connection runs the pending store immediately, and the upsert would
   recreate the row). The entry is removed in `afterUnloadDocument`.
2. `closeConnections(id)`.
3. `DELETE FROM documents WHERE id = $1` — revisions go via existing `onDelete: 'CASCADE'`.

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

### G7 — Publish `note.created` (F6)

Register `NoteCreatedDomainToIntegrationEvent.Handle` on the event processor
(`cqrs.NewEventHandler`), like `NotifyWorkspaceRenamed`.

### G8 — Note events on folder delete (F7)

Before saving deleted folders, load their notes with
`noteRepo.GetRecursiveChildrenFromFolder(folderID, true)`, call `PermanentlyDelete` on each and
save them first, so every removed note emits `NoteDeletedEvent`. `EmptyTrash` deletes trashed
notes before trashed folders.

## Out of scope

- Permanent deletion of soft-deleted workspaces and cascading cleanup.
- Any billing / entitlement code.

## Testing

Only non-obvious behaviour is unit-tested; the rest is covered by the manual E2E.

| Item | Test |
| --- | --- |
| G2 | Unit: store writes metadata `modified` (not `true`); `afterLoadDocument` transacts with `skipStoreHooks`. |
| E2E (manual, compose up) | Create note in web → edit → `documents` row exists. Commit → column stays `false` after the follow-up store. Permanently delete note → row, revisions and Meilisearch doc gone. Delete workspace → message on `events.integration.note.workspace.deleted` (Redpanda console). |

## Acceptance

- All tests above pass; `nx lint note document search-worker` clean.
- `apps/document/AGENTS.md` updated: store semantics, upsert, note-deleted consumer.
- `apps/search-worker/AGENTS.md` topic table updated.
