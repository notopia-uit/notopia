import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { InjectRepository } from '@nestjs/typeorm';
import { Traceable } from 'nestjs-otel';
import { LessThan, MoreThan, Repository } from 'typeorm';

import { BillingService, BillingUnavailableException, feature, PlanFeature } from '../billing';
import { CRON_CONFIG, CronConfig } from '../config';
import { DocumentEntity } from '../document/document.entity';
import { DocumentService } from '../document/document.service';
import { HocuspocusService } from '../hocuspocus/hocuspocus.service';
import { NoteService } from '../note/note.service';
import { abortReason } from './abort-reason';

@Injectable()
@Traceable()
export class SweeperCron {
  static readonly batchSize = 500;
  private readonly logger = new Logger(SweeperCron.name);
  private readonly cronConfig: CronConfig;

  constructor(
    @InjectRepository(DocumentEntity) private readonly repo: Repository<DocumentEntity>,
    private readonly documentService: DocumentService,
    private readonly hocuspocusService: HocuspocusService,
    private readonly noteService: NoteService,
    private readonly billingService: BillingService,
    configService: ConfigService
  ) {
    const cronConfig = configService.get<CronConfig>(CRON_CONFIG);
    if (!cronConfig) {
      throw new Error('CRON_CONFIG not found');
    }
    this.cronConfig = cronConfig;
  }

  async sweep(signal: AbortSignal): Promise<void> {
    const idleBefore = new Date(Date.now() - this.cronConfig.sweeper.idleSeconds * 1000);
    let afterId: string | undefined;
    let committed = 0;
    try {
      while (!signal.aborted) {
        const stale = await this.repo.find({
          select: { id: true },
          where: {
            modified: true,
            updatedAt: LessThan(idleBefore),
            ...(afterId && { id: MoreThan(afterId) }),
          },
          order: { id: 'ASC' },
          take: SweeperCron.batchSize,
        });
        if (stale.length === 0) break;
        afterId = stale.at(-1)!.id;
        committed += await this.sweepBatch(
          stale.map((d) => d.id),
          signal
        );
        if (stale.length < SweeperCron.batchSize) break;
      }
    } catch (err) {
      if (err instanceof BillingUnavailableException) {
        this.logger.warn({ err }, 'billing unavailable, sweeper skipped until next tick');
        return;
      }
      throw err;
    }
    if (signal.aborted) {
      this.logger.warn({ reason: abortReason(signal), committed }, 'sweeper stopped');
      return;
    }
    this.logger.log({ committed }, 'sweeper done');
  }

  private async sweepBatch(documentIds: string[], signal: AbortSignal): Promise<number> {
    const workspaces = await this.noteService.batchGetNoteWorkspaces(documentIds);
    const workspaceIds = [...new Set(workspaces.values())];
    const entitlements = await this.billingService.batchGetEntitlements(workspaceIds);
    this.warnMissing(documentIds.length - workspaces.size, workspaceIds.length - entitlements.size);
    let committed = 0;
    for (const [documentId, workspaceId] of workspaces) {
      const workspaceEntitlements = entitlements.get(workspaceId);
      if (!workspaceEntitlements || !feature(workspaceEntitlements, PlanFeature.AutoCommit)) {
        continue;
      }
      if (signal.aborted) break;
      try {
        const { revisionId, modifiedCleared, committedContentHash } =
          await this.documentService.commitIfDue({
            documentId,
            trigger: 'sweeper',
            force: true,
            minIntervalSeconds: 0,
          });
        if (modifiedCleared && committedContentHash) {
          this.hocuspocusService.clearModifiedIfCommitted(documentId, committedContentHash);
        }
        if (revisionId) committed++;
      } catch (err) {
        this.logger.error({ err, documentId }, 'sweeper commit failed');
      }
    }
    return committed;
  }

  private warnMissing(documentsWithoutNote: number, workspacesWithoutEntitlements: number) {
    if (documentsWithoutNote > 0 || workspacesWithoutEntitlements > 0) {
      this.logger.warn(
        { documentsWithoutNote, workspacesWithoutEntitlements },
        'sweeper skipped documents missing from note or billing'
      );
    }
  }
}
