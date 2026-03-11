import { Database } from '@hocuspocus/extension-database';
import { Server } from '@hocuspocus/server';
import { Provider } from '@nestjs/common';

import { DocumentRepository } from '../document/document.repository';

export const HocuspocusProvider: Provider = {
  provide: Server,
  useFactory: (documentRepository: DocumentRepository) => {
    return new Server({
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
    });
  },
  inject: [DocumentRepository],
};
