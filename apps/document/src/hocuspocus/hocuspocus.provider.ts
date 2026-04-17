import { Database } from '@hocuspocus/extension-database';
import { Logger as HocuspocusLogger } from '@hocuspocus/extension-logger';
import { Hocuspocus } from '@hocuspocus/server';
import { Logger, Provider } from '@nestjs/common';

import { AuthorizationService } from '#/authorization/authorization.service';
import { DocumentRepository } from '#/document/document.repository';
import { NoteService } from '#/note/note.service';

import { HocuspocusContext } from './hocuspocus-context';

export const HocuspocusProvider: Provider = {
  provide: Hocuspocus,
  useFactory: (
    documentRepository: DocumentRepository,
    noteService: NoteService,
    authorizationService: AuthorizationService,
    logger: Logger
  ) => {
    return new Hocuspocus({
      name: 'document', // TODO: Inject host
      extensions: [
        new HocuspocusLogger({
          log: (message) => logger.log(message),
        }),
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
        const note = await noteService.getNoteById({
          noteId: documentId,
          userId: context.user.id,
        });
        if (!note) {
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
        if (!userPermissionsRes.canWrite || note.trashed) {
          data.connectionConfig.readOnly = true;
        }
      },
    });
  },
  inject: [DocumentRepository, NoteService, AuthorizationService, Logger],
};
