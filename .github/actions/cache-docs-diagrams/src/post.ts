import { execFileSync } from 'node:child_process';
import { mkdir } from 'node:fs/promises';

import * as cache from '@actions/cache';
import * as core from '@actions/core';

try {
  const path = 'docs/src/public/diagrams/';
  await mkdir(path, { recursive: true });

  const hash = execFileSync(
    'bash',
    [
      '-o',
      'pipefail',
      '-c',
      "find docs/src/public/diagrams -type f -print0 | sort -z | xargs -0 -r sha256sum | sha256sum | cut -d ' ' -f 1",
    ],
    { encoding: 'utf8' }
  ).trim();
  const key = `docs-diagrams-${hash}`;

  if (key !== core.getState('matched-key')) {
    await cache.saveCache([path], key);
  }
} catch (error) {
  core.setFailed(error instanceof Error ? error.message : String(error));
}
