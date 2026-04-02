import { RevisionEntity } from './revision.entity';
import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectDataSource } from '@nestjs/typeorm';
import { Traceable } from 'nestjs-otel';
import { DataSource } from 'typeorm';

export interface PaginatedRevisions {
  data: RevisionEntity[];
  page: number;
  limit: number;
  total: number;
}

@Injectable()
@Traceable()
export class RevisionService {
  constructor(@InjectDataSource() private readonly dataSource: DataSource) {}

  async getRevision(revisionId: string): Promise<RevisionEntity> {
    const revision = await this.dataSource
      .getRepository(RevisionEntity)
      .findOneBy({ id: revisionId });
    if (!revision) {
      throw new NotFoundException(`Revision ${revisionId} not found`);
    }
    return revision;
  }

  async getRevisionsByDocumentId(
    documentId: string,
    page: number,
    limit: number
  ): Promise<PaginatedRevisions> {
    const [revisions, total] = await this.dataSource
      .getRepository(RevisionEntity)
      .findAndCount({
        where: {
          document: {
            id: documentId,
          },
        },
        skip: (page - 1) * limit,
        take: limit,
      });
    return {
      data: revisions,
      page,
      limit,
      total,
    };
  }

  // No checking exist first
  async renameRevision(revisionId: string, name: string | null): Promise<void> {
    const result = await this.dataSource
      .getRepository(RevisionEntity)
      .update(revisionId, { name });
    if (result.affected === 0) {
      throw new NotFoundException(`Revision ${revisionId} not found`);
    }
  }

  async deleteRevision(revisionId: string): Promise<void> {
    await this.dataSource.getRepository(RevisionEntity).softDelete(revisionId);
  }
}
