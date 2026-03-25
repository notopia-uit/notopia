import { AuthorizationService } from '../authorization/authorization.service';
import { User } from '../common/user';
import { RevisionEntity } from '../revision/revision.entity';
import { StorageService } from '../storage/storage.service';
import { DocumentEntity } from './document.entity';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import {
  Injectable,
  NotFoundException,
  UnauthorizedException,
} from '@nestjs/common';
import { InjectDataSource } from '@nestjs/typeorm';
import { randomUUID } from 'crypto';
import { Traceable } from 'nestjs-otel';
import { DataSource } from 'typeorm';
import { Doc as YDoc, applyUpdate } from 'yjs';

@Injectable()
@Traceable()
export class DocumentService {
  constructor(
    private readonly editor: ServerBlockNoteEditor,
    private readonly storageService: StorageService,
    @InjectDataSource() private readonly dataSource: DataSource,
    private readonly authorizationService: AuthorizationService
  ) {}

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

  async getAttachmentUploadUrl(documentId: string, user: User) {
    const hasPermission =
      await this.authorizationService.hasWriteNotePermission(
        documentId,
        user.id
      );
    if (!hasPermission) {
      throw new UnauthorizedException(
        `User ${user.id} does not have permission to upload attachment to ${documentId}`
      );
    }
    const key = `document-attachments/${documentId}/${randomUUID()}`;
    const { uploadUrl, publicUrl } =
      await this.storageService.generateAttachmentPresignedUploadUrl(key);
    return {
      url: publicUrl,
      uploadUrl: uploadUrl,
    };
  }
}
