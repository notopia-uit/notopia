import { existsSync, mkdirSync, readdirSync, statSync } from 'fs';
import * as fs from 'fs/promises';
import * as path from 'path';

import { ShareNoteSearch } from '@notopia-uit/api-gen';
import markdownToTxt from 'markdown-to-txt';
import { v5 as uuidv5 } from 'uuid';

// Constants
const NAMESPACE = '6ba7b810-9dad-11d1-80b4-00c04fd430c8';
const WORKSPACE_ID = '00000000-0000-4000-8000-000000000110';
const SOURCE_DIR = path.join(__dirname, '../../../submodule/trshpuppy-obsidian-notes');
const OUTPUT_DIR = path.join(__dirname, 'data');
console.log(`Source Directory: ${SOURCE_DIR}`);

// Regex
const frontmatterRegex = /^---\n[^]*?\n---\n/;
const wikiLinkRegex = /\[\[([^\]]+?)\]\]/g;
const markdownLinkRegex = /\[([^\]]+)\]\(([^)]+)\)/g;
const tagRegex = /(^|\s)#([^\s#\]]+)/g;

function markdownToPlainText(markdown: string): string {
  let text = markdown.replace(frontmatterRegex, '');
  text = text.replace(wikiLinkRegex, '$1');
  text = text.replace(markdownLinkRegex, '$1');
  text = text.replace(tagRegex, '$2');

  return markdownToTxt(text);
}

function extractTags(content: string): string[] {
  const stripped = content.replace(frontmatterRegex, '');
  const matches = Array.from(stripped.matchAll(tagRegex));
  const tags = new Set<string>();
  for (const match of matches) {
    if (match[2]) tags.add(match[2]);
  }
  return Array.from(tags);
}

async function processFile(filePath: string): Promise<void> {
  const relPath = path.relative(SOURCE_DIR, filePath);
  const pathWithoutExt = relPath.replace(/\.md$/, '');
  const id = uuidv5(pathWithoutExt, NAMESPACE);

  const folderPath = path.dirname(relPath);
  const folderId = uuidv5(folderPath === '.' ? '' : folderPath, NAMESPACE);
  const folderName = folderPath === '.' ? '' : path.basename(folderPath);

  const content = await fs.readFile(filePath, 'utf-8');

  const note: ShareNoteSearch = {
    id,
    workspaceId: WORKSPACE_ID,
    folderId,
    folderName,
    name: path.basename(pathWithoutExt),
    plainTextContent: markdownToPlainText(content),
    tags: extractTags(content),
  };

  await fs.writeFile(path.join(OUTPUT_DIR, `${id}.json`), JSON.stringify(note, null, 2));
}

async function run() {
  if (!existsSync(OUTPUT_DIR)) mkdirSync(OUTPUT_DIR, { recursive: true });

  const files: string[] = [];
  const walk = (dir: string) => {
    for (const file of readdirSync(dir)) {
      const fullPath = path.join(dir, file);
      if (statSync(fullPath).isDirectory()) {
        if (!file.startsWith('.')) walk(fullPath);
      } else if (file.endsWith('.md')) {
        files.push(fullPath);
      }
    }
  };
  walk(SOURCE_DIR);

  console.log(`Processing ${files.length} files...`);
  await Promise.all(files.map(processFile));
  console.log('Transformation completed.');
}

run().catch(console.error);
