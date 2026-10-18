import { Injectable, Logger } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Traceable } from 'nestjs-otel';
import { In, Repository } from 'typeorm';

import {
  BillingService,
  BillingUnavailableException,
  isUnlimited,
  limit,
  PlanLimit,
} from '../billing';
import { NoteService } from '../note/note.service';
import { RevisionEntity } from '../revision/revision.entity';
import { planRevisionDeletes, type RevisionRef } from './revision-purge';

@Injectable()
@Traceable()
export class RevisionPurgeCron {
  private static readonly batchSize = 500;
  private readonly logger = new Logger(RevisionPurgeCron.name);

  constructor(
    @InjectRepository(RevisionEntity) private readonly repo: Repository<RevisionEntity>,
    private readonly noteService: NoteService,
    private readonly billingService: BillingService
  ) {}

  async purge(): Promise<void> {
    try {
      await this.purgeAll();
    } catch (err) {
      if (err instanceof BillingUnavailableException) {
        this.logger.warn({ err }, 'billing unavailable, revision purge skipped until next tick');
        return;
      }
      throw err;
    }
  }

  private async purgeAll(): Promise<void> {
    const plans = await this.billingService.listPlans();
    const days = plans
      .map((plan) => limit(plan, PlanLimit.RevisionHistoryDays))
      .filter((d) => !isUnlimited(d));
    if (days.length === 0) {
      this.logger.debug('no plan limits revision history, revision purge skipped');
      return;
    }
    const minDays = Math.min(...days);
    const rows: { documentId: string }[] = await this.repo
      .createQueryBuilder('r')
      .select('DISTINCT r."documentId"', 'documentId')
      .where('r."createdAt" < now() - make_interval(days => :minDays)', { minDays })
      .getRawMany();
    const documentIds = rows.map((r) => r.documentId);
    let deleted = 0;
    for (let i = 0; i < documentIds.length; i += RevisionPurgeCron.batchSize) {
      deleted += await this.purgeBatch(documentIds.slice(i, i + RevisionPurgeCron.batchSize));
    }
    this.logger.log({ minDays, documents: documentIds.length, deleted }, 'revision purge done');
  }

  private async purgeBatch(documentIds: string[]): Promise<number> {
    const workspaces = await this.noteService.batchGetNoteWorkspaces(documentIds);
    const entitlements = await this.billingService.batchGetEntitlements([
      ...new Set(workspaces.values()),
    ]);
    const retentionDaysByDocument = new Map<string, number>();
    for (const [documentId, workspaceId] of workspaces) {
      const workspaceEntitlements = entitlements.get(workspaceId);
      if (workspaceEntitlements) {
        retentionDaysByDocument.set(
          documentId,
          limit(workspaceEntitlements, PlanLimit.RevisionHistoryDays)
        );
      }
    }
    const revisions: RevisionRef[] = await this.repo
      .createQueryBuilder('r')
      .select('r.id', 'id')
      .addSelect('r."documentId"', 'documentId')
      .addSelect('r."createdAt"', 'createdAt')
      .where('r."documentId" IN (:...documentIds)', { documentIds })
      .getRawMany();
    const ids = planRevisionDeletes(revisions, retentionDaysByDocument, new Date());
    if (ids.length > 0) {
      await this.repo.delete({ id: In(ids) });
    }
    this.logger.debug(
      { documents: documentIds.length, deleted: ids.length },
      'revision purge batch done'
    );
    return ids.length;
  }
}
