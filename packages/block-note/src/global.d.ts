import '@blocknote/core';

import { createBlockNoteSchema } from './blocknote';

declare module '@blocknote/core' {
  export type MySchema = ReturnType<typeof createBlockNoteSchema>;
  export type DefaultStyleSchema = MySchema['styleSchema'];
  export type DefaultBlockSchema = MySchema['blockSchema'];
  export type DefaultInlineContentSchema = MySchema['inlineContentSchema'];

  export type MyEditor = MySchema['BlockNoteEditor'];
}
