import type { MySchema } from '@blocknote/core';
import { Inject, Injectable, Logger } from '@nestjs/common';
import { ClientKafka } from '@nestjs/microservices';
import { InjectDataSource, InjectRepository } from '@nestjs/typeorm';
import { type DocumentCommittedEvent } from '@notopia-uit/api-share-gen';
import { Traceable } from 'nestjs-otel';
import { lastValueFrom } from 'rxjs';
import { DataSource, Repository } from 'typeorm';
import { v7 as uuidv7 } from 'uuid';

import { AuthorizationService } from '../authorization/authorization.service';
import { BLOCKNOTE_SCHEMA } from '../blocknote';
import { BlocknoteEditorService } from '../blocknote/blocknote-editor.service';
import { KAFKA_CLIENT } from '../kafka/token';
import { RevisionEntity } from '../revision/revision.entity';
import { StorageService } from '../storage/storage.service';
import { DocumentNotFoundException } from './document-not-found.exception';
import { DocumentPermissionException } from './document-permission.exception';
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
    @Inject(KAFKA_CLIENT) private readonly kafkaClient: ClientKafka,
    @Inject(BLOCKNOTE_SCHEMA) private readonly blocknoteSchema: MySchema
  ) {}

  async commitDocument({
    documentId,
    userId,
  }: {
    documentId: string;
    userId: string;
  }): Promise<string> {
    this.logger.debug({ documentId, userId }, 'commitDocument: checking permission');
    return this.dataSource.transaction(async (manager) => {
      const document = await manager.findOne(DocumentEntity, {
        where: { id: documentId },
        lock: { mode: 'pessimistic_write' },
      });
      if (!document) {
        throw new DocumentNotFoundException(documentId);
      }
      const revisionId = uuidv7();
      this.logger.debug({ documentId, revisionId }, 'commitDocument: saving revision');
      const blocknoteEditorService = new BlocknoteEditorService({
        schema: this.blocknoteSchema,
        initialContent: document.data,
      });
      await manager.save(RevisionEntity, {
        id: revisionId,
        document,
        content: blocknoteEditorService.blocks(),
      });
      await manager.update(DocumentEntity, { id: documentId }, { modified: false });
      // TODO: Consider refactor into a module named "EventBus", which manages event topic
      const { tags, outgoingLinkIds } = blocknoteEditorService.extractTagsAndOutgoingLinkIds();
      this.logger.debug({ documentId, tags, outgoingLinkIds }, 'commitDocument: emitting event');
      await lastValueFrom(
        this.kafkaClient.emit('events.integration.document.document.committed', {
          id: documentId,
          userId,
          tags,
          outgoingLinkIds,
          content: blocknoteEditorService.blocks(),
        } satisfies DocumentCommittedEvent)
      );
      this.logger.log({ documentId, revisionId }, 'commitDocument: done');
      return revisionId;
    });
  }

  async getAttachmentUploadUrl({
    documentId,
    filename,
    userId,
  }: {
    documentId: string;
    filename: string;
    userId: string;
  }) {
    this.logger.debug(
      { documentId, filename, userId },
      'getAttachmentUploadUrl: checking permission'
    );
    const hasPermission = await this.authorizationService.hasNotePermission({
      documentId,
      memberId: userId,
      permission: 'write',
    });
    if (!hasPermission) {
      throw new DocumentPermissionException(documentId, userId);
    }
    const key = `document-attachments/${documentId}/${uuidv7()}-${filename}`;
    const { uploadUrl, publicUrl } =
      await this.storageService.generateAttachmentPresignedUploadUrl(key);
    this.logger.log({ documentId, filename, key }, 'getAttachmentUploadUrl: generated upload URL');
    return {
      url: publicUrl,
      uploadUrl: uploadUrl,
    };
  }

  async updateDataById(id: string, data: Buffer): Promise<void> {
    this.logger.debug({ id, dataSize: data.length }, 'updateDataById: updating document data');
    await this.repo.update(id, { data, modified: true });
    this.logger.log({ id, dataSize: data.length }, 'updateDataById: document data updated');
  }

  async getById(id: string): Promise<DocumentEntity | null> {
    this.logger.debug({ id }, 'getById: fetching document');
    const document = await this.repo.findOneBy({ id });
    this.logger.log({ id }, 'getById: document fetched');
    return document;
  }
}
