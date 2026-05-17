import { readFileSync } from 'node:fs';
import path from 'node:path';

import { type MyBlock } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { createServerBlockNoteSchema } from '@notopia-uit/lib/server';

import { parseSeedMarkdownToBlocks } from './blocknote-seed-transform';

function countInlineTypes(blocks: MyBlock[]): {
  references: number;
  tags: number;
} {
  let references = 0;
  let tags = 0;

  const walk = (block: MyBlock): void => {
    if (Array.isArray(block.content)) {
      for (const inline of block.content) {
        if (inline.type === 'reference') {
          references += 1;
        } else if (inline.type === 'tag') {
          tags += 1;
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

describe('seed markdown parsing', () => {
  let editor: ServerBlockNoteEditor;

  beforeAll(() => {
    const schema = createServerBlockNoteSchema();
    editor = ServerBlockNoteEditor.create({ schema });
  });

  it.each<[string, string, number, number]>([
    ['000c2575-1847-5293-8117-2415a2328ef8.md', 'PowerUp.ps1', 4, 0],
    ['0093caa1-295e-5100-b082-4a562d3f1f2c.md', 'Directory Busting', 8, 0],
    ['00d2e172-1642-52d7-adf3-a40ff4587f66.md', 'LDAP', 11, 0],
  ])('parses %s (%s)', async (file, _name, expectedRefs, expectedTags) => {
    const filePath = path.join(__dirname, '..', 'seed-data', file);
    if (!filePath) {
      test.skip(`File not found: ${file}`);
    }
    const markdown = readFileSync(filePath, 'utf-8');
    const blocks = await parseSeedMarkdownToBlocks(editor, markdown);
    const counts = countInlineTypes(blocks);

    expect(blocks.length).toBeGreaterThan(0);
    expect(counts.references).toBe(expectedRefs);
    expect(counts.tags).toBe(expectedTags);
  });
});
