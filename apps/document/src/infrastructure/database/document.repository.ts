import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';

import { DocumentEntity } from '../../domain/document.entity';

@Injectable()
export class DocumentRepository {
  constructor(
    @InjectRepository(DocumentEntity)
    private readonly repo: Repository<DocumentEntity>
  ) {}

  async save(document: DocumentEntity) {
    await this.repo.save(document);
  }

  async updateDataById(id: string, data: Buffer) {
    await this.repo.update(id, { data });
  }

  async getById(id: string) {
    return this.repo.findOneBy({ id });
  }
}
