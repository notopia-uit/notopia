import { HttpUserGuard } from '../common/user.guard';
import { RevisionService } from './revision.service';
import { Injectable, UseGuards } from '@nestjs/common';
import { RevisionApi as RevisionApiDefinition } from '@notopia-uit/api-document-nestjs-server/api';
import type {
  GetRevisions200Response,
  RenameRevisionRequest,
  RevisionWithContent,
} from '@notopia-uit/api-document-nestjs-server/models';
import { Traceable } from 'nestjs-otel';

@Injectable()
@UseGuards(HttpUserGuard)
@Traceable()
export class RevisionApi extends RevisionApiDefinition {
  constructor(private readonly revisionService: RevisionService) {
    super();
  }

  async deleteRevision(revisionId: string): Promise<void> {
    await this.revisionService.deleteRevision(revisionId);
  }

  async getRevisionWithContent(
    revisionId: string
  ): Promise<RevisionWithContent> {
    const revisionEntity = await this.revisionService.getRevision(revisionId);
    return {
      id: revisionEntity.id,
      name: revisionEntity.name,
      content: revisionEntity.content,
      createdAt: revisionEntity.createdAt.toISOString(),
    };
  }

  async getRevisions(
    documentId: string,
    page: number,
    limit: number
  ): Promise<GetRevisions200Response> {
    const result = await this.revisionService.getRevisionsByDocumentId(
      documentId,
      page,
      limit
    );
    const totalPages = Math.ceil(result.total / result.limit);
    return {
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
  }

  async renameRevision(
    revisionId: string,
    renameRevisionRequest: RenameRevisionRequest
  ): Promise<void> {
    await this.revisionService.renameRevision(
      revisionId,
      renameRevisionRequest.name
    );
  }
}
