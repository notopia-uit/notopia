import { randomUUID } from 'crypto';

import { type MyBlock, type MySchema } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import {
  Inject,
  Injectable,
  Logger,
  NotFoundException,
  UnauthorizedException,
} from '@nestjs/common';
import { ClientKafka } from '@nestjs/microservices';
import { InjectDataSource, InjectRepository } from '@nestjs/typeorm';
import { type ShareDocumentCommittedEvent } from '@notopia-uit/api-gen';
import { Traceable } from 'nestjs-otel';
import { lastValueFrom } from 'rxjs';
import { DataSource, Repository } from 'typeorm';
import { Doc as YDoc, applyUpdate } from 'yjs';

import { AuthorizationService } from '#/authorization/authorization.service';
import { BLOCKNOTE_SCHEMA } from '#/blocknote/blocknote.module';
import { KAFKA_CLIENT } from '#/common/token';
import { RevisionEntity } from '#/revision/revision.entity';
import { StorageService } from '#/storage/storage.service';

import { DocumentEntity } from './document.entity';

@Injectable()
@Traceable()
export class DocumentService {
  private readonly logger = new Logger(DocumentService.name);

  constructor(
    @InjectRepository(DocumentEntity)
    private readonly repo: Repository<DocumentEntity>,
    private readonly storageService: StorageService,
    @InjectDataSource() private readonly dataSource: DataSource,
    private readonly authorizationService: AuthorizationService,
    @Inject(BLOCKNOTE_SCHEMA) private readonly blocknoteSchema: MySchema,
    @Inject(KAFKA_CLIENT) private readonly kafkaClient: ClientKafka
  ) {}

  toYDoc(entity: DocumentEntity): YDoc {
    const doc = new YDoc();
    applyUpdate(doc, new Uint8Array(entity.data));
    return doc;
  }

  private bufferToBlockNote(data: Buffer, editor: ServerBlockNoteEditor): MyBlock[] {
    const yDoc = new YDoc();
    applyUpdate(yDoc, new Uint8Array(data));
    return editor.yDocToBlocks(yDoc);
  }

  extractTagsAndOutgoingLinkIds(editor: ServerBlockNoteEditor): {
    tags: string[];
    outgoingLinkIds: string[];
  } {
    const tags = new Set<string>();
    const linkIds = new Set<string>();

    editor.editor.forEachBlock((block) => {
      if (!Array.isArray(block.content)) {
        return false;
      }
      for (const inlineNode of block.content) {
        switch (inlineNode.type) {
          case 'tag':
            tags.add(inlineNode.props.tag);
            break;
          case 'reference':
            linkIds.add(inlineNode.props.noteId);
            break;
        }
      }

      return true;
    });

    return {
      tags: Array.from(tags),
      outgoingLinkIds: Array.from(linkIds),
    };
  }

  async commitDocument({
    documentId,
    userId,
  }: {
    documentId: string;
    userId: string;
  }): Promise<string> {
    this.logger.debug({ documentId, userId }, `commitDocument: checking permission`);
    const editor = ServerBlockNoteEditor.create({
      schema: this.blocknoteSchema,
    });
    return this.dataSource.transaction(async (manager) => {
      const document = await manager.findOne(DocumentEntity, {
        where: { id: documentId },
        lock: { mode: 'pessimistic_write' },
      });
      if (!document) {
        this.logger.warn({ documentId }, `commitDocument: document not found`);
        throw new NotFoundException(`Document ${documentId} not found`);
      }
      const revisionId = randomUUID();
      this.logger.debug(
        { documentId, userId },
        `commitDocument: saving revision revisionId=${revisionId}`
      );
      await manager.save(RevisionEntity, {
        id: revisionId,
        document,
        content: this.bufferToBlockNote(document.data, editor),
      });
      await manager.update(DocumentEntity, { id: documentId }, { modified: false });
      // TODO: Consider refactor into a module named "EventBus", which manages event topic
      const { tags, outgoingLinkIds } = this.extractTagsAndOutgoingLinkIds(editor);
      this.logger.debug(
        { documentId, userId, tags, outgoingLinkIds },
        `commitDocument: emitting event with tags and outgoingLinkIds`
      );
      await lastValueFrom(
        this.kafkaClient.emit('events.integration.document.document.committed', {
          id: documentId,
          userId,
          tags,
          outgoingLinkIds,
          content: editor.editor.document satisfies MyBlock[],
        } satisfies ShareDocumentCommittedEvent)
      );
      this.logger.log({ documentId, userId, revisionId }, `commitDocument: done`);
      return revisionId;
    });
  }

  async getAttachmentUploadUrl(documentId: string, userId: string) {
    this.logger.debug({ documentId, userId }, `getAttachmentUploadUrl: checking permission`);
    const hasPermission = await this.authorizationService.hasNotePermission({
      documentId,
      memberId: userId,
      permission: 'write',
    });
    if (!hasPermission) {
      this.logger.warn(
        { documentId, userId },
        `getAttachmentUploadUrl: user does not have permission`
      );
      throw new UnauthorizedException(
        `User ${userId} does not have permission to upload attachment to ${documentId}`
      );
    }
    const key = `document-attachments/${documentId}/${randomUUID()}`;
    const { uploadUrl, publicUrl } =
      await this.storageService.generateAttachmentPresignedUploadUrl(key);
    this.logger.log({ documentId, userId, key }, `getAttachmentUploadUrl: generated upload URL`);
    return {
      url: publicUrl,
      uploadUrl: uploadUrl,
    };
  }

  async updateDataById(id: string, data: Buffer): Promise<void> {
    this.logger.debug({ id, dataSize: data.length }, `updateDataById: updating document data`);
    await this.repo.update(id, { data, modified: true });
    this.logger.log({ id, dataSize: data.length }, `updateDataById: document data updated`);
  }

  async getById(id: string): Promise<DocumentEntity | null> {
    this.logger.debug({ id }, `getById: fetching document`);
    const document = await this.repo.findOneBy({ id });
    this.logger.log({ id }, `getById: document fetched`);
    return document;
  }
}
