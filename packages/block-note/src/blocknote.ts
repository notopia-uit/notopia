// TODO: reconsider rename those block note file to block-note
import { BlockNoteSchema as OriginalBlockNoteSchema } from '@blocknote/core';

import { createBlockNoteReferenceSpec } from './blocknote-reference';
import { createBlockNoteTagSpec } from './blocknote-tag';

export function createBlockNoteSchema(type: 'client' | 'server' = 'client') {
  return OriginalBlockNoteSchema.create().extend({
    inlineContentSpecs: {
      reference: createBlockNoteReferenceSpec(type),
      tag: createBlockNoteTagSpec(),
    },
  });
}
