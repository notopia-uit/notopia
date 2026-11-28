import { mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

import { generateLlmsFiles } from 'vitepress-openapi';
import { expect, it } from 'vitest';

import { restLlmsOptions } from './rest-llms';

it('writes OpenAPI Markdown without replacing general indexes', async ({ onTestFinished }) => {
  const outDir = await mkdtemp(join(tmpdir(), 'docs-llms-'));
  onTestFinished(() => rm(outDir, { recursive: true, force: true }));
  const indexes = ['llms.txt', 'llms-full.txt'];
  await Promise.all(indexes.map((file) => writeFile(join(outDir, file), `General ${file}`)));
  await generateLlmsFiles(
    {
      outDir,
      site: { base: '/notopia/pr/123/' },
      pages: ['api/rest.md', 'api/rest/operations/createNote.md'],
    },
    restLlmsOptions
  );

  for (const file of indexes) {
    expect(await readFile(join(outDir, file), 'utf8')).toBe(`General ${file}`);
  }
  expect(await readFile(join(outDir, 'api/rest.md'), 'utf8')).toContain(
    '/notopia/pr/123/api/rest/operations/createNote.md'
  );
  const operation = await readFile(join(outDir, 'api/rest/operations/createNote.md'), 'utf8');
  expect(operation).toContain('/note/notes');
  expect(operation).not.toContain('<OAOperation');
});
