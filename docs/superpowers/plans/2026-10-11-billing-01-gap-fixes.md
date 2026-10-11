# Billing 01 — Gap Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix document persistence and note/workspace deletion events so billing, auto-commit and retention can rely on them.

**Architecture:** Document store becomes an upsert that mirrors the Yjs `modified` metadata; note publishes `note.created`, `note.deleted` (including notes removed with folders) and `workspace.deleted` integration events through its existing domain-event → event-processor path; document and search-worker consume `note.deleted`.

**Tech Stack:** NestJS 11, TypeORM, Hocuspocus v4.0.0, Yjs, vitest + @suites; Go, Watermill cqrs, oapi-codegen (share types).

**Spec:** `docs/superpowers/specs/2026-10-11-billing-01-gap-fixes-design.md`

## Global Constraints

- Hocuspocus v4: `onStoreDocument` fires on any change; skip with transaction origin `{ source: 'local', skipStoreHooks: true }` (`LocalTransactionOrigin` from `@hocuspocus/server`).
- Topic names: `events.integration.note.note.created`, `events.integration.note.note.deleted`, `events.integration.note.workspace.deleted`.
- Kafka payload types only from `api/share` (`@notopia-uit/api-share-gen`, `pkg/api/share`), never hand-written.
- Workspace delete stays soft. No billing code.
- Skip obvious tests (CLAUDE.md "Superpowers"). Run `oxfmt <changed js files>` and `nx lint <project> --fix` before each commit. Conventional commits.

## Review Focus

- A note deleted while users have it open: the forced store on disconnect must not recreate the `documents` row (Task 2 test).
- Commit then idle: the column must stay `modified=false` after the follow-up store (Task 1 test).
- Folder with nested sub-folders and notes permanently deleted: every nested note gets `note.deleted` (Task 4 manual check).
- A document loaded from DB with `modified=true` (edited, never committed, server restarted): clients must see `modified: true` (Task 1 test).
- `note.deleted` delivered twice: both consumers stay idempotent (Tasks 2, 5 — delete of a missing row/doc is a no-op).

---

### Task 1: Document store upsert + `modified` mirror

**Files:**
- Create: `apps/document/src/hocuspocus/document-metadata.ts`
- Create: `apps/document/src/hocuspocus/document-metadata.spec.ts`
- Modify: `apps/document/src/document/document.service.ts` (`updateDataById` → `upsertData`)
- Modify: `apps/document/src/hocuspocus/hocuspocus.ts`

**Interfaces:**
- Produces: `readModified(doc: Y.Doc): boolean`, `applyStoredModified(doc: Y.Doc, modified: boolean): void`, `DocumentService.upsertData(id: string, data: Buffer, modified: boolean): Promise<void>`

- [ ] **Step 1: Write the failing test**

```ts
// apps/document/src/hocuspocus/document-metadata.spec.ts
import * as Y from 'yjs';

import { applyStoredModified, readModified } from './document-metadata';

describe('document metadata', () => {
  it('applies the stored flag with a store-skipping origin', () => {
    const doc = new Y.Doc();
    const origins: unknown[] = [];
    doc.on('afterTransaction', (tr) => origins.push(tr.origin));

    applyStoredModified(doc, true);

    expect(readModified(doc)).toBe(true);
    expect(origins).toEqual([{ source: 'local', skipStoreHooks: true }]);
  });

  it('reads false when metadata is missing', () => {
    expect(readModified(new Y.Doc())).toBe(false);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `nx test document -- src/hocuspocus/document-metadata.spec.ts`
Expected: FAIL, cannot resolve `./document-metadata`.

- [ ] **Step 3: Implement**

```ts
// apps/document/src/hocuspocus/document-metadata.ts
import type { LocalTransactionOrigin } from '@hocuspocus/server';
import { YDocMetadataMap } from '@notopia-uit/lib/yjs';
import type * as Y from 'yjs';

const metadataOf = (doc: Y.Doc) => doc.getMap('metadata') as YDocMetadataMap;

export const readModified = (doc: Y.Doc): boolean =>
  metadataOf(doc).get('metadata')?.modified ?? false;

export const applyStoredModified = (doc: Y.Doc, modified: boolean): void => {
  const origin: LocalTransactionOrigin = { source: 'local', skipStoreHooks: true };
  doc.transact(() => metadataOf(doc).set('metadata', { modified }), origin);
};
```

`document.service.ts` — replace `updateDataById`:

```ts
  async upsertData(id: string, data: Buffer, modified: boolean): Promise<void> {
    this.logger.debug({ id, dataSize: data.length, modified }, 'upsertData: saving');
    await this.repo.upsert({ id, data, modified }, ['id']);
    this.logger.log({ id, dataSize: data.length }, 'upsertData: saved');
  }
```

`hocuspocus.ts` — store and a new `afterLoadDocument` hook (keep `fetch` unchanged):

```ts
      afterLoadDocument: async ({ documentName, document }) => {
        const entity = await documentService.getById(documentName);
        applyStoredModified(document, entity?.modified ?? false);
      },
```

```ts
          store: async ({ documentName: id, state, document }) => {
            await documentService.upsertData(id, state, readModified(document));
          },
```

Imports: `import { applyStoredModified, readModified } from './document-metadata';`. `grep -rn updateDataById apps/document/src` must return nothing.

- [ ] **Step 4: Run tests**

Run: `nx test document`
Expected: PASS.

- [ ] **Step 5: Lint, format, commit**

```bash
oxfmt apps/document/src/hocuspocus apps/document/src/document
nx lint document --fix
git add apps/document/src
git commit -m "fix(document): upsert document state and mirror yjs modified flag"
```

---

### Task 2: Document consumes `note.deleted`

**Files:**
- Modify: `apps/document/src/hocuspocus/hocuspocus.ts` (deleted-id guard)
- Modify: `apps/document/src/hocuspocus/hocuspocus.service.ts` (`onNoteDeleted`)
- Modify: `apps/document/src/hocuspocus/hocuspocus.controller.ts`
- Modify: `apps/document/src/document/document.service.ts` (`deleteById`)
- Create: `apps/document/src/hocuspocus/hocuspocus.service.spec.ts`

**Interfaces:**
- Consumes: `NoteDeletedEvent` (`{ id: string }`) from `@notopia-uit/api-share-gen`.
- Produces: `Hocuspocus.markDeleted(id: string): void`, `DocumentService.deleteById(id: string): Promise<void>`, `HocuspocusService.onNoteDeleted(id: string): Promise<void>`

- [ ] **Step 1: Write the failing test**

```ts
// apps/document/src/hocuspocus/hocuspocus.service.spec.ts
import { TestBed, type Mocked } from '@suites/unit';

import { DocumentService } from '../document/document.service';
import { Hocuspocus } from './hocuspocus';
import { HocuspocusService } from './hocuspocus.service';

describe('HocuspocusService.onNoteDeleted', () => {
  let service: HocuspocusService;
  let hocuspocus: Mocked<Hocuspocus>;
  let documentService: Mocked<DocumentService>;

  beforeEach(async () => {
    const { unit, unitRef } = await TestBed.solitary(HocuspocusService).compile();
    service = unit;
    hocuspocus = unitRef.get(Hocuspocus);
    documentService = unitRef.get(DocumentService);
  });

  it('blocks stores before closing connections, then deletes the row', async () => {
    const calls: string[] = [];
    hocuspocus.markDeleted.mockImplementation(() => void calls.push('mark'));
    hocuspocus.closeConnections.mockImplementation(() => void calls.push('close'));
    documentService.deleteById.mockImplementation(async () => void calls.push('delete'));

    await service.onNoteDeleted('n1');

    expect(calls).toEqual(['mark', 'close', 'delete']);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `nx test document -- src/hocuspocus/hocuspocus.service.spec.ts`
Expected: FAIL, `onNoteDeleted` / `markDeleted` not defined.

- [ ] **Step 3: Implement**

`hocuspocus.ts` — add to the class and guard the store:

```ts
  private readonly deletedDocumentIds = new Set<string>();

  markDeleted(id: string): void {
    if (this.hocuspocus.documents.has(id)) {
      this.deletedDocumentIds.add(id);
    }
  }

  closeConnections(id: string): void {
    this.hocuspocus.closeConnections(id);
  }
```

In the server config:

```ts
      afterUnloadDocument: async ({ documentName }) => {
        this.deletedDocumentIds.delete(documentName);
      },
```

```ts
          store: async ({ documentName: id, state, document }) => {
            if (this.deletedDocumentIds.has(id)) return;
            await documentService.upsertData(id, state, readModified(document));
          },
```

`document.service.ts`:

```ts
  async deleteById(id: string): Promise<void> {
    await this.repo.delete({ id });
    this.logger.log({ id }, 'deleteById: document deleted');
  }
```

`hocuspocus.service.ts` — inject `DocumentService` (module already imports `DocumentModule`) and add:

```ts
  async onNoteDeleted(id: string): Promise<void> {
    this.hocuspocus.markDeleted(id);
    this.hocuspocus.closeConnections(id);
    await this.documentService.deleteById(id);
  }
```

`hocuspocus.controller.ts`:

```ts
  @EventPattern('events.integration.note.note.deleted')
  async handleNoteDeleted(@Payload() data: NoteDeletedEvent) {
    this.logger.log({ id: data.id }, 'handleNoteDeleted: received');
    await this.hocuspocusService.onNoteDeleted(data.id);
  }
```

(import `EventPattern` from `@nestjs/microservices`, `NoteDeletedEvent` from `@notopia-uit/api-share-gen`).

- [ ] **Step 4: Run tests**

Run: `nx test document`
Expected: PASS.

- [ ] **Step 5: Lint, format, commit**

```bash
oxfmt apps/document/src
nx lint document --fix
git add apps/document/src
git commit -m "feat(document): delete document on note.deleted"
```

---

### Task 3: Note publishes `note.created` and `note.deleted`

**Files:**
- Create: `internal/note/app/e_notedeletedintegration.go`
- Modify: `internal/note/app/app.go` (`Events` struct), `internal/note/app/wire.go`
- Modify: `internal/note/controller/event/event.go` (`setup`)
- Regenerate: `cmd/note/wire_gen.go` via `nx run note:gen:wire`

**Interfaces:**
- Consumes: `domain.NoteDeletedEvent` (`BaseEvent.AggregateID` = note id), existing `IntegrationEventNoteDeleted{ID}` and its topic mapping.
- Produces: `NoteDeletedDomainToIntegrationEventHandler.Handle(ctx, *domain.NoteDeletedEvent) error`

- [ ] **Step 1: Implement the handler**

```go
// internal/note/app/e_notedeletedintegration.go
package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type NoteDeletedDomainToIntegrationEventHandler struct {
	integrationPublisher IntegrationPublisher
}

func NewNoteDeletedDomainToIntegrationEventHandler(
	integrationPublisher IntegrationPublisher,
) *NoteDeletedDomainToIntegrationEventHandler {
	return &NoteDeletedDomainToIntegrationEventHandler{integrationPublisher: integrationPublisher}
}

var ProvideNoteDeletedDomainToIntegrationEventHandler = NewNoteDeletedDomainToIntegrationEventHandler

func (h *NoteDeletedDomainToIntegrationEventHandler) Handle(ctx context.Context, event *domain.NoteDeletedEvent) error {
	if err := h.integrationPublisher.Publish(ctx, IntegrationEventNoteDeleted{ID: event.AggregateID}); err != nil {
		return fmt.Errorf("failed to publish note deleted integration event: %w", err)
	}
	slog.InfoContext(ctx, "Note deleted integration event published", slog.String("note_id", event.AggregateID.String()))
	return nil
}
```

- [ ] **Step 2: Wire it**

`app/app.go` `Events` struct: add `NoteDeletedDomainToIntegrationEvent *NoteDeletedDomainToIntegrationEventHandler`.
`app/wire.go`: add `ProvideNoteDeletedDomainToIntegrationEventHandler` next to `ProvideNoteCreatedDomainToIntegrationEventHandler`.

`controller/event/event.go` `setup()`, after the `NotifyWorkspaceSlugChangedHandler` block:

```go
	if _, err := e.eventProcessor.AddHandler(cqrs.NewEventHandler(
		"NoteCreatedToIntegrationHandler",
		e.app.Events.NoteCreatedDomainToIntegrationEvent.Handle,
	)); err != nil {
		return fmt.Errorf("failed to add event handler NoteCreatedToIntegrationHandler: %w", err)
	}

	if _, err := e.eventProcessor.AddHandler(cqrs.NewEventHandler(
		"NoteDeletedToIntegrationHandler",
		e.app.Events.NoteDeletedDomainToIntegrationEvent.Handle,
	)); err != nil {
		return fmt.Errorf("failed to add event handler NoteDeletedToIntegrationHandler: %w", err)
	}
```

- [ ] **Step 3: Regenerate wire, build, lint**

```bash
nx run note:gen:wire
nx build note
nx lint note
```
Expected: all succeed.

- [ ] **Step 4: Commit**

```bash
git add internal/note cmd/note/wire_gen.go
git commit -m "fix(note): publish note.created and note.deleted integration events"
```

---

### Task 4: Note events for notes removed with folders

**Files:**
- Modify: `internal/note/app/cmd_permanentlydeleteworkspaceitems.go` (`Handle`)
- Modify: `internal/note/app/cmd_emptytrash.go` (`Handle`)

**Interfaces:**
- Consumes: `domain.NoteRepo.GetRecursiveChildrenFromFolder(ctx, folderID uuid.UUID, forUpdate bool) ([]*domain.Note, error)`

- [ ] **Step 1: `PermanentlyDeleteWorkspaceItems` — delete notes first, including folder descendants**

Replace the body passed to `h.uow.Execute` with:

```go
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		folderRepo := r.Folder()
		noteRepo := r.Note()

		notesByID := map[uuid.UUID]*domain.Note{}
		if len(cmd.NoteIDs) > 0 {
			notes, err := noteRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.NoteRepoGetManyParams{
					IDs:       cmd.NoteIDs,
					ForUpdate: true,
				})
			if err != nil {
				return err
			}
			for _, n := range notes {
				notesByID[n.ID()] = n
			}
		}

		var folders []*domain.Folder
		if len(cmd.FolderIDs) > 0 {
			var err error
			folders, err = folderRepo.GetMany(ctx,
				//exhaustruct:ignore
				&domain.FolderRepoGetManyParams{
					IDs:       cmd.FolderIDs,
					ForUpdate: true,
				})
			if err != nil {
				return err
			}
			for _, folder := range folders {
				children, err := noteRepo.GetRecursiveChildrenFromFolder(ctx, folder.ID(), true)
				if err != nil {
					return err
				}
				for _, n := range children {
					notesByID[n.ID()] = n
				}
			}
		}

		notes := make([]*domain.Note, 0, len(notesByID))
		for _, n := range notesByID {
			n.PermanentlyDelete(cmd.UserID)
			notes = append(notes, n)
		}
		if err := noteRepo.SaveMany(ctx, notes); err != nil {
			return err
		}

		for _, folder := range folders {
			folder.PermanentlyDelete(cmd.UserID)
		}
		return folderRepo.SaveMany(ctx, folders)
	})
```

Remove the old `// NOTE: We delegate the infra persistence to cascading delete things (folder)` comment pair (no longer true for notes). Add `"github.com/google/uuid"` import if missing.

- [ ] **Step 2: `EmptyTrash` — notes before folders**

Move the `trashedNotes` block (GetMany → `PermanentlyDelete` loop → `SaveMany`) above the `trashedFolders` block. `TrashedBy: domain.TrashedByUnspecified` already includes notes trashed by a parent folder, so every note gets its event before the folder cascade.

- [ ] **Step 3: Build, lint**

```bash
nx build note && nx lint note
```

- [ ] **Step 4: Manual check (compose up, note running)**

Create folder A → sub-folder B → note N inside B; trash A; permanently delete A from the trash UI. In the Redpanda console, topic `events.integration.note.note.deleted` has a message with `id = N`.

- [ ] **Step 5: Commit**

```bash
git add internal/note/app
git commit -m "fix(note): emit note deleted events for notes removed with folders"
```

---

### Task 5: `workspace.deleted` integration event

**Files:**
- Create: `api/share/components/schemas/WorkspaceDeletedIntegrationEvent.yaml`
- Modify: `api/share/openapi.yaml`
- Create: `internal/note/app/e_workspacedeletedintegration.go`
- Modify: `internal/note/app/integrationevent.go`, `internal/note/component/eventtopic.go`, `internal/note/infra/integrationpublisher/integrationpublisher.go`, `internal/note/app/app.go`, `internal/note/app/wire.go`, `internal/note/controller/event/event.go`
- Regenerate: `pkg/api/share/share.gen.go`, `packages/api-share-gen`, `cmd/note/wire_gen.go`

**Interfaces:**
- Produces: topic `events.integration.note.workspace.deleted`, payload `share.WorkspaceDeletedIntegrationEvent{Id uuid.UUID}` (TS `WorkspaceDeletedIntegrationEvent { id: string }`). Consumed by billing (plan 03).

- [ ] **Step 1: Schema**

```yaml
# api/share/components/schemas/WorkspaceDeletedIntegrationEvent.yaml
# vim: ft=yaml.openapi
type: object

properties:
  id:
    type: string
    format: uuid

required:
  - id
```

`api/share/openapi.yaml` `components.schemas`, alphabetical:

```yaml
    WorkspaceDeletedIntegrationEvent:
      $ref: ./components/schemas/WorkspaceDeletedIntegrationEvent.yaml
```

Run: `nx run api:gen` → `pkg/api/share/share.gen.go` contains `type WorkspaceDeletedIntegrationEvent struct`. Run `nx lint api`.

- [ ] **Step 2: Integration event + topic + payload mapping**

`app/integrationevent.go`:

```go
type IntegrationEventWorkspaceDeleted struct {
	ID uuid.UUID
}
```
plus `_ IntegrationEvent = (*IntegrationEventWorkspaceDeleted)(nil)` and
`func (e IntegrationEventWorkspaceDeleted) isIntegrationEvent() {}`.

`component/eventtopic.go` `IntegrationEventToTopic`:

```go
	case app.IntegrationEventWorkspaceDeleted:
		return IntegrationEventTopicPrefix + "workspace.deleted", nil
```

`integrationpublisher.go` `transformIntegrationEvent`:

```go
	case app.IntegrationEventWorkspaceDeleted:
		return &share.WorkspaceDeletedIntegrationEvent{Id: e.ID}, nil
```

- [ ] **Step 3: Handler + wiring**

```go
// internal/note/app/e_workspacedeletedintegration.go
package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/note/domain"
)

type WorkspaceDeletedDomainToIntegrationEventHandler struct {
	integrationPublisher IntegrationPublisher
}

func NewWorkspaceDeletedDomainToIntegrationEventHandler(
	integrationPublisher IntegrationPublisher,
) *WorkspaceDeletedDomainToIntegrationEventHandler {
	return &WorkspaceDeletedDomainToIntegrationEventHandler{integrationPublisher: integrationPublisher}
}

var ProvideWorkspaceDeletedDomainToIntegrationEventHandler = NewWorkspaceDeletedDomainToIntegrationEventHandler

func (h *WorkspaceDeletedDomainToIntegrationEventHandler) Handle(ctx context.Context, event *domain.WorkspaceDeletedEvent) error {
	if err := h.integrationPublisher.Publish(ctx, IntegrationEventWorkspaceDeleted{ID: event.AggregateID}); err != nil {
		return fmt.Errorf("failed to publish workspace deleted integration event: %w", err)
	}
	slog.InfoContext(ctx, "Workspace deleted integration event published", slog.String("workspace_id", event.AggregateID.String()))
	return nil
}
```

Add `WorkspaceDeletedDomainToIntegrationEvent *WorkspaceDeletedDomainToIntegrationEventHandler` to `Events` in `app/app.go`, its provider to `app/wire.go`, and in `event.go` `setup()`:

```go
	if _, err := e.eventProcessor.AddHandler(cqrs.NewEventHandler(
		"WorkspaceDeletedToIntegrationHandler",
		e.app.Events.WorkspaceDeletedDomainToIntegrationEvent.Handle,
	)); err != nil {
		return fmt.Errorf("failed to add event handler WorkspaceDeletedToIntegrationHandler: %w", err)
	}
```

- [ ] **Step 4: Regenerate, build, lint**

```bash
nx run note:gen:wire && nx build note && nx lint note
```

- [ ] **Step 5: Commit**

```bash
git add api/share internal/note cmd/note/wire_gen.go pkg/api/share packages/api-share-gen
git commit -m "feat(note): publish workspace.deleted integration event"
```

---

### Task 6: search-worker consumes `note.deleted`

**Files:**
- Modify: `apps/search-worker/src/app.controller.ts`
- Modify: `apps/search-worker/src/app.service.ts`

**Interfaces:**
- Consumes: `NoteDeletedEvent` from `@notopia-uit/api-share-gen`.
- Produces: `AppService.handleNoteDeleted(id: string): Promise<void>`

- [ ] **Step 1: Implement**

`app.service.ts`:

```ts
  async handleNoteDeleted(id: string) {
    const index = this.meili.index(AppService.noteIndex);
    try {
      await index.deleteDocument(id);
      this.logger.log({ id }, 'handleNoteDeleted: removed');
    } catch (e) {
      if (e instanceof MeilisearchError) {
        throw new MeiliError(e);
      }
      throw e;
    }
  }
```

`app.controller.ts`:

```ts
  @EventPattern('events.integration.note.note.deleted')
  async handleNoteDeleted(@Payload() data: NoteDeletedEvent) {
    this.logger.log({ id: data.id }, 'handleNoteDeleted: received');
    await this.appService.handleNoteDeleted(data.id);
  }
```

(add `NoteDeletedEvent` to the type import).

- [ ] **Step 2: Lint, format, build, commit**

```bash
oxfmt apps/search-worker/src
nx lint search-worker --fix && nx build search-worker
git add apps/search-worker/src
git commit -m "feat(search-worker): remove deleted notes from index"
```

---

### Task 7: Docs + end-to-end verification

**Files:**
- Modify: `apps/document/AGENTS.md`, `apps/search-worker/AGENTS.md`, `internal/note/AGENTS.md`

- [ ] **Step 1: Update docs**

- `apps/document/AGENTS.md` "Yjs / Hocuspocus": store upserts the row and writes `modified` from the Yjs metadata; `afterLoadDocument` syncs the column into metadata with `skipStoreHooks`; `note.deleted` closes connections and deletes the row (store guarded by `deletedDocumentIds`). Remove the "sets `modified = true`" statement.
- `apps/search-worker/AGENTS.md` topic table: add `events.integration.note.note.deleted` → delete from `notes`.
- `internal/note/AGENTS.md` "Events": list published integration topics `note.created`, `note.updated`, `note.deleted`, `workspace.deleted`; note that folder deletes emit `NoteDeletedEvent` per descendant note.

- [ ] **Step 2: E2E (compose up; note, document, search-worker running)**

1. Create a note in web, type text → `SELECT id, modified FROM documents WHERE id = '<id>'` returns a row, `modified = true`.
2. Commit → wait 10 s → `modified = false` still.
3. Restart document service, reopen the note → UI shows the correct modified state.
4. Permanently delete the note while it is open in another tab → tab disconnects; `documents` and `revisions` rows gone; Meilisearch `notes` index no longer has the id.
5. Delete a workspace → `events.integration.note.workspace.deleted` message in the Redpanda console.

- [ ] **Step 3: Commit**

```bash
git add apps/document/AGENTS.md apps/search-worker/AGENTS.md internal/note/AGENTS.md
git commit -m "docs: document store semantics and deletion events"
```
