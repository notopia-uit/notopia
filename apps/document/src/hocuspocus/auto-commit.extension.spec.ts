import type { afterStoreDocumentPayload } from '@hocuspocus/server';
import * as Y from 'yjs';

import { BillingService, BillingUnavailableException, PlanTier } from '../billing';
import { DocumentService } from '../document/document.service';
import { NoteService } from '../note/note.service';
import { AutoCommitExtension } from './auto-commit.extension';
import { applyStoredModified } from './document-metadata';

const entitlements = (autoCommit: boolean) => ({
  plan: autoCommit ? PlanTier.Plus : PlanTier.Free,
  features: { auto_commit: autoCommit },
  limits: { auto_commit_min_interval_seconds: 300 },
});

function setup() {
  const documentService = {
    commitIfDue: vi.fn().mockResolvedValue({
      revisionId: 'r1',
      modifiedCleared: true,
      committedContentHash: 'h',
    }),
    clearModifiedIfCommitted: vi.fn(),
  };
  const noteService = {
    batchGetNoteWorkspaces: vi.fn().mockResolvedValue(new Map([['d1', 'w1']])),
  };
  const billingService = { getEntitlements: vi.fn() };
  const extension = new AutoCommitExtension(
    documentService as unknown as DocumentService,
    noteService as unknown as NoteService,
    billingService as unknown as BillingService
  );
  const document = new Y.Doc();
  applyStoredModified(document, true);
  const payload = { documentName: 'd1', document } as unknown as afterStoreDocumentPayload;
  return { extension, documentService, noteService, billingService, payload };
}

describe('AutoCommitExtension', () => {
  it('skips plans without auto_commit', async () => {
    const { extension, documentService, billingService, payload } = setup();
    billingService.getEntitlements.mockResolvedValue(entitlements(false));
    await extension.afterStoreDocument(payload);
    expect(documentService.commitIfDue).not.toHaveBeenCalled();
  });

  it('commits with the plan interval and clears modified against the committed hash', async () => {
    const { extension, documentService, billingService, payload } = setup();
    billingService.getEntitlements.mockResolvedValue(entitlements(true));
    await extension.afterStoreDocument(payload);
    expect(documentService.commitIfDue).toHaveBeenCalledWith({
      documentId: 'd1',
      trigger: 'auto',
      force: false,
      minIntervalSeconds: 300,
    });
    expect(documentService.clearModifiedIfCommitted).toHaveBeenCalledWith(payload.document, 'h');
  });

  it('does not throw when billing is unavailable', async () => {
    const { extension, documentService, billingService, payload } = setup();
    billingService.getEntitlements.mockRejectedValue(new BillingUnavailableException(undefined));
    await expect(extension.afterStoreDocument(payload)).resolves.toBeUndefined();
    expect(documentService.commitIfDue).not.toHaveBeenCalled();
  });

  it('skips when the workspace is missing', async () => {
    const { extension, documentService, noteService, billingService, payload } = setup();
    noteService.batchGetNoteWorkspaces.mockResolvedValue(new Map());
    await extension.afterStoreDocument(payload);
    expect(billingService.getEntitlements).not.toHaveBeenCalled();
    expect(documentService.commitIfDue).not.toHaveBeenCalled();
  });
});
