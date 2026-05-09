import { Database } from '@hocuspocus/extension-database';
import { Logger as HocuspocusLogger } from '@hocuspocus/extension-logger';
import { Hocuspocus } from '@hocuspocus/server';
import { Logger, Provider } from '@nestjs/common';

import { AuthorizationService } from '#/authorization/authorization.service';
import { DocumentService } from '#/document/document.service';
import { NoteService } from '#/note/note.service';

import { HocuspocusContext } from './hocuspocus-context';

export const HocuspocusProvider: Provider = {
  provide: Hocuspocus<HocuspocusContext>,
  useFactory: (
    documentService: DocumentService,
    noteService: NoteService,
    authorizationService: AuthorizationService,
    logger: Logger
  ) => {
    return new Hocuspocus<HocuspocusContext>({
      name: 'document', // TODO: Inject host
      extensions: [
        new HocuspocusLogger({
          log: (message) => logger.log(message),
        }),
        new Database({
          fetch: async ({ documentName: id, document: hocuspocusDocument }) => {
            const document = await documentService.getById(id);
            hocuspocusDocument.awareness.setLocalStateField('isModified', document?.modified);
            return document?.data ?? null;
          },
          store: async ({ documentName: id, state }) => {
            await documentService.updateDataById(id, state);
          },
        }),
      ],

      async onAuthenticate(data) {
        const documentId = data.documentName;
        const context = data.context;
        const note = await noteService.getNoteById({
          noteId: documentId,
          userId: context.user.id,
        });
        if (!note) {
          throw new Error(`Document with ID ${documentId} does not exist`);
        }
        const userPermissionsRes = await authorizationService.getUserDocumentPermissions(
          context.user.id,
          documentId
        );
        if (!userPermissionsRes.canRead) {
          throw new Error(
            `User ${context.user.id} does not have permission to access document ${documentId}`
          );
        }
        if (!userPermissionsRes.canWrite || note.trashed) {
          data.connectionConfig.readOnly = true;
        }
      },

      async onChange(data) {
        data.document.getMap('isModified').set('value', true);
        return Promise.resolve();
      },
    });
  },
  inject: [DocumentService, NoteService, AuthorizationService, Logger],
};
