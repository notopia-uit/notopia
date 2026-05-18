import {
  CustomInlineContentConfig,
  CustomInlineContentFromConfig,
  type InlineContentSpec,
} from '@blocknote/core';

export const ReferenceConfig = {
  type: 'reference',
  propSchema: {
    noteId: { default: 'unknown' },
  },
  content: 'none',
} as const satisfies CustomInlineContentConfig;

export type ReferenceInlineContentSpec = InlineContentSpec<typeof ReferenceConfig>;

export type ReferenceInline = CustomInlineContentFromConfig<typeof ReferenceConfig, any>;
