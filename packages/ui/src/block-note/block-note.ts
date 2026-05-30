import { BlockNoteSchema as OriginalBlockNoteSchema } from '@blocknote/core';

import { createBlockNoteReferenceSpec } from './reference';
import { createBlockNoteTagSpec } from './tag';

export function createBlockNoteSchema(apiUrl?: string) {
  return OriginalBlockNoteSchema.create().extend({
    inlineContentSpecs: {
      reference: createBlockNoteReferenceSpec(apiUrl),
      tag: createBlockNoteTagSpec(),
    },
  });
}
