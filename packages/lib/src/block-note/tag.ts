import {
  CustomInlineContentConfig,
  CustomInlineContentFromConfig,
  type InlineContentSpec,
} from '@blocknote/core';

export const BlockNoteTagConfig = {
  type: 'tag',
  propSchema: {
    tag: { default: '' },
  },
  content: 'none',
} as const satisfies CustomInlineContentConfig;

export type BlockNoteTagInlineContentSpec = InlineContentSpec<typeof BlockNoteTagConfig>;

export type TagInline = CustomInlineContentFromConfig<typeof BlockNoteTagConfig, any>;
