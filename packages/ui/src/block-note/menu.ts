import { MyEditor, filterSuggestionItems } from '@blocknote/core';
import { DefaultReactSuggestionItem } from '@blocknote/react';
import type { NoteNote } from '@notopia-uit/api-gen';
import type { Meilisearch } from 'meilisearch';

const getLocalDocumentTags = (editor: MyEditor): string[] => {
  const tags = new Set<string>();

  editor.forEachBlock((block) => {
    if (Array.isArray(block.content)) {
      for (const inlineNode of block.content) {
        if (inlineNode.type === 'tag') {
          tags.add(inlineNode.props.tag);
        }
      }
    }

    return true;
  });

  return Array.from(tags);
};

interface SearchResult {
  id: string;
  name: string;
}

export const searchNotesFromMeilisearch = async (
  client: Meilisearch | null,
  query: string
): Promise<SearchResult[]> => {
  if (!client || !query) {
    return [];
  }

  try {
    const index = client.index('notes');
    const results = await index.search(query, {
      limit: 10,
    });
    return results.hits as SearchResult[];
  } catch (error) {
    console.error('Error searching notes from Meilisearch:', error);
    return [];
  }
};

export const searchTagsFromMeilisearch = async (
  client: Meilisearch | null,
  query: string
): Promise<string[]> => {
  if (!client || !query) {
    return [];
  }

  try {
    const index = client.index('tags');
    const results = await index.search(query, {
      limit: 10,
    });
    return results.hits.map((hit: any) => hit.name || hit.tag || '');
  } catch (error) {
    console.error('Error searching tags from Meilisearch:', error);
    return [];
  }
};

export const getNoteMenuItems = (
  editor: MyEditor,
  query: string,
  notes: SearchResult[]
): DefaultReactSuggestionItem[] => {
  const items: DefaultReactSuggestionItem[] = notes.map((note) => ({
    title: note.name,
    subtext: `ID: ${note.id}`,
    onItemClick: () => {
      editor.insertInlineContent([{ type: 'reference', props: { noteId: note.id } }, ' ']);
    },
  }));
  return filterSuggestionItems(items, query);
};

export const getTagMenuItems = (editor: MyEditor, query: string, tags: string[]) => {
  const localTags = getLocalDocumentTags(editor);

  const combinedTags = Array.from(new Set([...localTags, ...tags]));

  const items: DefaultReactSuggestionItem[] = combinedTags.map((tag) => ({
    title: tag,
    onItemClick: () => {
      editor.insertInlineContent([{ type: 'tag', props: { tag } }, ' ']);
    },
  }));

  const filteredItems = filterSuggestionItems(items, query);

  if (query && !filteredItems.some((item) => item.title.toLowerCase() === query.toLowerCase())) {
    filteredItems.unshift({
      title: `Create new tag: #${query}`,
      onItemClick: () => {
        editor.insertInlineContent([{ type: 'tag', props: { tag: query } }, ' ']);
      },
    });
  }

  return filteredItems;
};
