import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';

import { DocumentEntity } from './document.entity';

@Injectable()
export class DocumentRepository {
  constructor(
    @InjectRepository(DocumentEntity)
    private readonly repo: Repository<DocumentEntity>
  ) {}

  async save(document: DocumentEntity): Promise<void> {
    await this.repo.save(document);
  }

  async updateDataById(id: string, data: Buffer): Promise<void> {
    await this.repo.update(id, { data });
  }

  async getById(id: string): Promise<DocumentEntity | null> {
    return this.repo.findOneBy({ id });
  }
}
