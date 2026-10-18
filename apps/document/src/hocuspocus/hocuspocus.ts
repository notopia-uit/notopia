import { hostname } from 'node:os';

import { Database as HocuspocusDatabase } from '@hocuspocus/extension-database';
import { Logger as HocuspocusLogger } from '@hocuspocus/extension-logger';
import { Redis as HocuspocusRedis } from '@hocuspocus/extension-redis';
import {
  type Extension,
  onAuthenticatePayload,
  onChangePayload,
  Hocuspocus as ServerHocuspocus,
} from '@hocuspocus/server';
import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { YDocMetadataMap } from '@notopia-uit/lib/yjs';

import { AuthenticationService } from '../authentication/authentication.service';
import { AuthorizationService } from '../authorization/authorization.service';
import { BillingService } from '../billing';
import {
  HOCUSPOCUS_CONFIG,
  HocuspocusConfig,
  REDIS_CONFIG,
  RedisConfig,
  SCALING_CONFIG,
  ScalingConfig,
} from '../config';
import { DocumentService } from '../document/document.service';
import { NoteService } from '../note/note.service';
import { AutoCommitExtension } from './auto-commit.extension';
import { ContributorFlushExtension } from './contributor-flush.extension';
import { ContributorTracker } from './contributor-tracker';
import { applyStoredModified, readModified } from './document-metadata';
import { HocuspocusContext } from './hocuspocus-context';

const DEFAULT_REDIS_PORT = 6379;

@Injectable()
export class Hocuspocus {
  private readonly logger = new Logger(Hocuspocus.name);
  readonly hocuspocus: ServerHocuspocus<HocuspocusContext>;
  private readonly deletedDocumentIds = new Set<string>();
  private readonly pendingStores = new Map<string, Set<Promise<void>>>();

  constructor(
    documentService: DocumentService,
    private readonly noteService: NoteService,
    private readonly authorizationService: AuthorizationService,
    private readonly authenticationService: AuthenticationService,
    private readonly contributors: ContributorTracker,
    billingService: BillingService,
    configService: ConfigService
  ) {
    const hocuspocusCfg = configService.get<HocuspocusConfig>(HOCUSPOCUS_CONFIG);
    if (!hocuspocusCfg) {
      throw new Error('HOCUSPOCUS_CONFIG not found');
    }
    const scalingCfg = configService.get<ScalingConfig>(SCALING_CONFIG);
    const redisCfg = configService.get<RedisConfig>(REDIS_CONFIG);
    if (!scalingCfg || !redisCfg) {
      throw new Error('SCALING_CONFIG or REDIS_CONFIG not found');
    }
    this.hocuspocus = new ServerHocuspocus<HocuspocusContext>({
      name: hostname(),
      debounce: hocuspocusCfg.debounceMs,
      maxDebounce: hocuspocusCfg.maxDebounceMs,
      onLoadDocument: async (data) => {
        this.logger.debug(
          { documentId: data.documentName, documentMetadata: data.document.getMap('metadata') },
          'Document loaded'
        );
        await Promise.resolve();
      },
      afterLoadDocument: async ({ documentName, document }) => {
        const entity = await documentService.getById(documentName);
        applyStoredModified(document, entity?.modified ?? false);
      },
      afterUnloadDocument: async ({ documentName }) => {
        this.deletedDocumentIds.delete(documentName);
        this.contributors.forget(documentName);
      },
      onAuthenticate: (...args) => this.onAuthenticate(...args),
      onChange: (...args) => this.onChange(...args),
      extensions: [
        new ContributorFlushExtension(contributors, documentService),
        ...this.scalingExtensions(scalingCfg, redisCfg),
        new HocuspocusLogger({
          log: (message) => this.logger.log(message),
        }),
        new HocuspocusDatabase({
          fetch: async ({ documentName: id }) => {
            this.logger.log({ documentId: id }, 'Fetching document from database');
            const document = await documentService.getById(id);
            this.logger.log({ documentId: id }, 'Fetched document from database');
            return document?.data ?? null;
          },
          store: async ({ documentName: id, state, document }) => {
            if (this.deletedDocumentIds.has(id)) return;
            await this.trackStore(
              id,
              documentService.upsertData(id, state, readModified(document))
            );
          },
        }),
        new AutoCommitExtension(documentService, noteService, billingService),
      ],
    });
  }

  // Hooks run by priority, not array order: the Redis extension (1000) runs after
  // ContributorFlushExtension (1100) and before the Database extension (100).
  private scalingExtensions(scalingCfg: ScalingConfig, redisCfg: RedisConfig): Extension[] {
    if (!scalingCfg.enabled) return [];
    const redisUrl = redisCfg.url;
    if (!redisUrl) {
      throw new Error('NOTOPIA_DOCUMENT_REDIS_URL is required when scaling is enabled');
    }
    // The extension bundles its own ioredis, so it gets connection options instead of a client.
    const url = new URL(redisUrl);
    return [
      new HocuspocusRedis({
        host: url.hostname,
        port: url.port ? Number(url.port) : DEFAULT_REDIS_PORT,
        options: {
          username: url.username ? decodeURIComponent(url.username) : undefined,
          password: url.password ? decodeURIComponent(url.password) : undefined,
          db: url.pathname.length > 1 ? Number(url.pathname.slice(1)) : undefined,
          tls: url.protocol === 'rediss:' ? {} : undefined,
        },
      }),
    ];
  }

  markDeleted(id: string): void {
    if (this.hocuspocus.documents.has(id)) {
      this.deletedDocumentIds.add(id);
    }
  }

  async waitForPendingStores(id: string): Promise<void> {
    const pending = this.pendingStores.get(id);
    if (pending) await Promise.allSettled(pending);
  }

  private async trackStore(id: string, store: Promise<void>): Promise<void> {
    let pending = this.pendingStores.get(id);
    if (!pending) {
      pending = new Set();
      this.pendingStores.set(id, pending);
    }
    pending.add(store);
    try {
      await store;
    } finally {
      pending.delete(store);
      if (pending.size === 0) this.pendingStores.delete(id);
    }
  }

  closeConnections(id: string): void {
    this.hocuspocus.closeConnections(id);
  }

  private async onAuthenticate(
    data: onAuthenticatePayload<HocuspocusContext>
  ): Promise<HocuspocusContext> {
    const documentId = data.documentName;
    this.logger.log({ documentId }, 'Authenticating user for document');
    const user = await this.authenticationService.validateToken(data.token);
    const note = await this.noteService.getNoteById({
      noteId: documentId,
      userId: user.id,
    });
    if (!note) {
      throw new Error(`Document with ID ${documentId} does not exist`);
    }
    const userPermissionsRes = await this.authorizationService.getUserDocumentPermissions(
      user.id,
      documentId
    );
    if (!userPermissionsRes.canRead) {
      throw new Error(`User ${user.id} does not have permission to access document ${documentId}`);
    }
    if (!userPermissionsRes.canWrite || note.trashed) {
      data.connectionConfig.readOnly = true;
    }
    return { user };
  }

  private async onChange(
    data: onChangePayload<HocuspocusContext>
  ): Promise<HocuspocusContext | void> {
    if (!data.connection) return Promise.resolve();

    this.contributors.add(data.documentName, data.context.user.id);

    const metadata = data.document.getMap('metadata') as YDocMetadataMap;
    const existing = metadata.get('metadata');
    if (!existing || !existing.modified) {
      metadata.set('metadata', { modified: true });
    }
  }
}
