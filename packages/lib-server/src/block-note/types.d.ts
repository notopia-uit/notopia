import '@blocknote/core';
import { BlockNoteEditorOptions } from '@blocknote/core';

import { createSchema } from './block-note';

declare module '@blocknote/core' {
  export type MySchema = ReturnType<typeof createSchema>;
  export type DefaultStyleSchema = MySchema['styleSchema'];
  export type DefaultBlockSchema = MySchema['blockSchema'];
  export type DefaultInlineContentSchema = MySchema['inlineContentSchema'];

  export type MyBlock = Block<DefaultBlockSchema, DefaultInlineContentSchema, DefaultStyleSchema>;
  export type MyEditor = MySchema['BlockNoteEditor'];
  export interface MyEditorOptions<
    BSchema extends BlockSchema = MySchema['blockSchema'],
    ISchema extends InlineContentSchema = MySchema['inlineContentSchema'],
    SSchema extends StyleSchema = MySchema['styleSchema'],
  > extends BlockNoteEditorOptions<BSchema, ISchema, SSchema> {}
}

import '@blocknote/server-util';

declare module '@blocknote/server-util' {
  type MySchema = ReturnType<typeof createSchema>;

  export interface ServerBlockNoteEditor<
    BSchema extends BlockSchema = MySchema['blockSchema'],
    ISchema extends InlineContentSchema = MySchema['inlineContentSchema'],
    SSchema extends StyleSchema = MySchema['styleSchema'],
  > {}
}
