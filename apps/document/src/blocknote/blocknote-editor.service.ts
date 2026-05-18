import type { MyBlock, MyEditor, MySchema, MyEditorOptions } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Doc as YDoc, applyUpdate } from 'yjs';

import { DocumentEntity } from '../document/document.entity';

export type BlocknoteEditorServiceOptions = {
  schema: MySchema;
  initialContent?: MyBlock[] | Buffer;
  editorOptions?: Omit<MyEditorOptions, 'initialContent'>;
};

export class BlocknoteEditorService {
  private readonly serverEditor: ServerBlockNoteEditor;
  private readonly editor: MyEditor;

  constructor(options: BlocknoteEditorServiceOptions) {
    this.serverEditor = ServerBlockNoteEditor.create({
      schema: options.schema,
      ...options.editorOptions,
    });
    if (options.initialContent) {
      if (Buffer.isBuffer(options.initialContent)) {
        const blocks = this.bufferToBlockNote(options.initialContent);
        this.replaceContent(blocks);
      } else {
        this.replaceContent(options.initialContent);
      }
    }
    this.editor = this.serverEditor.editor;
  }

  blocks(): MyBlock[] {
    return this.editor.document;
  }

  public static toYDoc(entity: DocumentEntity): YDoc {
    const doc = new YDoc();
    applyUpdate(doc, new Uint8Array(entity.data));
    return doc;
  }

  // Idk, this doesn't mutate the inner state of the editor?
  // it just require the editor to transform the data
  bufferToBlockNote(data: Buffer): MyBlock[] {
    const yDoc = new YDoc();
    applyUpdate(yDoc, new Uint8Array(data));
    return this.serverEditor.yDocToBlocks(yDoc);
  }

  replaceContent(blocks: MyBlock[]): void {
    this.editor.replaceBlocks(this.editor.document, blocks);
  }

  extractTagsAndOutgoingLinkIds(): {
    tags: string[];
    outgoingLinkIds: string[];
  } {
    const tags = new Set<string>();
    const linkIds = new Set<string>();

    this.editor.forEachBlock((block) => {
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
