// TODO: reconsider rename those blocknote file to block-note
import { createBlockNoteReferenceSpec } from './blocknote-reference';
import { blockNoteTagSpec } from './blocknote-tag';
import { BlockNoteSchema, defaultInlineContentSpecs } from '@blocknote/core';

const referenceSpec = createBlockNoteReferenceSpec();

export const schema = BlockNoteSchema.create({
  inlineContentSpecs: {
    ...defaultInlineContentSpecs,
    reference: referenceSpec,
    tag: blockNoteTagSpec,
  },
});
