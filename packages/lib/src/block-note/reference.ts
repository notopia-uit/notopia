import {
  CustomInlineContentConfig,
  CustomInlineContentFromConfig,
  type InlineContentSpec,
} from '@blocknote/core';

export const BlockNoteReferenceConfig = {
  type: 'reference',
  propSchema: {
    noteId: { default: 'unknown' },
  },
  content: 'none',
} as const satisfies CustomInlineContentConfig;

export type BlockNoteReferenceInlineContentSpec = InlineContentSpec<
  typeof BlockNoteReferenceConfig
>;

export type ReferenceInline = CustomInlineContentFromConfig<
  typeof BlockNoteReferenceConfig,
  any
>;
