import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Injectable, NotFoundException } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { applyUpdate, Doc as YDoc } from 'yjs';

import { DocumentEntity } from '../domain/document.entity';
import { DocumentRepository } from '../infrastructure/database/document.repository';

export interface AttachmentUploadUrl {
  url: string;
}

@Injectable()
export class DocumentService {
  constructor(
    private readonly documentRepository: DocumentRepository,
    private readonly editor: ServerBlockNoteEditor
  ) {}

  private toYDoc(entity: DocumentEntity): YDoc {
    const doc = new YDoc();
    applyUpdate(doc, new Uint8Array(entity.data));
    return doc;
  }

  yDocToBlockNote(yDoc: YDoc) {
    return this.editor.yDocToBlocks(yDoc);
  }

  extractTags(content: object[]): string[] {
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
