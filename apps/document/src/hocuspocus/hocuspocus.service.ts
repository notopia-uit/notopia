import { Connection } from '@hocuspocus/server';
import { Injectable } from '@nestjs/common';
import { Traceable } from 'nestjs-otel';

import { AuthorizationService } from '../authorization/authorization.service';
import { DocumentService } from '../document/document.service';
import { WorkspaceModel } from '../note/models';
import { NoteService } from '../note/note.service';
import { WorkspaceNoteNotFoundException } from '../note/workspace-note-not-found.exception';
import { Hocuspocus } from './hocuspocus';
import { HocuspocusContext } from './hocuspocus-context';

// Debounce id used by @hocuspocus/server 4.0.0 storeDocumentHooks
const STORE_DEBOUNCE_ID_PREFIX = 'onStoreDocument-';
const MAX_FLUSH_PASSES = 5;

@Injectable()
@Traceable()
export class HocuspocusService {
  constructor(
    private readonly noteService: NoteService,
    private readonly hocuspocus: Hocuspocus,
    private readonly authorizationService: AuthorizationService,
    private readonly documentService: DocumentService
  ) {}

  async onNoteDeleted(id: string): Promise<void> {
    this.hocuspocus.markDeleted(id);
    this.hocuspocus.closeConnections(id);
    await this.hocuspocus.waitForPendingStores(id);
    await this.documentService.deleteById(id);
  }

  // NOTE: We don't really need to check the permission canRead, because yeah
  // Because based on the casbin rules, it will be always canRead when this event fired
  async onRoleChanged({ workspaceId, userId }: { workspaceId: string; userId: string }) {
    for (const [documentName, document] of this.hocuspocus.hocuspocus.documents) {
      for (const connection of document.getConnections() as Connection<HocuspocusContext>[]) {
        const context = connection.context as HocuspocusContext;
        if (context.user.id !== userId) {
          continue;
        }
        let workspace: WorkspaceModel | undefined;
        try {
          workspace = await this.noteService.getWorkspaceByNote({
            noteId: documentName,
            userId,
          });
        } catch (e) {
          if (e instanceof WorkspaceNoteNotFoundException) {
            connection.close({
              code: 4001,
              reason: `Your access to document ${documentName} has been revoked due to workspace be found for this document.`,
            });
            continue;
          }
        }
        if (workspace?.id !== workspaceId) {
          continue;
        }
        const permissions = await this.authorizationService.getWorkspaceItemPermissions({
          workspaceId,
          memberId: userId,
        });
        if (!permissions.canRead) {
          connection.close({
            code: 4002,
            reason:
              'Your access to this document has been revoked because you no longer have read permission.',
          });
        } else {
          connection.readOnly = !permissions.canWrite;
        }
      }
    }
  }

  // NOTE: If we handle the published, then this should be adjusted
  async onMemberRemoved({ workspaceId, userId }: { workspaceId: string; userId: string }) {
    for (const [documentName, document] of this.hocuspocus.hocuspocus.documents) {
      for (const connection of document.getConnections() as Connection<HocuspocusContext>[]) {
        const context = connection.context as HocuspocusContext;
        if (context.user.id !== userId) {
          continue;
        }
        let workspace: WorkspaceModel | undefined;
        try {
          workspace = await this.noteService.getWorkspaceByNote({
            noteId: documentName,
            userId,
          });
        } catch (e) {
          if (e instanceof WorkspaceNoteNotFoundException) {
            connection.close({
              code: 4001,
              reason: `Your access to document ${documentName} has been revoked due to workspace be found for this document.`,
            });
            continue;
          }
        }
        if (workspace?.id !== workspaceId) {
          continue;
        }
        connection.close({
          code: 4002,
          reason:
            'Your access to this document has been revoked because you have been removed from the workspace.',
        });
      }
    }
  }

  /**
   * Runs the store pending in the Hocuspocus debounce now (or waits for the one running), so a
   * caller reading `documents.data` sees the latest edits. A document not loaded here is a no-op.
   *
   * Since @hocuspocus/server 4.1.2 a due store removes its debounce timer before it waits for the
   * previous store to finish, so for a moment it is neither debounced nor the running one. We
   * therefore repeat until the debouncer is idle, yielding a macrotask so that queued store can
   * take the mutex. Bounded, so a document being typed into cannot hold the caller forever.
   */
  async flushStore(documentId: string): Promise<void> {
    const document = this.hocuspocus.hocuspocus.documents.get(documentId);
    if (!document) {
      return;
    }
    const { debouncer } = this.hocuspocus.hocuspocus;
    const debounceId = `${STORE_DEBOUNCE_ID_PREFIX}${documentId}`;
    for (let pass = 0; pass < MAX_FLUSH_PASSES; pass++) {
      if (debouncer.isDebounced(debounceId)) {
        // executeNow awaits a running execution before starting the debounced one
        await debouncer.executeNow(debounceId);
      } else {
        await document.saveMutex.waitForUnlock();
      }
      await new Promise<void>((resolve) => setImmediate(resolve));
      if (!debouncer.isDebounced(debounceId) && !debouncer.isCurrentlyExecuting(debounceId)) {
        return;
      }
    }
  }

  /**
   * Clears the loaded doc's `modified` flag after a commit unless it holds newer content
   * (see `DocumentService.clearModifiedIfCommitted`). A document not loaded here has nothing
   * to clear.
   */
  clearModifiedIfCommitted(documentId: string, committedContentHash: string): boolean {
    const document = this.hocuspocus.hocuspocus.documents.get(documentId);
    if (!document) {
      return false;
    }
    return this.documentService.clearModifiedIfCommitted(document, committedContentHash);
  }
}
