import type { MySchema } from '@blocknote/core';
import { Inject, Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { ClientKafka } from '@nestjs/microservices';
import { InjectDataSource, InjectRepository } from '@nestjs/typeorm';
import { type DocumentCommittedEvent } from '@notopia-uit/api-share-gen';
import { Traceable } from 'nestjs-otel';
import { lastValueFrom } from 'rxjs';
import { DataSource, Repository } from 'typeorm';
import { v7 as uuidv7 } from 'uuid';

import { IDENTITY_CONFIG, IdentityConfig } from '#/config';

import { AuthorizationService } from '../authorization/authorization.service';
import { BLOCKNOTE_SCHEMA } from '../blocknote';
import { BlocknoteEditorService } from '../blocknote/blocknote-editor.service';
import { KAFKA_CLIENT } from '../kafka/token';
import { RevisionEntity } from '../revision/revision.entity';
import { StorageService } from '../storage/storage.service';
import { contentHash, type RevisionTrigger, selectAuthor, shouldCommit } from './commit-policy';
import { DocumentNotFoundException } from './document-not-found.exception';
import { DocumentPermissionException } from './document-permission.exception';
import { DocumentEntity } from './document.entity';

export type CommitResult = {
  revisionId: string | null;
  modifiedCleared: boolean;
};

@Injectable()
@Traceable()
export class DocumentService {
  private readonly logger = new Logger(DocumentService.name);
  private readonly identityConfig: IdentityConfig;

  constructor(
    @InjectRepository(DocumentEntity)
    private readonly repo: Repository<DocumentEntity>,
    private readonly storageService: StorageService,
    @InjectDataSource() private readonly dataSource: DataSource,
    private readonly authorizationService: AuthorizationService,
    @Inject(KAFKA_CLIENT) private readonly kafkaClient: ClientKafka,
    @Inject(BLOCKNOTE_SCHEMA) private readonly blocknoteSchema: MySchema,
    configService: ConfigService
  ) {
    const identityConfig = configService.get<IdentityConfig>(IDENTITY_CONFIG);
    if (!identityConfig) {
      throw new Error('IDENTITY_CONFIG not found');
    }
    this.identityConfig = identityConfig;
  }

  async commitIfDue({
    documentId,
    trigger,
    authorId,
    force,
    minIntervalSeconds,
  }: {
    documentId: string;
    trigger: RevisionTrigger;
    authorId?: string;
    force: boolean;
    minIntervalSeconds: number;
  }): Promise<CommitResult> {
    return this.dataSource.transaction(async (manager) => {
      const document = await manager.findOne(DocumentEntity, {
        where: { id: documentId },
        lock: { mode: 'pessimistic_write' },
      });
      if (!document) {
        throw new DocumentNotFoundException(documentId);
      }
      const now = new Date();
      if (
        !shouldCommit({
          modified: document.modified,
          lastCommittedAt: document.lastCommittedAt,
          now,
          force,
          minIntervalSeconds,
        })
      ) {
        this.logger.debug({ documentId, trigger }, 'commitIfDue: not due');
        return { revisionId: null, modifiedCleared: false };
      }
      const blocknoteEditorService = new BlocknoteEditorService({
        schema: this.blocknoteSchema,
        initialContent: document.data,
      });
      const blocks = blocknoteEditorService.blocks();
      const hash = contentHash(blocks);
      const latest = await manager.findOne(RevisionEntity, {
        select: { id: true, contentHash: true, createdAt: true },
        where: { document: { id: documentId } },
        order: { createdAt: 'DESC' },
      });
      if (latest?.contentHash === hash) {
        await manager.update(DocumentEntity, { id: documentId }, { modified: false });
        this.logger.log(
          { documentId, trigger, revisionId: latest.id },
          'commitIfDue: content unchanged, cleared modified'
        );
        return { revisionId: null, modifiedCleared: true };
      }
      const contributorIds = document.pendingContributorIds;
      const author = selectAuthor({
        authorId,
        contributors: contributorIds,
        serviceAccountUserId: this.identityConfig.serviceAccountUserId,
      });
      const revisionId = uuidv7();
      this.logger.debug({ documentId, revisionId, trigger }, 'commitIfDue: saving revision');
      await manager.save(RevisionEntity, {
        id: revisionId,
        document,
        content: blocks,
        authorId: author,
        trigger,
        contributorIds,
        contentHash: hash,
      });
      await manager.update(
        DocumentEntity,
        { id: documentId },
        { modified: false, lastCommittedAt: now, pendingContributorIds: [] }
      );
      // TODO: Consider refactor into a module named "EventBus", which manages event topic
      const { tags, outgoingLinkIds } = blocknoteEditorService.extractTagsAndOutgoingLinkIds();
      this.logger.debug({ documentId, tags, outgoingLinkIds }, 'commitIfDue: emitting event');
      await lastValueFrom(
        this.kafkaClient.emit('events.integration.document.document.committed', {
          id: documentId,
          userId: author,
          tags,
          outgoingLinkIds,
          content: blocks,
        } satisfies DocumentCommittedEvent)
      );
      this.logger.log({ documentId, revisionId, trigger }, 'commitIfDue: done');
      return { revisionId, modifiedCleared: true };
    });
  }

  async getLatestRevisionId(documentId: string): Promise<string | null> {
    const latest = await this.dataSource.getRepository(RevisionEntity).findOne({
      select: { id: true, createdAt: true },
      where: { document: { id: documentId } },
      order: { createdAt: 'DESC' },
    });
    return latest?.id ?? null;
  }

  async mergePendingContributors(id: string, contributors: string[]): Promise<boolean> {
    // Merged in SQL so concurrent instances never drop each other's ids; leaves updatedAt alone
    // because the sweeper's idle check follows content stores only.
    const [, affected] = (await this.repo.query(
      `UPDATE documents SET
         "pendingContributorIds" = ARRAY(
             SELECT x FROM unnest("pendingContributorIds") AS x
             WHERE x <> ALL($2::text[])
           ) || $2::text[]
       WHERE id = $1`,
      [id, contributors]
    )) as [unknown[], number];
    this.logger.debug({ id, contributors, affected }, 'mergePendingContributors: merged');
    return affected > 0;
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

  async upsertData(id: string, data: Buffer, modified: boolean): Promise<void> {
    this.logger.debug({ id, dataSize: data.length, modified }, 'upsertData: saving');
    await this.repo.upsert({ id, data, modified }, ['id']);
    this.logger.log({ id, dataSize: data.length }, 'upsertData: saved');
  }

  async deleteById(id: string): Promise<void> {
    await this.repo.delete({ id });
    this.logger.log({ id }, 'deleteById: document deleted');
  }

  async getById(id: string): Promise<DocumentEntity | null> {
    this.logger.debug({ id }, 'getById: fetching document');
    const document = await this.repo.findOneBy({ id });
    this.logger.log({ id }, 'getById: document fetched');
    return document;
  }
}
