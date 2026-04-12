import { BlockNoteSchema as OriginalBlockNoteSchema } from '@blocknote/core';

import { createServerBlockNoteReferenceSpec } from './reference';
import { createServerBlockNoteTagSpec } from './tag';

export function createServerBlockNoteSchema() {
  return OriginalBlockNoteSchema.create().extend({
    inlineContentSpecs: {
      reference: createServerBlockNoteReferenceSpec(),
      tag: createServerBlockNoteTagSpec(),
    },
  });
}
