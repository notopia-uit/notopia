import { Injectable } from '@nestjs/common';
import { ShareNoteSearch } from '@notopia-uit/api-gen';
import { MeiliSearch } from 'meilisearch';

type IndexNote = {
  id: string;
  name?: string;
  content?: string;
  tags?: string[];
};

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
    await index.addDocuments([noteSearch]);
  }

  private async blockNoteToMarkdown(_: any): Promise<string> {
    return Promise.resolve('Not implemented');
  }
}
