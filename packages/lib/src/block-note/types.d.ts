import * as blocknote from '@blocknote/core';

import { BlockNoteReferenceInlineContentSpec } from './reference';
import { BlockNoteTagInlineContentSpec } from './tag';

type MyInlineContentSpecs = blocknote.InlineContentSpecs & {
  tag: BlockNoteTagInlineContentSpec;
  reference: BlockNoteReferenceInlineContentSpec;
};

declare const myInlineContentSpecs: MyInlineContentSpecs;

const customSchema = new blocknote.CustomBlockNoteSchema({
  blockSpecs: blocknote.defaultBlockSpecs,
  inlineContentSpecs: myInlineContentSpecs,
  styleSpecs: blocknote.defaultStyleSpecs,
});

declare module '@blocknote/core' {
  export type BlockNoteSchema = typeof customSchema;
  export type DefaultStyleSchema = typeof customSchema.styleSchema;
  export type DefaultBlockSchema = typeof customSchema.blockSchema;
  export type DefaultInlineContentSchema = typeof customSchema.inlineContentSchema;

  export type Block = typeof customSchema.Block;
  export type BlockNoteEditor = typeof customSchema.BlockNoteEditor;
}
