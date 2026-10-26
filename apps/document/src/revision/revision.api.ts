import { BadRequestException, Logger } from '@nestjs/common';
import { RevisionApi as RevisionApiDefinition } from '@notopia-uit/api-document-nestjs-server/api';
import {
  type GetRevisions200Response,
  type RenameRevisionRequest,
  type Revision,
  RevisionTrigger as RevisionTriggerDto,
  type RevisionWithContent,
} from '@notopia-uit/api-document-nestjs-server/models';
import { Traceable } from 'nestjs-otel';

import { type RevisionTrigger } from '../document/commit-policy';
import { RevisionEntity } from './revision.entity';
import { RevisionService } from './revision.service';

const toRevisionTrigger = (trigger: RevisionTriggerDto): RevisionTrigger => {
  switch (trigger) {
    case RevisionTriggerDto.Manual:
      return 'manual';
    case RevisionTriggerDto.Auto:
      return 'auto';
    case RevisionTriggerDto.Sweeper:
      return 'sweeper';
    default:
      throw new BadRequestException(`Invalid revision trigger: ${String(trigger)}`);
  }
};

const toRevisionTriggerDto = (trigger: RevisionTrigger): RevisionTriggerDto => {
  switch (trigger) {
    case 'manual':
      return RevisionTriggerDto.Manual;
    case 'auto':
      return RevisionTriggerDto.Auto;
    case 'sweeper':
      return RevisionTriggerDto.Sweeper;
  }
};

const toRevisionDto = (revision: RevisionEntity): Revision => ({
  id: revision.id,
  name: revision.name,
  createdAt: revision.createdAt.toISOString(),
  authorId: revision.authorId,
  trigger: toRevisionTriggerDto(revision.trigger),
  contributorIds: revision.contributorIds,
});

@Traceable()
export class RevisionApi extends RevisionApiDefinition {
  private readonly logger = new Logger(RevisionApi.name);

  constructor(private readonly revisionService: RevisionService) {
    super();
  }

  override async deleteRevision(revisionId: string): Promise<void> {
    this.logger.log({ revisionId }, 'deleteRevision: received');
    try {
      await this.revisionService.deleteRevision(revisionId);
      this.logger.log({ revisionId }, 'deleteRevision: done');
    } catch (error) {
      this.logger.error({ err: error, revisionId }, 'deleteRevision: error');
      throw error;
    }
  }

  override async getRevisionWithContent(revisionId: string): Promise<RevisionWithContent> {
    this.logger.log({ revisionId }, 'getRevisionWithContent: received');
    try {
      const revisionEntity = await this.revisionService.getRevision(revisionId);
      const response: RevisionWithContent = {
        ...toRevisionDto(revisionEntity),
        content: revisionEntity.content,
      };
      this.logger.log({ revisionId }, 'getRevisionWithContent: done');
      this.logger.debug(
        { id: response.id, name: response.name },
        'getRevisionWithContent: response'
      );
      return response;
    } catch (error) {
      this.logger.error({ err: error, revisionId }, 'getRevisionWithContent: error');
      throw error;
    }
  }

  override async getRevisions(
    documentId: string,
    page: number,
    limit: number,
    trigger: RevisionTriggerDto | undefined
  ): Promise<GetRevisions200Response> {
    this.logger.log({ documentId, page, limit, trigger }, 'getRevisions: received');
    try {
      const result = await this.revisionService.getRevisionsByDocumentId(
        documentId,
        page,
        limit,
        trigger === undefined ? undefined : toRevisionTrigger(trigger)
      );
      const totalPages = Math.ceil(result.total / limit);
      const response: GetRevisions200Response = {
        data: result.data.map(toRevisionDto),
        pagination: {
          page: result.page,
          totalPages,
          currentTotal: result.currentTotal,
          total: result.total,
          hasNext: result.page < totalPages,
          hasPrev: result.page > 1,
        },
      };
      this.logger.log({ documentId, total: result.total }, 'getRevisions: done');
      this.logger.debug(
        { count: response.data.length, pagination: response.pagination },
        'getRevisions: response'
      );
      return response;
    } catch (error) {
      this.logger.error({ err: error, documentId }, 'getRevisions: error');
      throw error;
    }
  }

  override async renameRevision(
    revisionId: string,
    renameRevisionRequest: RenameRevisionRequest
  ): Promise<void> {
    this.logger.log({ revisionId }, 'renameRevision: received');
    try {
      await this.revisionService.renameRevision(revisionId, renameRevisionRequest.name);
      this.logger.log({ revisionId }, 'renameRevision: done');
    } catch (error) {
      this.logger.error({ err: error, revisionId }, 'renameRevision: error');
      throw error;
    }
  }
}
