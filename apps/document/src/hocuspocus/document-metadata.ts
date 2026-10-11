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
