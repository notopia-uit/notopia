import { type MyBlock, type MySchema } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Inject, Injectable } from '@nestjs/common';
import { Doc as YDoc, applyUpdate } from 'yjs';

import { DocumentEntity } from '../document/document.entity';

import { BLOCKNOTE_SCHEMA } from './token';

@Injectable()
export class BlocknoteService {
  constructor(@Inject(BLOCKNOTE_SCHEMA) private readonly blocknoteSchema: MySchema) {}

  createEditor(): ServerBlockNoteEditor {
    return ServerBlockNoteEditor.create({
      schema: this.blocknoteSchema,
    });
  }

  toYDoc(entity: DocumentEntity): YDoc {
    const doc = new YDoc();
    applyUpdate(doc, new Uint8Array(entity.data));
    return doc;
  }

  bufferToBlockNote(data: Buffer, editor: ServerBlockNoteEditor): MyBlock[] {
    const yDoc = new YDoc();
    applyUpdate(yDoc, new Uint8Array(data));
    return editor.yDocToBlocks(yDoc);
  }

  extractTagsAndOutgoingLinkIds(editor: ServerBlockNoteEditor): {
    tags: string[];
    outgoingLinkIds: string[];
  } {
    const tags = new Set<string>();
    const linkIds = new Set<string>();

    editor.editor.forEachBlock((block) => {
      if (!Array.isArray(block.content)) {
        return false;
      }
      for (const inlineNode of block.content) {
        switch (inlineNode.type) {
          case 'tag':
            tags.add(inlineNode.props.tag);
            break;
          case 'reference':
            linkIds.add(inlineNode.props.noteId);
            break;
        }
      }

      return true;
    });

    return {
      tags: Array.from(tags),
      outgoingLinkIds: Array.from(linkIds),
    };
  }
}
