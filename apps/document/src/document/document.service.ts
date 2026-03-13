import { PutObjectCommand, S3Client } from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import type { Client } from '@connectrpc/connect';
import {
  Inject,
  Injectable,
  NotFoundException,
  UnauthorizedException,
} from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { InjectDataSource } from '@nestjs/typeorm';
import {
  AuthorizationService,
  NotePermission,
} from '@notopia-uit/pb/authorization';
import { randomUUID } from 'crypto';
import { Traceable } from 'nestjs-otel';
import { DataSource } from 'typeorm';
import { applyUpdate, Doc as YDoc } from 'yjs';

import { AUTHORIZATION_SERVICE } from '../authorization/authorization.module';
import { User } from '../common/user';
import { S3Config } from '../config/config';
import { RevisionEntity } from '../revision/revision.entity';
import { DocumentEntity } from './document.entity';

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
    private readonly editor: ServerBlockNoteEditor,
    private readonly s3Client: S3Client,
    @InjectDataSource() private readonly dataSource: DataSource,
    @Inject(AUTHORIZATION_SERVICE)
    private readonly authorizationService: Client<typeof AuthorizationService>,
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

  BufferToBlockNote(data: Buffer) {
    const yDoc = new YDoc();
    applyUpdate(yDoc, new Uint8Array(data));
    return this.editor.yDocToBlocks(yDoc);
  }

  extractTags(_content: object[]): string[] {
    return [];
  }

  extractOutgoingLinkIds(_content: object[]): string[] {
    return [];
  }

  async commitDocument(documentId: string) {
    await this.dataSource.transaction(async (manager) => {
      const document = await manager.findOne(DocumentEntity, {
        where: { id: documentId },
        lock: { mode: 'pessimistic_write' },
      });
      if (!document) {
        throw new NotFoundException(`Document ${documentId} not found`);
      }
      const revision = manager.create(RevisionEntity, {
        id: randomUUID(),
        document,
        content: this.BufferToBlockNote(document.data),
      });
      await manager.save(revision);
      await manager.update(
        DocumentEntity,
        { id: documentId },
        { modified: false }
      );
    });
  }

  async getAttachmentUploadUrl(
    documentId: string,
    user: User
  ): Promise<AttachmentUploadUrl> {
    const permissionRes = await this.authorizationService.hasNotePermission({
      noteId: documentId,
      permission: NotePermission.WRITE,
      memberId: user.id,
    });
    if (!permissionRes.hasPermission) {
      throw new UnauthorizedException(
        `User ${user.id} does not have permission to upload attachment to ${documentId}`
      );
    }
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
