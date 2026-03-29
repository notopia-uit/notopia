import { AuthorizationService } from '../authorization/authorization.service';
import {
  BLOCKNOTE_SCHEMA,
  type Block,
  type BlockNoteEditor,
  type BlockNoteSchema,
} from '../blocknote/bn-schema.provider';
import { User } from '../common/user';
import { RevisionEntity } from '../revision/revision.entity';
import { StorageService } from '../storage/storage.service';
import { DocumentEntity } from './document.entity';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import {
  Inject,
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
    private readonly storageService: StorageService,
    @InjectDataSource() private readonly dataSource: DataSource,
    private readonly authorizationService: AuthorizationService,
    @Inject(BLOCKNOTE_SCHEMA) private readonly blocknoteSchema: BlockNoteSchema
  ) {}

  toYDoc(entity: DocumentEntity): YDoc {
    const doc = new YDoc();
    applyUpdate(doc, new Uint8Array(entity.data));
    return doc;
  }

  private bufferToBlockNote(data: Buffer, editor: BlockNoteEditor): Block[] {
    const yDoc = new YDoc();
    applyUpdate(yDoc, new Uint8Array(data));
    return editor.yDocToBlocks(yDoc);
  }

  extractTags(editor: BlockNoteEditor): string[] {
    const tags = new Set<string>();
    editor.editor.forEachBlock((block) => {
      if (Array.isArray(block.content)) {
        for (const inlineNode of block.content) {
          if (inlineNode.type === 'tag') {
            tags.add(inlineNode.props.tag);
          }
        }
      }

      return true;
    });

    return Array.from(tags);
  }

  extractOutgoingLinkIds(editor: BlockNoteEditor): string[] {
    const linkIds = new Set<string>();
    editor.editor.forEachBlock((block) => {
      if (Array.isArray(block.content)) {
        for (const inlineNode of block.content) {
          if (inlineNode.type === 'reference') {
            linkIds.add(inlineNode.props.noteId);
          }
        }
      }

      return true;
    });

    return Array.from(linkIds);
  }

  async commitDocument(documentId: string) {
    // TODO: instantiate editor here
    const editor = ServerBlockNoteEditor.create({
      schema: this.blocknoteSchema,
    });
    await this.dataSource.transaction(async (manager) => {
      const document = await manager.findOne(DocumentEntity, {
        where: { id: documentId },
        lock: { mode: 'pessimistic_write' },
      });
      if (!document) {
        throw new NotFoundException(`Document ${documentId} not found`);
      }
      manager.save(RevisionEntity, {
        id: randomUUID(),
        document,
        content: this.bufferToBlockNote(document.data, editor),
      });
      await manager.update(
        DocumentEntity,
        { id: documentId },
        { modified: false }
      );
      // TODO: Inject kafka client to publish document committed event
    });
  }

  async getAttachmentUploadUrl(documentId: string, user: User) {
    const hasPermission = await this.authorizationService.hasNotePermission({
      documentId,
      memberId: user.id,
      permission: 'write',
    });
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
