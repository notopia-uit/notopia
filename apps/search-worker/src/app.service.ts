import type { MyBlock, MySchema } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Inject, Injectable, Logger } from '@nestjs/common';
import { type EnqueuedTaskPromise, Meilisearch, MeilisearchError } from 'meilisearch';

import { MeiliError, MeiliTaskError } from './errors';
import { NoteSearch } from './model';
import { BLOCKNOTE_SCHEMA } from './token';
import { TombstoneStore } from './tombstone';

// TODO: Handle retry, reject, dedup? idempotent, log it out
// Handle the meilisearch setting
// Or just settup the consumer retries alerady in kafka js config

export type HandleNoteCreatedParams = Required<Pick<NoteSearch, 'id' | 'name' | 'workspaceId'>>;

export type HandleNoteUpdatedParams = Required<
  Pick<NoteSearch, 'id' | 'name' | 'folderId' | 'folderName'>
> &
  Pick<NoteSearch, 'trashed'>;

export type HandleDocumentCommittedParams = Required<Pick<NoteSearch, 'id' | 'tags'>> & {
  content: MyBlock[];
};

@Injectable()
export class AppService {
  private readonly logger = new Logger(AppService.name);
  private static readonly noteIndex = 'notes';

  constructor(
    private readonly meili: Meilisearch,
    private readonly tombstones: TombstoneStore,
    @Inject(BLOCKNOTE_SCHEMA) private readonly blocknoteSchema: MySchema
  ) {}

  async handleNoteCreated(params: HandleNoteCreatedParams) {
    this.logger.debug(
      { id: params.id, workspaceId: params.workspaceId },
      'handleNoteCreated: indexing'
    );
    if (
      await this.upsertUnlessDeleted({
        id: params.id,
        name: params.name,
        workspaceId: params.workspaceId,
      })
    ) {
      this.logger.log({ id: params.id }, 'handleNoteCreated: indexed');
    }
  }

  async handleNoteUpdated(params: HandleNoteUpdatedParams) {
    this.logger.debug({ id: params.id }, 'handleNoteUpdated: indexing');
    if (
      await this.upsertUnlessDeleted({
        id: params.id,
        name: params.name,
        folderId: params.folderId,
        folderName: params.folderName,
        trashed: params.trashed,
      })
    ) {
      this.logger.log({ id: params.id }, 'handleNoteUpdated: indexed');
    }
  }

  async handleNoteDeleted(id: string) {
    await this.tombstones.mark(id);
    await this.deleteNote(id);
    this.logger.log({ id }, 'handleNoteDeleted: removed');
  }

  async handleDocumentCommitted(params: HandleDocumentCommittedParams) {
    this.logger.debug(
      { id: params.id, tags: params.tags },
      'handleDocumentCommitted: extracting content'
    );
    const editor = ServerBlockNoteEditor.create({
      schema: this.blocknoteSchema,
    });
    const plainTextContent = await editor.blocksToMarkdownLossy(params.content);
    if (
      await this.upsertUnlessDeleted({
        id: params.id,
        tags: params.tags,
        plainTextContent,
      })
    ) {
      this.logger.log({ id: params.id }, 'handleDocumentCommitted: indexed');
    }
  }

  private async upsertUnlessDeleted(noteSearch: NoteSearch): Promise<boolean> {
    if (await this.tombstones.has(noteSearch.id)) {
      this.logger.log({ id: noteSearch.id }, 'skip indexing: note is deleted');
      return false;
    }
    await this.runTask(() => this.meili.index(AppService.noteIndex).updateDocuments([noteSearch]));
    if (await this.tombstones.has(noteSearch.id)) {
      await this.deleteNote(noteSearch.id);
      this.logger.log({ id: noteSearch.id }, 'removed note deleted during indexing');
      return false;
    }
    return true;
  }

  private async deleteNote(id: string) {
    await this.runTask(() => this.meili.index(AppService.noteIndex).deleteDocument(id));
  }

  private async runTask(enqueue: () => EnqueuedTaskPromise) {
    try {
      const task = await enqueue().waitTask();
      if (task.status !== 'succeeded') {
        throw new MeiliTaskError(task);
      }
    } catch (e) {
      if (e instanceof MeilisearchError) {
        throw new MeiliError(e);
      }
      throw e;
    }
  }
}
