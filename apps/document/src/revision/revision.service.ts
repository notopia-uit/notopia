import { Injectable, Logger, NotFoundException } from '@nestjs/common';
import { InjectDataSource } from '@nestjs/typeorm';
import { Traceable } from 'nestjs-otel';
import { DataSource } from 'typeorm';

import { RevisionEntity } from './revision.entity';

export interface PaginatedRevisions {
  data: RevisionEntity[];
  page: number;
  limit: number;
  total: number;
}

@Injectable()
@Traceable()
export class RevisionService {
  private readonly logger = new Logger(RevisionService.name);

  constructor(@InjectDataSource() private readonly dataSource: DataSource) {}

  async getRevision(revisionId: string): Promise<RevisionEntity> {
    this.logger.debug(`getRevision: revisionId=${revisionId}`);
    const revision = await this.dataSource
      .getRepository(RevisionEntity)
      .findOneBy({ id: revisionId });
    if (!revision) {
      this.logger.warn(`getRevision: not found revisionId=${revisionId}`);
      throw new NotFoundException(`Revision ${revisionId} not found`);
    }
    return revision;
  }

  async getRevisionsByDocumentId(
    documentId: string,
    page: number,
    limit: number
  ): Promise<PaginatedRevisions> {
    this.logger.debug(
      `getRevisionsByDocumentId: documentId=${documentId} page=${page} limit=${limit}`
    );
    const [revisions, total] = await this.dataSource.getRepository(RevisionEntity).findAndCount({
      where: {
        document: {
          id: documentId,
        },
      },
      skip: (page - 1) * limit,
      take: limit,
    });
    this.logger.debug(`getRevisionsByDocumentId: found total=${total} documentId=${documentId}`);
    return {
      data: revisions,
      page,
      limit,
      total,
    };
  }

  // No checking exist first
  async renameRevision(revisionId: string, name: string | null): Promise<void> {
    this.logger.debug(`renameRevision: revisionId=${revisionId} name=${name}`);
    const result = await this.dataSource.getRepository(RevisionEntity).update(revisionId, { name });
    if (result.affected === 0) {
      this.logger.warn(`renameRevision: not found revisionId=${revisionId}`);
      throw new NotFoundException(`Revision ${revisionId} not found`);
    }
    this.logger.log(`renameRevision: done revisionId=${revisionId}`);
  }

  async deleteRevision(revisionId: string): Promise<void> {
    this.logger.debug(`deleteRevision: revisionId=${revisionId}`);
    await this.dataSource.getRepository(RevisionEntity).softDelete(revisionId);
    this.logger.log(`deleteRevision: done revisionId=${revisionId}`);
  }
}
