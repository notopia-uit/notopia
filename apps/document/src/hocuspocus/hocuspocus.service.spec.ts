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

  it('blocks stores before closing connections, then deletes the row', async () => {
    const { service, hocuspocus, documentService } = ctx;
    const calls: string[] = [];
    (hocuspocus.markDeleted as unknown as Mock).mockImplementation(() => void calls.push('mark'));
    (hocuspocus.closeConnections as unknown as Mock).mockImplementation(
      () => void calls.push('close')
    );
    (documentService.deleteById as unknown as Mock).mockImplementation(
      async () => void calls.push('delete')
    );

    await service.onNoteDeleted('n1');

    expect(calls).toEqual(['mark', 'close', 'delete']);
  });
});
