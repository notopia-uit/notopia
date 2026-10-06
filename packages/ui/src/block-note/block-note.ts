import { BlockNoteSchema as OriginalBlockNoteSchema } from '@blocknote/core';

import { createBlockNoteReferenceSpec } from './reference';
import { createBlockNoteTagSpec } from './tag';

export function createBlockNoteSchema(options: { apiUrl: string; aiApiUrl: string }) {
  return OriginalBlockNoteSchema.create().extend({
    inlineContentSpecs: {
      reference: createBlockNoteReferenceSpec(options),
      tag: createBlockNoteTagSpec(),
    },
  });
}
