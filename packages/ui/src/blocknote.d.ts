import type { schema } from './components/ui/blocknote-editor';
import '@blocknote/core';

declare module '@blocknote/core' {
  type DefaultStyleSchema = typeof schema.styleSchema;
  type DefaultBlockSchema = typeof schema.blockSchema;
  type DefaultInlineContentSchema = typeof schema.inlineContentSchema;
}
