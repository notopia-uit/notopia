import { Database } from '@hocuspocus/extension-database';
import { Hocuspocus } from '@hocuspocus/server';
import { Provider } from '@nestjs/common';

import { AuthorizationService } from '../authorization/authorization.service';
import { DocumentRepository } from '../document/document.repository';
import { NoteService } from '../note/note.service';
import { HocuspocusContext } from './hocuspocus-context';

export const HocuspocusProvider: Provider = {
  provide: Hocuspocus,
  useFactory: (
    documentRepository: DocumentRepository,
    noteService: NoteService,
    authorizationService: AuthorizationService
  ) => {
    return new Hocuspocus({
      name: 'document', // TODO: Inject host
      extensions: [
        new Database({
          fetch: async ({ documentName: id }) => {
            const document = await documentRepository.getById(id);
            return document?.data ?? null;
          },
          store: async ({ documentName: id, state }) => {
            await documentRepository.updateDataById(id, state);
          },
        }),
      ],

      async onAuthenticate(data) {
        const documentId = data.documentName;
        const context = data.context as HocuspocusContext;
        const noteExistenceRes =
          await noteService.checkNoteExistence(documentId);
        if (!noteExistenceRes.exists) {
          throw new Error(`Document with ID ${documentId} does not exist`);
        }
        const userPermissionsRes =
          await authorizationService.getUserNotePermissions(
            context.user.id,
            documentId
          );
        if (!userPermissionsRes.canRead) {
          throw new Error(
            `User ${context.user.id} does not have permission to access document ${documentId}`
          );
        }
        if (!userPermissionsRes.canWrite) {
          data.connectionConfig.readOnly = true;
        }
      },

      async beforeHandleMessage(data) {
        const { context, connection } = data;

        const response = await authorizationService.getUserNotePermissions(
          context.userId,
          data.documentName
        );

        if (!response.canRead) {
          connection.close();
          throw new Error(
            `User ${context.userId} does not have permission to access document ${data.documentName}`
          );
        }
        if (response.canWrite) {
          connection.readOnly = false;
        } else {
          connection.readOnly = true;
        }
      },
    });
  },
  inject: [DocumentRepository, NoteService, AuthorizationService],
};
