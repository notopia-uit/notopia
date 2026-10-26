import { Inject, Injectable } from '@nestjs/common';
import type { ConfigType } from '@nestjs/config';
import { Meilisearch, MeilisearchApiError } from 'meilisearch';
import { z } from 'zod';

import { AiError } from '#/common/ai-error';
import { servicesConfig } from '#/config.factory';

export type NoteHit = {
  id: string;
  workspaceId: string;
  name: string;
  folderId?: string;
  folderName?: string;
  tags?: string[];
  excerpt: string;
};

@Injectable()
export class SearchService {
  private readonly meili: Meilisearch;

  constructor(@Inject(servicesConfig.KEY) services: ConfigType<typeof servicesConfig>) {
    this.meili = new Meilisearch({
      host: services.meiliHost,
      apiKey: services.meiliApiKey,
      timeout: 5000,
    });
  }

  async searchNotes({
    query,
    workspaceId,
    limit = 5,
  }: {
    query: string;
    workspaceId: string;
    limit?: number;
  }): Promise<NoteHit[]> {
    z.uuid().parse(workspaceId);
    const result = await this.meili.index('notes').search(query, {
      filter: `workspaceId = "${workspaceId}"`,
      limit: Math.min(20, Math.max(1, limit)),
      attributesToRetrieve: [
        'id',
        'workspaceId',
        'name',
        'folderId',
        'folderName',
        'tags',
        'plainTextContent',
        'trashed',
      ],
    });
    return result.hits
      .filter(
        (hit) =>
          !hit.trashed && hit.workspaceId === workspaceId && z.uuid().safeParse(hit.id).success
      )
      .map((hit) => ({
        id: String(hit.id),
        workspaceId,
        name: String(hit.name ?? '').slice(0, 1000),
        folderId: hit.folderId as string | undefined,
        folderName: hit.folderName as string | undefined,
        tags: hit.tags as string[] | undefined,
        excerpt: String(hit.plainTextContent ?? '').slice(0, 1500),
      }));
  }

  async readIndexedNoteContent(noteId: string, workspaceId: string) {
    z.uuid().parse(noteId);
    z.uuid().parse(workspaceId);
    try {
      const document = await this.meili.index('notes').getDocument(noteId, {
        fields: ['id', 'workspaceId', 'name', 'plainTextContent', 'trashed'],
      });
      if (document.workspaceId !== workspaceId || document.trashed) return null;
      const content = String(document.plainTextContent ?? '');
      return {
        noteId,
        name: document.name,
        content: content.slice(0, 20000),
        truncated: content.length > 20000,
        source: 'committed-search-index',
        mayBeStale: true,
      };
    } catch (error) {
      if (error instanceof MeilisearchApiError && error.cause?.code === 'document_not_found')
        return null;
      throw new AiError(503, 'searchUnavailable', 'Indexed note content is unavailable.');
    }
  }
}
