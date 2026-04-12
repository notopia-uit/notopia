import {
  CustomInlineContentConfig,
  type InlineContentSpec,
} from '@blocknote/core';

export const BlockNoteTagConfig = {
  type: 'tag',
  propSchema: {
    tag: { default: '' },
  },
  content: 'none',
} as const satisfies CustomInlineContentConfig;

export type BlockNoteTagInlineContentSpec = InlineContentSpec<
  typeof BlockNoteTagConfig
>;
