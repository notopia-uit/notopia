import type {
  afterStoreDocumentPayload,
  Extension,
  onStoreDocumentPayload,
} from '@hocuspocus/server';
import { Logger } from '@nestjs/common';

import { DocumentService } from '../document/document.service';
import { ContributorTracker } from './contributor-tracker';

// Must run before the Hocuspocus Redis extension (priority 1000): an instance that misses the
// Redis store lock gets SkipFurtherHooksError, so later hooks would never flush its editors.
export const CONTRIBUTOR_FLUSH_PRIORITY = 1100;

export class ContributorFlushExtension implements Extension {
  readonly priority = CONTRIBUTOR_FLUSH_PRIORITY;
  private readonly logger = new Logger(ContributorFlushExtension.name);

  constructor(
    private readonly tracker: ContributorTracker,
    private readonly documentService: DocumentService
  ) {}

  async onStoreDocument({ documentName }: onStoreDocumentPayload): Promise<void> {
    await this.flush(documentName);
  }

  // A brand-new document has no row during its first onStoreDocument (the Database extension
  // creates it later in the same store), so that flush finds nothing to update. Retry once the
  // store is done; otherwise a document unloaded after a single store would lose its editors.
  async afterStoreDocument({ documentName }: afterStoreDocumentPayload): Promise<void> {
    await this.flush(documentName);
  }

  private async flush(documentName: string): Promise<void> {
    const contributors = this.tracker.peek(documentName);
    if (contributors.length === 0) return;
    try {
      const flushed = await this.documentService.mergePendingContributors(
        documentName,
        contributors
      );
      if (flushed) this.tracker.clear(documentName, contributors);
    } catch (err) {
      this.logger.error(
        { err, documentId: documentName },
        'contributor flush failed, retrying on next store'
      );
    }
  }
}
