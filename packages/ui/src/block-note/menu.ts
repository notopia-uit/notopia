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

// Private Use Area chars (U+E000/U+E001), used as Meilisearch highlight tags.
// They are extremely unlikely to appear in note text, so splitting is unambiguous.
export const HIGHLIGHT_PRE_TAG = '\uE000';
export const HIGHLIGHT_POST_TAG = '\uE001';

export interface HighlightSegment {
  text: string;
  highlighted: boolean;
}

export const parseHighlight = (formatted: string): HighlightSegment[] => {
  const segments: HighlightSegment[] = [];
  let rest = formatted;

  while (rest.length > 0) {
    const start = rest.indexOf(HIGHLIGHT_PRE_TAG);
    if (start === -1) {
      segments.push({ text: rest, highlighted: false });
      break;
    }

    if (start > 0) {
      segments.push({ text: rest.slice(0, start), highlighted: false });
    }

    const end = rest.indexOf(HIGHLIGHT_POST_TAG, start + HIGHLIGHT_PRE_TAG.length);
    if (end === -1) {
      segments.push({ text: rest.slice(start + HIGHLIGHT_PRE_TAG.length), highlighted: true });
      break;
    }

    segments.push({
      text: rest.slice(start + HIGHLIGHT_PRE_TAG.length, end),
      highlighted: true,
    });
    rest = rest.slice(end + HIGHLIGHT_POST_TAG.length);
  }

  return segments;
};

export interface SearchResult {
  id: string;
  name: string;
  formattedName?: string;
  contentSnippet?: string;
}

export const searchNotesFromMeilisearch = async (
  client: Meilisearch,
  query: string
): Promise<SearchResult[]> => {
  try {
    const index = client.index<ShareNoteSearch>('notes');
    const results = await index.search(query, {
      limit: 10,
      attributesToHighlight: ['name', 'plainTextContent'],
      attributesToCrop: ['plainTextContent'],
      cropLength: 30,
      cropMarker: '…',
      highlightPreTag: HIGHLIGHT_PRE_TAG,
      highlightPostTag: HIGHLIGHT_POST_TAG,
    });
    return results.hits.map((hit) => ({
      id: hit.id,
      name: hit.name,
      formattedName: hit._formatted?.name,
      contentSnippet: hit._formatted?.plainTextContent,
    }));
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

export type NoteSuggestionItem = DefaultReactSuggestionItem & {
  formattedName?: string;
  contentSnippet?: string;
};

export const getNoteMenuItems = (
  editor: MyEditor,
  query: string,
  notes: SearchResult[]
): NoteSuggestionItem[] => {
  const items: NoteSuggestionItem[] = notes.map((note) => ({
    title: note.name,
    formattedName: note.formattedName,
    contentSnippet: note.contentSnippet,
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
