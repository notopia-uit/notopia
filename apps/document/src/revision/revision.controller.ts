import { Injectable } from '@nestjs/common';
import {
  GetRevisions200Response,
  RenameRevisionRequest,
  Revision,
  RevisionApi,
} from '@notopia-uit/api-document-nestjs-server';

import { RevisionService } from './revision.service';

@Injectable()
export class RevisionController extends RevisionApi {
  constructor(private readonly revisionService: RevisionService) {
    super();
  }

  async deleteRevision(revisionId: string, _request: Request): Promise<void> {
    await this.revisionService.deleteRevision(revisionId);
  }

  async getRevision(revisionId: string, _request: Request): Promise<Revision> {
    const revision = await this.revisionService.getRevision(revisionId);
    return {
      id: revision.id,
      name: revision.name,
      content: revision.content,
      createdAt: revision.createdAt,
    };
  }

  async getRevisions(
    documentId: string,
    page: number,
    limit: number,
    _request: Request
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
        content: r.content,
        createdAt: r.createdAt,
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
    renameRevisionRequest: RenameRevisionRequest,
    _request: Request
  ): Promise<void> {
    await this.revisionService.renameRevision(
      revisionId,
      renameRevisionRequest.name
    );
  }
}
