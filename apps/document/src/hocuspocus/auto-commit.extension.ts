import type { Extension, afterStoreDocumentPayload } from '@hocuspocus/server';
import { Logger } from '@nestjs/common';

import {
  BillingService,
  BillingUnavailableException,
  feature,
  limit,
  PlanFeature,
  PlanLimit,
} from '../billing';
import { DocumentService } from '../document/document.service';
import { NoteService } from '../note/note.service';
import { readModified } from './document-metadata';
import type { HocuspocusContext } from './hocuspocus-context';

// Hooks run by descending priority and the Database extension has the default (100). This must be
// lower so the content is stored before commitIfDue reads it from the database.
export const AUTO_COMMIT_PRIORITY = 50;

export class AutoCommitExtension implements Extension {
  readonly priority = AUTO_COMMIT_PRIORITY;
  private readonly logger = new Logger(AutoCommitExtension.name);

  constructor(
    private readonly documentService: DocumentService,
    private readonly noteService: NoteService,
    private readonly billingService: BillingService
  ) {}

  async afterStoreDocument({
    documentName,
    document,
  }: afterStoreDocumentPayload<HocuspocusContext>): Promise<void> {
    if (!readModified(document)) return;
    try {
      // HACK: bypasses the auth check (internal lookup with no user). The proper fix is a
      // read-only service-account role in authorization so getWorkspaceByNote can be used.
      const workspaceIds = await this.noteService.batchGetNoteWorkspaces([documentName]);
      const workspaceId = workspaceIds.get(documentName);
      if (!workspaceId) {
        this.logger.warn({ documentId: documentName }, 'workspace not found, auto-commit skipped');
        return;
      }
      const entitlements = await this.billingService.getEntitlements(workspaceId);
      if (!feature(entitlements, PlanFeature.AutoCommit)) {
        this.logger.debug(
          { documentId: documentName, workspaceId, plan: entitlements.plan },
          'auto-commit not in plan, skipped'
        );
        return;
      }
      const { modifiedCleared, committedContentHash } = await this.documentService.commitIfDue({
        documentId: documentName,
        trigger: 'auto',
        force: false,
        minIntervalSeconds: limit(entitlements, PlanLimit.AutoCommitMinIntervalSeconds),
      });
      if (modifiedCleared && committedContentHash) {
        this.documentService.clearModifiedIfCommitted(document, committedContentHash);
      }
    } catch (err) {
      if (err instanceof BillingUnavailableException) {
        this.logger.warn(
          { err, documentId: documentName },
          'billing unavailable, auto-commit skipped'
        );
        return;
      }
      this.logger.error({ err, documentId: documentName }, 'auto-commit skipped');
    }
  }
}
