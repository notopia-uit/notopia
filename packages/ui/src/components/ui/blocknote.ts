// TODO: reconsider rename those block note file to block-note
import {
  createBlockNoteReferenceSpec,
  getNoteNameFn,
} from './blocknote-reference';
import { blockNoteTagSpec } from './blocknote-tag';
import { BlockNoteSchema as OriginalBlockNoteSchema } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';

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

export type BlockNoteSchema = ReturnType<typeof createBlockNoteSchema>;

export type Block = BlockNoteSchema['Block'];

export type BlockNoteEditor = BlockNoteSchema['BlockNoteEditor'];

export function createBlockNoteServerEditor(schema: BlockNoteSchema) {
  return ServerBlockNoteEditor.create({ schema });
}

export type BlockNoteServerEditor = ReturnType<
  typeof createBlockNoteServerEditor
>;
