import type { LocalTransactionOrigin } from '@hocuspocus/server';
import { YDocMetadataMap } from '@notopia-uit/lib/yjs';
import * as Y from 'yjs';

const metadataOf = (doc: Y.Doc) => doc.getMap('metadata') as YDocMetadataMap;

export const readModified = (doc: Y.Doc): boolean =>
  metadataOf(doc).get('metadata')?.modified ?? false;

export const applyStoredModified = (doc: Y.Doc, modified: boolean): void => {
  const origin: LocalTransactionOrigin = { source: 'local', skipStoreHooks: true };
  doc.transact(() => metadataOf(doc).set('metadata', { modified }), origin);
};

const sameStateIgnoringClient = (a: Uint8Array, b: Uint8Array, ignoredClientId: number) => {
  const left = Y.decodeStateVector(a);
  const right = Y.decodeStateVector(b);
  left.delete(ignoredClientId);
  right.delete(ignoredClientId);
  if (left.size !== right.size) return false;
  for (const [client, clock] of left) {
    if (right.get(client) !== clock) return false;
  }
  return true;
};

/**
 * Clears the in-memory `modified` flag after a commit, but only when the loaded doc holds
 * exactly the state that was committed.
 *
 * Why: a commit reads `documents.data`, i.e. the last *stored* state. Edits can already be in
 * memory but not stored yet (inside the Hocuspocus debounce window, or made while the commit
 * transaction ran). Clearing the flag then would be wrong: `onChange` only sets `modified` when
 * it is false, so it would not set it again, the next store would write the newer content with
 * `modified=false`, and auto-commit / the sweeper would never commit it.
 *
 * The comparison skips this server doc's own client id: the server only writes metadata
 * (`applyStoredModified` on load, this function), never content, and those writes may not be
 * stored yet. Comparing them byte for byte would keep the flag after every load-then-commit.
 *
 * Returns true when the flag was cleared.
 */
export const clearModifiedIfCommitted = (doc: Y.Doc, committedStateVector: Uint8Array): boolean => {
  if (!sameStateIgnoringClient(Y.encodeStateVector(doc), committedStateVector, doc.clientID)) {
    return false;
  }
  metadataOf(doc).set('metadata', { modified: false });
  return true;
};
