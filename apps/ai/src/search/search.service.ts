import { Inject, Injectable, Logger } from '@nestjs/common';
import type { ConfigType } from '@nestjs/config';
import { Meilisearch } from 'meilisearch';

import { servicesConfig } from '#/config.factory';

export type NoteHit = {
  id: string;
  name: string;
  folderId?: string;
  folderName?: string;
  tags?: string[];
};

@Injectable()
export class SearchService {
  private static readonly noteIndex = 'notes';
  private readonly logger = new Logger(SearchService.name);
  private readonly meili: Meilisearch;

  constructor(@Inject(servicesConfig.KEY) services: ConfigType<typeof servicesConfig>) {
    this.meili = new Meilisearch({
      host: services.meiliHost,
      apiKey: services.meiliApiKey,
    });
  }

  async searchNotes({
    query,
    workspaceId,
    limit = 5,
  }: {
    query: string;
    workspaceId?: string;
    limit?: number;
  }): Promise<NoteHit[]> {
    this.logger.debug({ query, workspaceId }, 'Searching notes for tool call');
    try {
      const result = await this.meili.index(SearchService.noteIndex).search(query, {
        filter: workspaceId ? `workspaceId = "${workspaceId}"` : undefined,
        limit,
      });
      return result.hits.map((hit) => ({
        id: String(hit.id),
        name: String(hit.name ?? ''),
        folderId: hit.folderId as string | undefined,
        folderName: hit.folderName as string | undefined,
        tags: hit.tags as string[] | undefined,
      }));
    } catch (error) {
      this.logger.error({ err: error, query }, 'Failed to search notes');
      throw error;
    }
  }
}
