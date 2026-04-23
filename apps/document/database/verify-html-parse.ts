import { type MySchema, type Block } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { createServerBlockNoteSchema } from '@notopia-uit/lib/server';
import { marked } from 'marked';
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { parseSeedMarkdownToBlocks } from './seeds/blocknote-seed-transform';

type InlineNode = {
  type?: string;
  [key: string]: unknown;
};

type MutableBlock = Block & {
  content?: unknown;
  children?: unknown;
};

function isInlineNodeArray(value: unknown): value is InlineNode[] {
  return Array.isArray(value);
}

function isReferenceInline(node: InlineNode): boolean {
  return node.type === 'reference';
}

function isTagInline(node: InlineNode): boolean {
  return node.type === 'tag';
}

function countInlineTypes(blocks: Block[]): { references: number; tags: number } {
  let references = 0;
  let tags = 0;

  const walk = (block: Block): void => {
    const current = block as MutableBlock;

    if (isInlineNodeArray(current.content)) {
      for (const inline of current.content) {
        if (isReferenceInline(inline)) {
          references += 1;
        }
        if (isTagInline(inline)) {
          tags += 1;
        }
      }
    }

    if (Array.isArray(current.children)) {
      for (const child of current.children) {
        walk(child as Block);
      }
    }
  };

  for (const block of blocks) {
    walk(block);
  }

  return { references, tags };
}

async function main(): Promise<void> {
  const __dirname = path.dirname(fileURLToPath(import.meta.url));
  const blockNoteSchema: MySchema = createServerBlockNoteSchema();
  const editor = ServerBlockNoteEditor.create({ schema: blockNoteSchema });

  const target =
    process.argv[2] ??
    path.join(
      __dirname,
      'seed-data',
      '000c2575-1847-5293-8117-2415a2328ef8.md'
    );

  const markdown = readFileSync(target, 'utf-8');
  const html = marked.parse(markdown, { async: false, gfm: true });
  console.log(`htmlLength=${typeof html === 'string' ? html.length : 0}`);
  if (typeof html === 'string') {
    console.log(html.slice(0, 300));
  }
  const blocks = (await parseSeedMarkdownToBlocks(editor, markdown)) as Block[];
  const counts = countInlineTypes(blocks);

  console.log(`file=${target}`);
  console.log(`blocks=${blocks.length}`);
  console.log(`referenceNodes=${counts.references}`);
  console.log(`tagNodes=${counts.tags}`);
}

void main();
