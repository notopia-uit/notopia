import { mkdir, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

import { expect, it } from 'vitest';

it('copies Scalar below the VitePress API landing page', async ({ onTestFinished }) => {
  const root = await mkdtemp(join(tmpdir(), 'docs-scalar-'));
  onTestFinished(() => rm(root, { recursive: true, force: true }));
  const source = join(root, 'scalar');
  const output = join(root, 'site');
  await mkdir(join(source, 'assets'), { recursive: true });
  await mkdir(join(output, 'api'), { recursive: true });
  await writeFile(join(source, 'index.html'), 'Scalar');
  await writeFile(join(source, 'assets', 'app.js'), 'asset');
  await writeFile(join(output, 'api', 'index.html'), 'VitePress');

  const { copyScalarOutput } = await import('./scalar-output');
  await copyScalarOutput(source, output);

  expect(await readFile(join(output, 'api/index.html'), 'utf8')).toBe('VitePress');
  expect(await readFile(join(output, 'api/openapi-scalar/index.html'), 'utf8')).toBe('Scalar');
  expect(await readFile(join(output, 'api/openapi-scalar/assets/app.js'), 'utf8')).toBe('asset');
});

it('rejects a missing Scalar entry point', async ({ onTestFinished }) => {
  const root = await mkdtemp(join(tmpdir(), 'docs-scalar-'));
  onTestFinished(() => rm(root, { recursive: true, force: true }));
  const { copyScalarOutput } = await import('./scalar-output');
  await expect(copyScalarOutput(root, join(root, 'site'))).rejects.toThrow('api-web:build');
});
