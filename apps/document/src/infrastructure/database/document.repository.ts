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

  async save(document: DocumentEntity): Promise<void> {
    await this.repo.save(document);
  }

  async getById(documentId: string): Promise<DocumentEntity | null> {
    return this.repo.findOneBy({ id: documentId });
  }
}
