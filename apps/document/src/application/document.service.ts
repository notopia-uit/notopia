import { Injectable, NotFoundException } from '@nestjs/common';
import { randomUUID } from 'crypto';

import { DocumentEntity } from '../domain/document.entity';
import { DocumentRepository } from '../infrastructure/database/document.repository';

export interface AttachmentUploadUrl {
  url: string;
}

export interface TagModel {
  id: string;
  name: string;
}

@Injectable()
export class DocumentService {
  constructor(private readonly documentRepository: DocumentRepository) {}

  extractTags(content: object[]): TagModel[] {
    return [];
  }

  extractOutgoingLinkIds(content: object[]): string[] {
    return [];
  }

  async createDocument(data: Buffer): Promise<DocumentEntity> {
    const document = new DocumentEntity();
    document.id = randomUUID();
    document.data = data;
    await this.documentRepository.save(document);
    return document;
  }

  async getDocument(documentId: string): Promise<DocumentEntity> {
    const document = await this.documentRepository.getById(documentId);
    if (!document) {
      throw new NotFoundException(`Document ${documentId} not found`);
    }
    return document;
  }

  async getAttachmentUploadUrl(
    documentId: string
  ): Promise<AttachmentUploadUrl> {
    return { url: '' };
  }
}
