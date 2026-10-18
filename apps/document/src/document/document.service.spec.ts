import type { MySchema } from '@blocknote/core';
import type { ConfigService } from '@nestjs/config';
import type { ClientKafka } from '@nestjs/microservices';
import { type Observable, of } from 'rxjs';
import type { DataSource, EntityManager, Repository } from 'typeorm';
import { vi } from 'vitest';
import * as Y from 'yjs';

import type { AuthorizationService } from '../authorization/authorization.service';
import { RevisionEntity } from '../revision/revision.entity';
import type { StorageService } from '../storage/storage.service';
import { contentHash } from './commit-policy';
import { DocumentEntity } from './document.entity';
import { DocumentService } from './document.service';

const BLOCKS = [{ type: 'paragraph' }];

vi.mock('../blocknote/blocknote-editor.service', () => ({
  BlocknoteEditorService: class {
    blocks() {
      return BLOCKS;
    }
    extractTagsAndOutgoingLinkIds() {
      return { tags: [], outgoingLinkIds: [] };
    }
  },
}));

const storedData = () => {
  const doc = new Y.Doc();
  doc.getText('content').insert(0, 'hello');
  return Buffer.from(Y.encodeStateAsUpdate(doc));
};

function setup({
  document,
  latest,
}: {
  document: Partial<DocumentEntity>;
  latest: Partial<RevisionEntity> | null;
}) {
  const manager = {
    findOne: vi.fn<(entity: unknown) => Promise<unknown>>(async (entity) =>
      entity === DocumentEntity ? document : latest
    ),
    save: vi.fn<(...args: unknown[]) => Promise<void>>(async () => undefined),
    update: vi.fn<(...args: unknown[]) => Promise<void>>(async () => undefined),
  };
  const dataSource = {
    transaction: (fn: (m: EntityManager) => Promise<unknown>) =>
      fn(manager as unknown as EntityManager),
  };
  const kafkaClient = {
    emit: vi.fn<(...args: unknown[]) => Observable<undefined>>(() => of(undefined)),
  };
  const configService = { get: () => ({ serviceAccountUserId: 'svc' }) };
  const service = new DocumentService(
    {} as Repository<DocumentEntity>,
    {} as StorageService,
    dataSource as unknown as DataSource,
    {} as AuthorizationService,
    kafkaClient as unknown as ClientKafka,
    {} as MySchema,
    configService as unknown as ConfigService
  );
  return { service, manager, kafkaClient };
}

describe('DocumentService.commitIfDue', () => {
  const baseDocument = (overrides: Partial<DocumentEntity> = {}): Partial<DocumentEntity> => ({
    id: 'd1',
    data: storedData(),
    modified: true,
    lastCommittedAt: null,
    pendingContributorIds: ['a', 'b'],
    ...overrides,
  });

  it('skips when not due', async () => {
    const { service, manager } = setup({
      document: baseDocument({ lastCommittedAt: new Date() }),
      latest: null,
    });

    const result = await service.commitIfDue({
      documentId: 'd1',
      trigger: 'auto',
      force: false,
      minIntervalSeconds: 300,
    });

    expect(result).toEqual({
      revisionId: null,
      modifiedCleared: false,
      committedStateVector: null,
    });
    expect(manager.update).not.toHaveBeenCalled();
  });

  it('only clears modified when the content equals the latest revision', async () => {
    const { service, manager, kafkaClient } = setup({
      document: baseDocument(),
      latest: { id: 'r0', contentHash: contentHash(BLOCKS) },
    });

    const result = await service.commitIfDue({
      documentId: 'd1',
      trigger: 'sweeper',
      force: true,
      minIntervalSeconds: 0,
    });

    expect(result.revisionId).toBeNull();
    expect(result.modifiedCleared).toBe(true);
    expect(result.committedStateVector).toBeInstanceOf(Uint8Array);
    expect(manager.save).not.toHaveBeenCalled();
    expect(manager.update).toHaveBeenCalledWith(DocumentEntity, { id: 'd1' }, { modified: false });
    expect(kafkaClient.emit).not.toHaveBeenCalled();
  });

  it('saves a revision attributed to the latest contributor and emits the event', async () => {
    const { service, manager, kafkaClient } = setup({
      document: baseDocument(),
      latest: { id: 'r0', contentHash: 'older' },
    });

    const result = await service.commitIfDue({
      documentId: 'd1',
      trigger: 'auto',
      force: false,
      minIntervalSeconds: 300,
    });

    expect(result.revisionId).toEqual(expect.any(String));
    expect(result.modifiedCleared).toBe(true);
    expect(manager.save).toHaveBeenCalledWith(
      RevisionEntity,
      expect.objectContaining({
        id: result.revisionId,
        authorId: 'b',
        trigger: 'auto',
        contributorIds: ['a', 'b'],
        contentHash: contentHash(BLOCKS),
      })
    );
    expect(manager.update).toHaveBeenCalledWith(
      DocumentEntity,
      { id: 'd1' },
      expect.objectContaining({ modified: false, pendingContributorIds: [] })
    );
    expect(kafkaClient.emit.mock.calls[0][1]).toMatchObject({ id: 'd1', userId: 'b' });
  });
});
