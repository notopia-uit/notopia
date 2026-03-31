import { Injectable } from '@nestjs/common';
import { ShareNoteSearch } from '@notopia-uit/api-gen';
import { MeiliSearch, MeiliSearchError } from 'meilisearch';

type IndexNote = {
  id: string;
  name?: string;
  content?: string;
  tags?: string[];
};

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

@Injectable()
export class AppService {
  constructor(private readonly meili: MeiliSearch) {}

  async indexNote(note: IndexNote) {
    const index = this.meili.index('notes');
    const noteSearch: Partial<ShareNoteSearch> = {
      id: note.id,
      name: note.name,
      tags: note.tags,
      plainTextContent: note.content
        ? await this.blockNoteToMarkdown(note.content)
        : undefined,
    };
    try {
      await index.addDocuments([noteSearch]);
    } catch (e) {
      if (e instanceof MeiliSearchError) {
        // FIXME: ????? What, retry? not process anymore? how many time
        throw new Error(`Failed to index note ${note.id}: ${e.message}`);
      }
    }
  }

  private async blockNoteToMarkdown(_: any): Promise<string> {
    return Promise.resolve('Not implemented');
  }
}
