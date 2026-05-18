import {
  CustomInlineContentConfig,
  CustomInlineContentFromConfig,
  type InlineContentSpec,
} from '@blocknote/core';

export const TagConfig = {
  type: 'tag',
  propSchema: {
    tag: { default: '' },
  },
  content: 'none',
} as const satisfies CustomInlineContentConfig;

export type TagInlineContentSpec = InlineContentSpec<typeof TagConfig>;

export type TagInline = CustomInlineContentFromConfig<typeof TagConfig, any>;
