import { ConfigService } from '@nestjs/config';
import { Repository } from 'typeorm';

import { BillingService, BillingUnavailableException, PlanTier } from '../billing';
import { CRON_CONFIG } from '../config';
import { DocumentEntity } from '../document/document.entity';
import { DocumentService } from '../document/document.service';
import { HocuspocusService } from '../hocuspocus/hocuspocus.service';
import { NoteService } from '../note/note.service';
import { SweeperCron } from './sweeper.cron';

const ids = (prefix: string, n: number) =>
  Array.from({ length: n }, (_, i) => ({ id: `${prefix}${String(i).padStart(4, '0')}` }));

function setup(pages: { id: string }[][]) {
  const repo = { find: vi.fn<() => Promise<{ id: string }[]>>() };
  for (const page of pages) repo.find.mockResolvedValueOnce(page);
  repo.find.mockResolvedValue([]);
  const noteService = {
    batchGetNoteWorkspaces: vi.fn<(documentIds: string[]) => Promise<Map<string, string>>>(
      async (documentIds) => new Map(documentIds.map((id) => [id, 'w1']))
    ),
  };
  const billingService = {
    batchGetEntitlements: vi
      .fn<() => Promise<Map<string, unknown>>>()
      .mockResolvedValue(
        new Map([['w1', { plan: PlanTier.Plus, features: { auto_commit: true }, limits: {} }]])
      ),
  };
  const documentService = {
    commitIfDue: vi.fn<() => Promise<unknown>>().mockResolvedValue({
      revisionId: 'r',
      modifiedCleared: true,
      committedContentHash: 'h',
    }),
  };
  const hocuspocusService = { clearModifiedIfCommitted: vi.fn<() => boolean>() };
  const configService = {
    get: (token: symbol) => (token === CRON_CONFIG ? { sweeper: { idleSeconds: 600 } } : undefined),
  };
  const sweeper = new SweeperCron(
    repo as unknown as Repository<DocumentEntity>,
    documentService as unknown as DocumentService,
    hocuspocusService as unknown as HocuspocusService,
    noteService as unknown as NoteService,
    billingService as unknown as BillingService,
    configService as unknown as ConfigService
  );
  return { sweeper, repo, documentService, billingService, hocuspocusService };
}

describe('SweeperCron', () => {
  it('pages through every stale document and stops after a short page', async () => {
    const { sweeper, repo, documentService, hocuspocusService } = setup([
      ids('a', SweeperCron.batchSize),
      ids('b', 3),
    ]);

    await sweeper.sweep(new AbortController().signal);

    expect(repo.find).toHaveBeenCalledTimes(2);
    expect(documentService.commitIfDue).toHaveBeenCalledTimes(SweeperCron.batchSize + 3);
    expect(hocuspocusService.clearModifiedIfCommitted).toHaveBeenCalledWith('b0002', 'h');
  });

  it('skips the tick without committing when billing is unavailable', async () => {
    const { sweeper, documentService, billingService } = setup([ids('a', 2)]);
    billingService.batchGetEntitlements.mockRejectedValueOnce(
      new BillingUnavailableException(new Error('down'))
    );

    await expect(sweeper.sweep(new AbortController().signal)).resolves.toBeUndefined();

    expect(documentService.commitIfDue).not.toHaveBeenCalled();
  });

  it('does nothing once the signal is aborted', async () => {
    const { sweeper, repo } = setup([ids('a', 2)]);

    await sweeper.sweep(AbortSignal.abort());

    expect(repo.find).not.toHaveBeenCalled();
  });
});
