import { TestBed } from '@suites/unit';
import { type Mock } from 'vitest';

import { DocumentService } from '../document/document.service';
import { Hocuspocus } from './hocuspocus';
import { HocuspocusService } from './hocuspocus.service';

async function setup() {
  const { unit, unitRef } = await TestBed.solitary(HocuspocusService).compile();
  return {
    service: unit,
    hocuspocus: unitRef.get(Hocuspocus),
    documentService: unitRef.get(DocumentService),
  };
}

describe('HocuspocusService.onNoteDeleted', () => {
  let ctx: Awaited<ReturnType<typeof setup>>;

  beforeEach(async () => {
    ctx = await setup();
  });

  it('blocks stores, closes connections, drains in-flight stores, then deletes the row', async () => {
    const { service, hocuspocus, documentService } = ctx;
    const calls: string[] = [];
    (hocuspocus.markDeleted as unknown as Mock).mockImplementation(() => void calls.push('mark'));
    (hocuspocus.closeConnections as unknown as Mock).mockImplementation(
      () => void calls.push('close')
    );
    (hocuspocus.waitForPendingStores as unknown as Mock).mockImplementation(
      async () => void calls.push('drain')
    );
    (documentService.deleteById as unknown as Mock).mockImplementation(
      async () => void calls.push('delete')
    );

    await service.onNoteDeleted('n1');

    expect(calls).toEqual(['mark', 'close', 'drain', 'delete']);
  });
});

describe('HocuspocusService.flushStore', () => {
  function setupFlush(debounced: boolean, executing: boolean[] = []) {
    const unlock = Promise.withResolvers<void>();
    const saveMutex = { waitForUnlock: vi.fn<() => Promise<void>>(() => unlock.promise) };
    const isCurrentlyExecuting = vi.fn<(id: string) => boolean>().mockReturnValue(false);
    executing.forEach((value) => isCurrentlyExecuting.mockReturnValueOnce(value));
    const debouncer = {
      isDebounced: vi
        .fn<(id: string) => boolean>()
        .mockReturnValueOnce(debounced)
        .mockReturnValue(false),
      isCurrentlyExecuting,
      executeNow: vi.fn<(id: string) => Promise<void>>().mockResolvedValue(undefined),
    };
    const hocuspocus = {
      hocuspocus: { documents: new Map([['d1', { saveMutex }]]), debouncer },
    } as unknown as Hocuspocus;
    const service = new HocuspocusService({} as never, hocuspocus, {} as never, {} as never);
    return { service, debouncer, saveMutex, unlock };
  }

  it('runs the debounced store now', async () => {
    const { service, debouncer } = setupFlush(true);
    await service.flushStore('d1');
    expect(debouncer.executeNow).toHaveBeenCalledWith('onStoreDocument-d1');
  });

  it('waits for a store already running', async () => {
    const { service, debouncer, unlock } = setupFlush(false);
    let flushed = false;
    const flush = service.flushStore('d1').then(() => (flushed = true));
    await Promise.resolve();
    expect(flushed).toBe(false);
    unlock.resolve();
    await flush;
    expect(flushed).toBe(true);
    expect(debouncer.executeNow).not.toHaveBeenCalled();
  });

  it('waits again for a queued store that started after the running one finished', async () => {
    const { service, saveMutex, unlock } = setupFlush(false, [true, false]);
    unlock.resolve();
    await service.flushStore('d1');
    expect(saveMutex.waitForUnlock).toHaveBeenCalledTimes(2);
  });

  it('does nothing for a document not loaded here', async () => {
    const { service, debouncer } = setupFlush(true);
    await service.flushStore('other');
    expect(debouncer.executeNow).not.toHaveBeenCalled();
  });
});
