// src/editor/schema.ts
import { CreateBlocknoteReferenceSpec } from './blocknote-reference';
import { blocknoteTagSpec } from './blocknote-tag';
import { BlockNoteSchema, defaultInlineContentSpecs } from '@blocknote/core';

const referenceSpec = CreateBlocknoteReferenceSpec();

export const schema = BlockNoteSchema.create({
  inlineContentSpecs: {
    ...defaultInlineContentSpecs,
    reference: referenceSpec,
    tag: blocknoteTagSpec,
  },
});
