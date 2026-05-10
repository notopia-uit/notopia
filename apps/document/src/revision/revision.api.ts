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

  async deleteRevision(revisionId: string): Promise<void> {
    this.logger.log(`deleteRevision: received revisionId=${revisionId}`);
    await this.revisionService.deleteRevision(revisionId);
    this.logger.log(`deleteRevision: done revisionId=${revisionId}`);
  }

  async getRevisionWithContent(revisionId: string): Promise<RevisionWithContent> {
    this.logger.log(`getRevisionWithContent: received revisionId=${revisionId}`);
    const revisionEntity = await this.revisionService.getRevision(revisionId);
    const response: RevisionWithContent = {
      id: revisionEntity.id,
      name: revisionEntity.name,
      content: revisionEntity.content,
      createdAt: revisionEntity.createdAt.toISOString(),
    };
    this.logger.log(`getRevisionWithContent: done revisionId=${revisionId}`);
    this.logger.debug({ id: response.id, name: response.name }, 'getRevisionWithContent: response');
    return response;
  }

  async getRevisions(
    documentId: string,
    page: number,
    limit: number
  ): Promise<GetRevisions200Response> {
    this.logger.log(`getRevisions: received documentId=${documentId} page=${page} limit=${limit}`);
    const result = await this.revisionService.getRevisionsByDocumentId(documentId, page, limit);
    const totalPages = Math.ceil(result.total / result.limit);
    const response: GetRevisions200Response = {
      data: result.data.map((r) => ({
        id: r.id,
        name: r.name,
        createdAt: r.createdAt.toISOString(),
      })),
      pagination: {
        page: result.page,
        limit: result.limit,
        total: result.total,
        totalPages,
        hasNext: result.page < totalPages,
        hasPrev: result.page > 1,
      },
    };
    this.logger.log(`getRevisions: done documentId=${documentId} total=${result.total}`);
    this.logger.debug(
      { count: response.data.length, pagination: response.pagination },
      'getRevisions: response'
    );
    return response;
  }

  async renameRevision(
    revisionId: string,
    renameRevisionRequest: RenameRevisionRequest
  ): Promise<void> {
    this.logger.log(`renameRevision: received revisionId=${revisionId}`);
    await this.revisionService.renameRevision(revisionId, renameRevisionRequest.name);
    this.logger.log(`renameRevision: done revisionId=${revisionId}`);
  }
}
