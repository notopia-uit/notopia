import '@blocknote/core';
import { Block, InlineContent } from '@blocknote/core';

import { createServerBlockNoteSchema } from './block-note';

declare module '@blocknote/core' {
  export type MySchema = ReturnType<typeof createServerBlockNoteSchema>;
  export type DefaultStyleSchema = MySchema['styleSchema'];
  export type DefaultBlockSchema = MySchema['blockSchema'];
  export type DefaultInlineContentSchema = MySchema['inlineContentSchema'];

  export type MyBlock = Block<
    DefaultBlockSchema,
    DefaultInlineContentSchema,
    DefaultStyleSchema
  >;
  export type MyEditor = MySchema['BlockNoteEditor'];
  export type InlineNode = InlineContent<InlineSchema, StyleSchema>;
}
