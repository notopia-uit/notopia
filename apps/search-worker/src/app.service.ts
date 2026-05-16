import { MyBlock, type MySchema } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Inject, Injectable, Logger } from '@nestjs/common';
import { Meilisearch, MeilisearchError } from 'meilisearch';
import { NoteSearch } from 'model';
import { BLOCKNOTE_SCHEMA } from 'token';

import { MeiliError } from './errors';

// TODO: Handle retry, reject, dedup? idempotent, log it out
// Handle the meilisearch setting
// Or just settup the consumer retries alerady in kafka js config

//  const index = this.meili.index('notes');
//  await index.updateSettings({
//    primaryKey: 'id',
//    searchableAttributes: ['name', 'plainTextContent', 'tags'],
//    filterableAttributes: ['tags'],
//    sortableAttributes: ['createdAt'], // Missing field!
//  });

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
    @Inject(BLOCKNOTE_SCHEMA) private readonly blocknoteSchema: MySchema
  ) {}

  async handleNoteCreated(params: HandleNoteCreatedParams) {
    this.logger.debug({ id: params.id, workspaceId: params.workspaceId }, 'handleNoteCreated: indexing');
    const index = this.meili.index(AppService.noteIndex);
    const noteSearch: NoteSearch = {
      id: params.id,
      name: params.name,
      workspaceId: params.workspaceId,
    };
    try {
      await index.updateDocuments([noteSearch]);
      this.logger.log({ id: params.id }, 'handleNoteCreated: indexed');
    } catch (e) {
      if (e instanceof MeilisearchError) {
        throw new MeiliError(e);
      }
      throw e;
    }
  }

  async handleNoteUpdated(params: HandleNoteUpdatedParams) {
    this.logger.debug({ id: params.id }, 'handleNoteUpdated: indexing');
    const index = this.meili.index(AppService.noteIndex);
    const noteSearch: NoteSearch = {
      id: params.id,
      name: params.name,
      folderId: params.folderId,
      folderName: params.folderName,
      trashed: params.trashed,
    };
    try {
      await index.updateDocuments([noteSearch]);
      this.logger.log({ id: params.id }, 'handleNoteUpdated: indexed');
    } catch (e) {
      if (e instanceof MeilisearchError) {
        throw new MeiliError(e);
      }
      throw e;
    }
  }

  async handleDocumentCommitted(params: HandleDocumentCommittedParams) {
    this.logger.debug(
      { id: params.id, tags: params.tags },
      'handleDocumentCommitted: extracting content'
    );
    const index = this.meili.index(AppService.noteIndex);
    const editor = ServerBlockNoteEditor.create({
      schema: this.blocknoteSchema,
    });
    const plainTextContent = await editor.blocksToMarkdownLossy(params.content);
    const noteSearch: NoteSearch = {
      id: params.id,
      tags: params.tags,
      plainTextContent,
    };
    try {
      await index.updateDocuments([noteSearch]);
      this.logger.log({ id: params.id }, 'handleDocumentCommitted: indexed');
    } catch (e) {
      if (e instanceof MeilisearchError) {
        throw new MeiliError(e);
      }
      throw e;
    }
  }
}
