import { PutObjectCommand, S3Client } from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Injectable, NotFoundException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { randomUUID } from 'crypto';
import { Traceable } from 'nestjs-otel';
import { applyUpdate, Doc as YDoc } from 'yjs';

import { S3Config } from '../config/config';
import { DocumentEntity } from './document.entity';
import { DocumentRepository } from './document.repository';

export interface AttachmentUploadUrl {
  url: string;
  uploadUrl: string;
}

@Injectable()
@Traceable()
export class DocumentService {
  private readonly bucketName: string;
  private readonly s3Endpoint: string;
  private static readonly s3UrlExpirationSeconds = 3600;

  constructor(
    private readonly documentRepository: DocumentRepository,
    private readonly editor: ServerBlockNoteEditor,
    private readonly s3Client: S3Client,
    configService: ConfigService
  ) {
    const s3Config = configService.get<S3Config>('s3')!;
    this.bucketName = s3Config.bucketName;
    this.s3Endpoint = s3Config.endpoint;
  }

  toYDoc(entity: DocumentEntity): YDoc {
    const doc = new YDoc();
    applyUpdate(doc, new Uint8Array(entity.data));
    return doc;
  }

  yDocToBlockNote(yDoc: YDoc) {
    return this.editor.yDocToBlocks(yDoc);
  }

  extractTags(_content: object[]): string[] {
    return [];
  }

  extractOutgoingLinkIds(_content: object[]): string[] {
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
    const key = `document-attachments/${documentId}/${randomUUID()}`;
    const command = new PutObjectCommand({
      Bucket: this.bucketName,
      Key: key,
    });
    const presignedUrl = await getSignedUrl(this.s3Client, command, {
      expiresIn: DocumentService.s3UrlExpirationSeconds,
    });
    const attachmentUrl = `${this.s3Endpoint}/${this.bucketName}/${key}`;
    return {
      url: attachmentUrl,
      uploadUrl: presignedUrl,
    };
  }
}
