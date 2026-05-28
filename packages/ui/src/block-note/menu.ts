import { MyEditor, filterSuggestionItems } from '@blocknote/core';
import { DefaultReactSuggestionItem } from '@blocknote/react';
import { ShareNoteSearch } from '@notopia-uit/api-gen';
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

export interface SearchResult {
  id: string;
  name: string;
}

export const searchNotesFromMeilisearch = async (
  client: Meilisearch,
  query: string
): Promise<SearchResult[]> => {
  try {
    const index = client.index<ShareNoteSearch>('notes');
    const results = await index.search(query, {
      limit: 10,
    });
    return results.hits;
  } catch (error) {
    console.error('Error searching notes from Meilisearch:', error);
    return [];
  }
};

export const searchTagsFromMeilisearch = async (
  client: Meilisearch,
  query: string
): Promise<string[]> => {
  try {
    const index = client.index('notes');
    const results = await index.searchForFacetValues({
      facetName: 'tags',
      facetQuery: query,
      limit: 10,
    });
    return results.facetHits.map((hit) => hit.value);
  } catch (error) {
    console.error('Error searching tags from Meilisearch:', error);
    return [];
  }
};

export const searchNotesByTag = async (
  client: Meilisearch,
  tag: string
): Promise<SearchResult[]> => {
  try {
    const index = client.index<ShareNoteSearch>('notes');
    const results = await index.search('', {
      filter: [`tags = "${tag}"`],
      limit: 50,
    });
    return results.hits;
  } catch (error) {
    console.error('Error searching notes by tag from Meilisearch:', error);
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
