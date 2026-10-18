import * as Y from 'yjs';

import { applyStoredModified, clearModifiedIfCommitted, readModified } from './document-metadata';

describe('document metadata', () => {
  it('applies the stored flag with a store-skipping origin', () => {
    const doc = new Y.Doc();
    const origins: unknown[] = [];
    doc.on('afterTransaction', (tr) => origins.push(tr.origin));

    applyStoredModified(doc, true);

    expect(readModified(doc)).toBe(true);
    expect(origins).toEqual([{ source: 'local', skipStoreHooks: true }]);
  });

  it('reads false when metadata is missing', () => {
    expect(readModified(new Y.Doc())).toBe(false);
  });
});

describe('clearModifiedIfCommitted', () => {
  const loadedFromStore = () => {
    const client = new Y.Doc();
    client.getText('content').insert(0, 'hello');
    const stored = Y.encodeStateAsUpdate(client);
    const server = new Y.Doc();
    Y.applyUpdate(server, stored);
    applyStoredModified(server, true);
    return { client, server, committedStateVector: Y.encodeStateVectorFromUpdate(stored) };
  };

  it('clears the flag when the loaded doc holds exactly the committed state', () => {
    const { server, committedStateVector } = loadedFromStore();

    expect(clearModifiedIfCommitted(server, committedStateVector)).toBe(true);
    expect(readModified(server)).toBe(false);
  });

  it('keeps the flag when a newer edit is not stored yet', () => {
    const { client, server, committedStateVector } = loadedFromStore();
    const before = Y.encodeStateVector(client);
    client.getText('content').insert(5, ' world');
    Y.applyUpdate(server, Y.encodeStateAsUpdate(client, before));

    expect(clearModifiedIfCommitted(server, committedStateVector)).toBe(false);
    expect(readModified(server)).toBe(true);
  });
});
