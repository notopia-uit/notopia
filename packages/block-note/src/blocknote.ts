// TODO: reconsider rename those block note file to block-note
import {
  createBlockNoteReferenceSpec,
  getNoteNameFn,
} from './blocknote-reference';
import { blockNoteTagSpec } from './blocknote-tag';
import { BlockNoteSchema as OriginalBlockNoteSchema } from '@blocknote/core';

export function createBlockNoteSchema({
  baseUrl,
  getNoteName,
}: {
  baseUrl: string;
  getNoteName: getNoteNameFn;
}) {
  return OriginalBlockNoteSchema.create().extend({
    inlineContentSpecs: {
      reference: createBlockNoteReferenceSpec({
        baseUrl,
        getNoteName,
      }),
      tag: blockNoteTagSpec,
    },
  });
}
