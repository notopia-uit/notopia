import * as Y from 'yjs';

import { applyStoredModified, readModified } from './document-metadata';

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
