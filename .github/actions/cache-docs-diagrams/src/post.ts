import { createHash } from 'node:crypto';
import { mkdir, readFile, readdir } from 'node:fs/promises';
import { join, relative } from 'node:path';

import * as cache from '@actions/cache';
import * as core from '@actions/core';

import { cachePrefix, diagramsPath } from './constants.ts';

try {
  await mkdir(diagramsPath, { recursive: true });

  const entries = await readdir(diagramsPath, { recursive: true, withFileTypes: true });
  const files = entries
    .filter((entry) => entry.isFile())
    .map((entry) => join(entry.parentPath, entry.name))
    .sort();
  const hash = createHash('sha256');

  for (const file of files) {
    const fileHash = createHash('sha256')
      .update(await readFile(file))
      .digest('hex');
    hash.update(JSON.stringify([relative(diagramsPath, file), fileHash]));
  }

  const key = `${cachePrefix}${hash.digest('hex')}`;

  if (key !== core.getState('matched-key')) {
    await cache.saveCache([diagramsPath], key);
  }
} catch (error) {
  core.setFailed(error instanceof Error ? error.message : String(error));
}
