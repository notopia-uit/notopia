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

  async save(revision: RevisionEntity): Promise<void> {
    await this.repo.save(revision);
  }

  async getById(revisionId: string): Promise<RevisionEntity | null> {
    return this.repo.findOneBy({ id: revisionId });
  }

  async getByDocumentId(documentId: string): Promise<RevisionEntity[]> {
    return this.repo.findBy({ documentId });
  }

  async delete(revisionId: string): Promise<void> {
    await this.repo.delete({ id: revisionId });
  }
}
