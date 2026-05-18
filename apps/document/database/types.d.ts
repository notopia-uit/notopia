import '@blocknote/core';

declare module '@blocknote/core' {
  export type MyInlineContent = InlineContent<InlineSchema, StyleSchema>;
}
