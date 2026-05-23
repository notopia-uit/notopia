import { Logger } from '@nestjs/common';
import { RevisionApi as RevisionApiDefinition } from '@notopia-uit/api-document-nestjs-server/api';
import type {
  GetRevisions200Response,
  RenameRevisionRequest,
  RevisionWithContent,
} from '@notopia-uit/api-document-nestjs-server/models';
import { Traceable } from 'nestjs-otel';

import { RevisionService } from './revision.service';

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
        id: revisionEntity.id,
        name: revisionEntity.name,
        content: revisionEntity.content,
        createdAt: revisionEntity.createdAt.toISOString(),
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
    limit: number
  ): Promise<GetRevisions200Response> {
    this.logger.log({ documentId, page, limit }, 'getRevisions: received');
    try {
      const result = await this.revisionService.getRevisionsByDocumentId(documentId, page, limit);
      const totalPages = Math.ceil(result.total / result.currentTotal);
      const response: GetRevisions200Response = {
        data: result.data.map((r) => ({
          id: r.id,
          name: r.name,
          createdAt: r.createdAt.toISOString(),
        })),
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
