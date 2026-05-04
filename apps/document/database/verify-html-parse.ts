import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { type MyBlock, type MySchema } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { createServerBlockNoteSchema } from '@notopia-uit/lib/server';
import { marked } from 'marked';

import { parseSeedMarkdownToBlocks } from './seeds/blocknote-seed-transform';

function countInlineTypes(blocks: MyBlock[]): {
  references: number;
  tags: number;
} {
  let references = 0;
  let tags = 0;

  const walk = (block: MyBlock): void => {
    if (Array.isArray(block.content)) {
      for (const inline of block.content) {
        switch (inline.type) {
          case 'reference':
            references += 1;
            break;
          case 'tag':
            tags += 1;
            break;
        }
      }
    }

    for (const child of block.children) {
      walk(child as MyBlock);
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
    process.argv[2] ?? path.join(__dirname, 'seed-data', '000c2575-1847-5293-8117-2415a2328ef8.md');

  const markdown = readFileSync(target, 'utf-8');
  const html = marked.parse(markdown, { async: false, gfm: true });
  console.log(`htmlLength=${typeof html === 'string' ? html.length : 0}`);
  if (typeof html === 'string') {
    console.log(html.slice(0, 300));
  }
  const blocks = await parseSeedMarkdownToBlocks(editor, markdown);
  const counts = countInlineTypes(blocks);

  console.log(`file=${target}`);
  console.log(`blocks=${blocks.length}`);
  console.log(`referenceNodes=${counts.references}`);
  console.log(`tagNodes=${counts.tags}`);
}

void main();
