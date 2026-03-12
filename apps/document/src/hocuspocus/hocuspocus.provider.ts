import { Database } from '@hocuspocus/extension-database';
import { Hocuspocus } from '@hocuspocus/server';
import { Provider } from '@nestjs/common';

import { DocumentRepository } from '../document/document.repository';

export const HocuspocusProvider: Provider = {
  provide: Hocuspocus,
  useFactory: (documentRepository: DocumentRepository) => {
    return new Hocuspocus({
      name: 'document', // # TODO: Inject host
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
