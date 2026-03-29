import { BlockNoteSchema as OriginalBlockNoteSchema } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import {
  blockNoteTagSpec,
  createBlockNoteReferenceSpec,
} from '@notopia-uit/ui';

export const BLOCKNOTE_SCHEMA = Symbol('BLOCKNOTE_SCHEMA');

export function BlockNoteSchemaProvider() {
  return OriginalBlockNoteSchema.create().extend({
    inlineContentSpecs: {
      reference: createBlockNoteReferenceSpec(),
      tag: blockNoteTagSpec,
    },
  });
}

export type BlockNoteSchema = ReturnType<typeof BlockNoteSchemaProvider>;

export type Block = BlockNoteSchema['Block'];

export function BlockNoteEditorProvider() {
  return ServerBlockNoteEditor.create({
    schema: BlockNoteSchemaProvider(),
  });
}

export type BlockNoteEditor = ReturnType<typeof BlockNoteEditorProvider>;
