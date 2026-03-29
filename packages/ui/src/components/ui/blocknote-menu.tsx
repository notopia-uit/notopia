import { schema } from './blocknote-editor';
import { filterSuggestionItems } from '@blocknote/core';
import { DefaultReactSuggestionItem } from '@blocknote/react';
import type { NoteNote } from '@notopia-uit/api-gen';

// TODO: handle the note and tag fetch, this is triggered when the menu open only, not on query change

const getLocalDocumentTags = (
  editor: typeof schema.BlockNoteEditor
): string[] => {
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

export const getNoteMenuItems = (
  editor: typeof schema.BlockNoteEditor,
  query: string,
  notes: NoteNote[]
): DefaultReactSuggestionItem[] => {
  const items: DefaultReactSuggestionItem[] = notes.map((note) => ({
    title: note.name,
    subtext: `ID: ${note.id}`,
    onItemClick: () => {
      editor.insertInlineContent([
        { type: 'reference', props: { noteId: note.id } },
        ' ',
      ]);
    },
  }));
  return filterSuggestionItems(items, query);
};

export const getTagMenuItems = async (
  editor: typeof schema.BlockNoteEditor,
  query: string,
  tags: string[]
) => {
  const localTags = getLocalDocumentTags(editor);

  const combinedTags = Array.from(new Set([...localTags, ...tags]));

  const items: DefaultReactSuggestionItem[] = combinedTags.map((tag) => ({
    title: tag,
    onItemClick: () => {
      editor.insertInlineContent([{ type: 'tag', props: { tag } }, ' ']);
    },
  }));

  const filteredItems = filterSuggestionItems(items, query);

  if (
    query &&
    !filteredItems.some(
      (item) => item.title.toLowerCase() === query.toLowerCase()
    )
  ) {
    filteredItems.unshift({
      title: `Create new tag: #${query}`,
      onItemClick: () => {
        editor.insertInlineContent([
          { type: 'tag', props: { tag: query } },
          ' ',
        ]);
      },
    });
  }

  return filteredItems;
};
