import type { MySchema } from '@blocknote/core';
import type { Meilisearch } from 'meilisearch';

import { AppService } from './app.service';
import { MeiliTaskError } from './errors';
import { MemoryTombstoneStore } from './tombstone';

function fakeMeili(onUpdate?: (id: string) => Promise<void>) {
  const docs = new Set<string>();
  let status = 'succeeded';
  const task = (run: () => Promise<void> | void) => ({
    waitTask: async () => {
      await run();
      return { uid: 1, type: 'documentAdditionOrUpdate', status, error: null };
    },
  });
  const index = {
    updateDocuments: vi.fn((batch: { id: string }[]) =>
      task(async () => {
        for (const { id } of batch) {
          docs.add(id);
          await onUpdate?.(id);
        }
      })
    ),
    deleteDocument: vi.fn((id: string) => task(() => void docs.delete(id))),
  };
  return {
    meili: { index: () => index } as unknown as Meilisearch,
    index,
    docs,
    failTasks: () => (status = 'failed'),
  };
}

const update = (id: string) => ({ id, name: 'n', folderId: 'f', folderName: 'f' });

describe('AppService tombstones', () => {
  it('skips a late upsert after note.deleted', async () => {
    const { meili, index, docs } = fakeMeili();
    const service = new AppService(meili, new MemoryTombstoneStore(60, 10), {} as MySchema);

    await service.handleNoteDeleted('n1');
    await service.handleNoteUpdated(update('n1'));

    expect(index.updateDocuments).not.toHaveBeenCalled();
    expect(docs.has('n1')).toBe(false);
  });

  it('removes the note when note.deleted lands while the upsert is in flight', async () => {
    const tombstones = new MemoryTombstoneStore(60, 10);
    const { meili, docs } = fakeMeili((id) => tombstones.mark(id));
    const service = new AppService(meili, tombstones, {} as MySchema);

    await service.handleNoteUpdated(update('n1'));

    expect(docs.has('n1')).toBe(false);
  });

  it('throws MeiliTaskError when the Meilisearch task fails', async () => {
    const { meili, failTasks } = fakeMeili();
    failTasks();
    const service = new AppService(meili, new MemoryTombstoneStore(60, 10), {} as MySchema);

    await expect(service.handleNoteUpdated(update('n1'))).rejects.toBeInstanceOf(MeiliTaskError);
  });
});
