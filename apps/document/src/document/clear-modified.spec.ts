import { ServerBlockNoteEditor } from '@blocknote/server-util';
import type { ConfigService } from '@nestjs/config';
import type { ClientKafka } from '@nestjs/microservices';
import { createSchema } from '@notopia-uit/lib-server/block-note';
import type { DataSource, Repository } from 'typeorm';
import * as Y from 'yjs';

import type { AuthorizationService } from '../authorization/authorization.service';
import { BlocknoteEditorService } from '../blocknote/blocknote-editor.service';
import { applyStoredModified, readModified } from '../hocuspocus/document-metadata';
import type { StorageService } from '../storage/storage.service';
import { contentHash } from './commit-policy';
import type { DocumentEntity } from './document.entity';
import { DocumentService } from './document.service';

const schema = createSchema();

const service = new DocumentService(
  {} as Repository<DocumentEntity>,
  {} as StorageService,
  {} as DataSource,
  {} as AuthorizationService,
  {} as ClientKafka,
  schema,
  { get: () => ({ serviceAccountUserId: 'svc' }) } as unknown as ConfigService
);

const firstText = (doc: Y.Doc): Y.XmlText => {
  const walk = (node: Y.XmlElement | Y.XmlFragment): Y.XmlText | undefined => {
    for (const child of node.toArray()) {
      if (child instanceof Y.XmlText) return child;
      const found = walk(child as Y.XmlElement);
      if (found) return found;
    }
    return undefined;
  };
  const text = walk(doc.getXmlFragment('prosemirror'));
  if (!text) throw new Error('no text node');
  return text;
};

// Loaded doc = stored data + the server's own metadata write on load; the client keeps editing.
const loadedFromStore = () => {
  const client = ServerBlockNoteEditor.create({ schema }).blocksToYDoc([
    { type: 'paragraph', content: 'hello world' },
  ]);
  const stored = Buffer.from(Y.encodeStateAsUpdate(client));
  const server = new Y.Doc();
  Y.applyUpdate(server, stored);
  applyStoredModified(server, true);
  const committedContentHash = contentHash(
    new BlocknoteEditorService({ schema, initialContent: stored }).blocks()
  );
  const pushClientEdit = (edit: (text: Y.XmlText) => void) => {
    const before = Y.encodeStateVector(client);
    edit(firstText(client));
    Y.applyUpdate(server, Y.encodeStateAsUpdate(client, before));
  };
  return { server, committedContentHash, pushClientEdit };
};

describe('DocumentService.clearModifiedIfCommitted', () => {
  it('clears the flag when the loaded doc holds the committed content', () => {
    const { server, committedContentHash } = loadedFromStore();

    expect(service.clearModifiedIfCommitted(server, committedContentHash)).toBe(true);
    expect(readModified(server)).toBe(false);
  });

  it('keeps the flag when a newer insertion is not stored yet', () => {
    const { server, committedContentHash, pushClientEdit } = loadedFromStore();
    pushClientEdit((text) => text.insert(0, 'oh '));

    expect(service.clearModifiedIfCommitted(server, committedContentHash)).toBe(false);
    expect(readModified(server)).toBe(true);
  });

  it('keeps the flag when a newer deletion-only edit is not stored yet', () => {
    const { server, committedContentHash, pushClientEdit } = loadedFromStore();
    const vectorBefore = Y.encodeStateVector(server);
    pushClientEdit((text) => text.delete(0, 3));

    expect(Y.encodeStateVector(server)).toEqual(vectorBefore);
    expect(service.clearModifiedIfCommitted(server, committedContentHash)).toBe(false);
    expect(readModified(server)).toBe(true);
  });
});
