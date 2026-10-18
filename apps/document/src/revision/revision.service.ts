import { Injectable, Logger } from '@nestjs/common';
import { InjectDataSource } from '@nestjs/typeorm';
import { Traceable } from 'nestjs-otel';
import { DataSource } from 'typeorm';

import { type RevisionTrigger } from '../document/commit-policy';
import { RevisionNotFoundException } from './revision-not-found.exception';
import { RevisionEntity } from './revision.entity';

export interface PaginatedRevisions {
  data: RevisionEntity[];
  page: number;
  currentTotal: number;
  total: number;
}

@Injectable()
@Traceable()
export class RevisionService {
  private readonly logger = new Logger(RevisionService.name);

  constructor(@InjectDataSource() private readonly dataSource: DataSource) {}

  async getRevision(revisionId: string): Promise<RevisionEntity> {
    this.logger.debug({ revisionId }, 'getRevision');
    const revision = await this.dataSource
      .getRepository(RevisionEntity)
      .findOneBy({ id: revisionId });
    if (!revision) {
      throw new RevisionNotFoundException(revisionId);
    }
    return revision;
  }

  async getRevisionsByDocumentId(
    documentId: string,
    page: number,
    limit: number,
    trigger?: RevisionTrigger
  ): Promise<PaginatedRevisions> {
    this.logger.debug({ documentId, page, limit, trigger }, 'getRevisionsByDocumentId');
    const [revisions, total] = await this.dataSource.getRepository(RevisionEntity).findAndCount({
      where: {
        document: {
          id: documentId,
        },
        ...(trigger && { trigger }),
      },
      order: { createdAt: 'DESC' },
      skip: (page - 1) * limit,
      take: limit,
    });
    this.logger.debug({ documentId, total }, 'getRevisionsByDocumentId: found');
    return {
      data: revisions,
      page,
      currentTotal: revisions.length,
      total,
    };
  }

  async renameRevision(revisionId: string, name: string | null): Promise<void> {
    this.logger.debug({ revisionId, name }, 'renameRevision');
    const result = await this.dataSource.getRepository(RevisionEntity).update(revisionId, { name });
    if (result.affected === 0) {
      throw new RevisionNotFoundException(revisionId);
    }
    this.logger.log({ revisionId }, 'renameRevision: done');
  }

  async deleteRevision(revisionId: string): Promise<void> {
    this.logger.debug({ revisionId }, 'deleteRevision');
    const result = await this.dataSource.getRepository(RevisionEntity).delete(revisionId);
    if (result.affected === 0) {
      throw new RevisionNotFoundException(revisionId);
    }
    this.logger.log({ revisionId }, 'deleteRevision: done');
  }
}
