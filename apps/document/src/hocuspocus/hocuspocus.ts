import { hostname } from 'node:os';

import { Database as HocuspocusDatabase } from '@hocuspocus/extension-database';
import { Logger as HocuspocusLogger } from '@hocuspocus/extension-logger';
import { Redis as HocuspocusRedis, type RedisInstance } from '@hocuspocus/extension-redis';
import {
  type Extension,
  onAuthenticatePayload,
  onChangePayload,
  Hocuspocus as ServerHocuspocus,
} from '@hocuspocus/server';
import { BeforeApplicationShutdown, Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { YDocMetadataMap } from '@notopia-uit/lib/yjs';
import Redis from 'ioredis';

import { AuthenticationService } from '../authentication/authentication.service';
import { AuthorizationService } from '../authorization/authorization.service';
import { BillingService } from '../billing';
import {
  AUTO_COMMIT_CONFIG,
  AutoCommitConfig,
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
import { applyStoredModified, hasModifiedMetadata, readModified } from './document-metadata';
import { HocuspocusContext } from './hocuspocus-context';

const SHUTDOWN_UNLOAD_TIMEOUT_MS = 10_000;

@Injectable()
export class Hocuspocus implements BeforeApplicationShutdown {
  private readonly logger = new Logger(Hocuspocus.name);
  readonly hocuspocus: ServerHocuspocus<HocuspocusContext>;
  private readonly deletedDocumentIds = new Set<string>();
  private readonly pendingStores = new Map<string, Set<Promise<void>>>();
  private onAllUnloaded: (() => void) | undefined;

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
    const autoCommitCfg = configService.get<AutoCommitConfig>(AUTO_COMMIT_CONFIG);
    if (!autoCommitCfg) {
      throw new Error('AUTO_COMMIT_CONFIG not found');
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
        // The Redis extension (priority 1000) runs first and may have merged a peer's in-memory
        // state, whose metadata is newer than the DB column.
        if (hasModifiedMetadata(document)) return;
        const entity = await documentService.getById(documentName);
        applyStoredModified(document, entity?.modified ?? false);
      },
      afterUnloadDocument: async ({ documentName }) => {
        this.deletedDocumentIds.delete(documentName);
        this.contributors.forget(documentName);
        if (this.hocuspocus.getDocumentsCount() === 0) this.onAllUnloaded?.();
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
        new AutoCommitExtension(
          documentService,
          noteService,
          billingService,
          autoCommitCfg.checkFloorSeconds
        ),
      ],
    });
  }

  // Hooks run by priority, not array order: the Redis extension (1000) runs after
  // ContributorFlushExtension (1100) and before the Database extension (100).
  // Its afterStoreDocument waits disconnectDelay (~1 s) for local-origin stores inside the
  // saveMutex, which delays AutoCommitExtension (50) by that much; accepted, do not reorder.
  private scalingExtensions(scalingCfg: ScalingConfig, redisCfg: RedisConfig): Extension[] {
    if (!scalingCfg.enabled) return [];
    const redisUrl = redisCfg.url;
    if (!redisUrl) {
      throw new Error('NOTOPIA_DOCUMENT_REDIS_URL is required when scaling is enabled');
    }
    // The extension depends on ioredis 5.6.1, whose URL parsing calls the deprecated url.parse();
    // clients from the app's ioredis parse the URL with WHATWG URL instead.
    return [
      new HocuspocusRedis({
        createClient: () => new Redis(redisUrl) as unknown as RedisInstance,
      }),
    ];
  }

  // Runs in beforeApplicationShutdown, not onApplicationShutdown: TypeORM closes its DataSource
  // in onApplicationShutdown, and the final stores still need it. HocuspocusGateway already
  // closed every connection in onModuleDestroy, which triggers those stores.
  async beforeApplicationShutdown(): Promise<void> {
    await this.waitForUnload();
    await Promise.allSettled([...this.pendingStores.values()].flatMap((stores) => [...stores]));
    await this.hocuspocus.hooks('onDestroy', { instance: this.hocuspocus });
  }

  private async waitForUnload(): Promise<void> {
    if (this.hocuspocus.getDocumentsCount() === 0) return;
    let timer: NodeJS.Timeout | undefined;
    const unloaded = new Promise<boolean>((resolve) => {
      this.onAllUnloaded = () => resolve(true);
      timer = setTimeout(() => resolve(false), SHUTDOWN_UNLOAD_TIMEOUT_MS);
    });
    this.hocuspocus.flushPendingStores();
    const done = await unloaded;
    clearTimeout(timer);
    this.onAllUnloaded = undefined;
    if (!done) {
      this.logger.warn(
        { documents: this.hocuspocus.getDocumentsCount() },
        'documents still loaded at shutdown'
      );
    }
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
