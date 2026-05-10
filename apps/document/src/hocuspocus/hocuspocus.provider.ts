import { Database } from '@hocuspocus/extension-database';
import { Logger as HocuspocusLogger } from '@hocuspocus/extension-logger';
import { Hocuspocus } from '@hocuspocus/server';
import { Logger, Provider } from '@nestjs/common';

import { AuthenticationService } from '#/authentication/authentication.service';
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
    logger: Logger,
    authenticationService: AuthenticationService
  ) => {
    return new Hocuspocus<HocuspocusContext>({
      name: 'document', // TODO: Inject host
      extensions: [
        new HocuspocusLogger({
          log: (message) => logger.log(message),
        }),
        new Database({
          fetch: async ({ documentName: id }) => {
            const document = await documentService.getById(id);
            return document?.data ?? null;
          },
          store: async ({ documentName: id, state }) => {
            await documentService.updateDataById(id, state);
          },
        }),
      ],

      async onAuthenticate(data): Promise<HocuspocusContext> {
        const documentId = data.documentName;
        const user = await authenticationService.validateToken(data.token);
        const note = await noteService.getNoteById({
          noteId: documentId,
          userId: user.id,
        });
        if (!note) {
          throw new Error(`Document with ID ${documentId} does not exist`);
        }
        const userPermissionsRes = await authorizationService.getUserDocumentPermissions(
          user.id,
          documentId
        );
        if (!userPermissionsRes.canRead) {
          throw new Error(
            `User ${user.id} does not have permission to access document ${documentId}`
          );
        }
        if (!userPermissionsRes.canWrite || note.trashed) {
          data.connectionConfig.readOnly = true;
        }
        return { user };
      },

      async onChange(data) {
        data.document.getMap('isModified').set('value', true);
        return Promise.resolve();
      },
    });
  },
  inject: [DocumentService, NoteService, AuthorizationService, Logger, AuthenticationService],
};
