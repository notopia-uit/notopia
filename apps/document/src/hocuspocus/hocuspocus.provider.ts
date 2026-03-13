import { Client } from '@connectrpc/connect';
import { Database } from '@hocuspocus/extension-database';
import { Hocuspocus } from '@hocuspocus/server';
import { Provider } from '@nestjs/common';
import {
  AuthorizationService,
  NotePermission,
} from '@notopia-uit/pb/authorization';
import { NoteService } from '@notopia-uit/pb/note';

import { AUTHORIZATION_SERVICE } from '../authorization/authorization.module';
import { DocumentRepository } from '../document/document.repository';
import { NOTE_SERVICE } from '../note/note.module';
import { HocuspocusContext } from './hocuspocus-context';

export const HocuspocusProvider: Provider = {
  provide: Hocuspocus,
  useFactory: (
    documentRepository: DocumentRepository,
    noteService: Client<typeof NoteService>,
    authorizationService: Client<typeof AuthorizationService>
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
        const response = await authorizationService.hasNotePermission({
          memberId: context.user.id,
          noteId: documentId,
          permission: NotePermission.READ,
        });
        if (!response.hasPermission) {
          throw new Error('Unauthorized');
        }
      },
    });
  },
  inject: [DocumentRepository, NOTE_SERVICE, AUTHORIZATION_SERVICE],
};
