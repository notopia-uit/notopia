import { BlockNoteSchema } from '@blocknote/core';

import { createReferenceSpec } from './reference';
import { createTagSpec } from './tag';

export function createSchema() {
  return BlockNoteSchema.create().extend({
    inlineContentSpecs: {
      reference: createReferenceSpec(),
      tag: createTagSpec(),
    },
  });
}
