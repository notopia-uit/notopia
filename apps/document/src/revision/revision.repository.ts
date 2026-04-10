import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Traceable } from 'nestjs-otel';
import { Repository } from 'typeorm';

import { RevisionEntity } from './revision.entity';

@Injectable()
@Traceable()
export class RevisionRepository {
  constructor(
    @InjectRepository(RevisionEntity)
    private readonly repo: Repository<RevisionEntity>
  ) {}

  async save(revision: RevisionEntity) {
    await this.repo.save(revision);
  }

  async getById(revisionId: string) {
    return this.repo.findOneBy({ id: revisionId });
  }

  async getByDocumentId(documentId: string, page: number, limit: number) {
    return this.repo.findAndCount({
      where: {
        document: {
          id: documentId,
        },
      },
      take: limit,
      skip: (page - 1) * limit,
      order: {
        createdAt: 'DESC',
      },
    });
  }

  async delete(revisionId: string) {
    await this.repo.delete({ id: revisionId });
  }
}
