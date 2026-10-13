import { Database as HocuspocusDatabase } from '@hocuspocus/extension-database';
import { Logger as HocuspocusLogger } from '@hocuspocus/extension-logger';
import {
  onAuthenticatePayload,
  onChangePayload,
  Hocuspocus as ServerHocuspocus,
} from '@hocuspocus/server';
import { Injectable, Logger } from '@nestjs/common';
import { YDocMetadataMap } from '@notopia-uit/lib/yjs';

import { AuthenticationService } from '../authentication/authentication.service';
import { AuthorizationService } from '../authorization/authorization.service';
import { DocumentService } from '../document/document.service';
import { NoteService } from '../note/note.service';
import { applyStoredModified, readModified } from './document-metadata';
import { HocuspocusContext } from './hocuspocus-context';

@Injectable()
export class Hocuspocus {
  private readonly logger = new Logger(Hocuspocus.name);
  readonly hocuspocus: ServerHocuspocus<HocuspocusContext>;
  private readonly deletedDocumentIds = new Set<string>();
  private readonly pendingStores = new Map<string, Set<Promise<void>>>();

  constructor(
    documentService: DocumentService,
    private readonly noteService: NoteService,
    private readonly authorizationService: AuthorizationService,
    private readonly authenticationService: AuthenticationService
  ) {
    this.hocuspocus = new ServerHocuspocus<HocuspocusContext>({
      name: 'document', // TODO: Inject host
      onLoadDocument: async (data) => {
        this.logger.debug(
          { documentId: data.documentName, documentMetadata: data.document.getMap('metadata') },
          'Document loaded'
        );
        await Promise.resolve();
      },
      afterLoadDocument: async ({ documentName, document }) => {
        const entity = await documentService.getById(documentName);
        applyStoredModified(document, entity?.modified ?? false);
      },
      afterUnloadDocument: async ({ documentName }) => {
        this.deletedDocumentIds.delete(documentName);
      },
      onAuthenticate: (...args) => this.onAuthenticate(...args),
      onChange: (...args) => this.onChange(...args),
      extensions: [
        new HocuspocusLogger({
          log: (message) => this.logger.log(message),
        }),
        new HocuspocusDatabase({
          fetch: async ({ documentName: id }) => {
            this.logger.log({ documentId: id }, 'Fetching document from database');
            const document = await documentService.getById(id);
            this.logger.log({ documentId: id }, 'Fetched document from database');
            return document?.data ?? null;
          },
          store: async ({ documentName: id, state, document }) => {
            if (this.deletedDocumentIds.has(id)) return;
            await this.trackStore(
              id,
              documentService.upsertData(id, state, readModified(document))
            );
          },
        }),
      ],
    });
  }

  markDeleted(id: string): void {
    if (this.hocuspocus.documents.has(id)) {
      this.deletedDocumentIds.add(id);
    }
  }

  async waitForPendingStores(id: string): Promise<void> {
    const pending = this.pendingStores.get(id);
    if (pending) await Promise.allSettled(pending);
  }

  private async trackStore(id: string, store: Promise<void>): Promise<void> {
    let pending = this.pendingStores.get(id);
    if (!pending) {
      pending = new Set();
      this.pendingStores.set(id, pending);
    }
    pending.add(store);
    try {
      await store;
    } finally {
      pending.delete(store);
      if (pending.size === 0) this.pendingStores.delete(id);
    }
  }

  closeConnections(id: string): void {
    this.hocuspocus.closeConnections(id);
  }

  private async onAuthenticate(
    data: onAuthenticatePayload<HocuspocusContext>
  ): Promise<HocuspocusContext> {
    const documentId = data.documentName;
    this.logger.log({ documentId }, 'Authenticating user for document');
    const user = await this.authenticationService.validateToken(data.token);
    const note = await this.noteService.getNoteById({
      noteId: documentId,
      userId: user.id,
    });
    if (!note) {
      throw new Error(`Document with ID ${documentId} does not exist`);
    }
    const userPermissionsRes = await this.authorizationService.getUserDocumentPermissions(
      user.id,
      documentId
    );
    if (!userPermissionsRes.canRead) {
      throw new Error(`User ${user.id} does not have permission to access document ${documentId}`);
    }
    if (!userPermissionsRes.canWrite || note.trashed) {
      data.connectionConfig.readOnly = true;
    }
    return { user };
  }

  private async onChange(
    data: onChangePayload<HocuspocusContext>
  ): Promise<HocuspocusContext | void> {
    if (!data.connection) return Promise.resolve();

    const metadata = data.document.getMap('metadata') as YDocMetadataMap;
    const existing = metadata.get('metadata');
    if (!existing || !existing.modified) {
      metadata.set('metadata', { modified: true });
    }
  }
}
