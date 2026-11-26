import * as cache from '@actions/cache';
import * as core from '@actions/core';

import { cachePrefix, diagramsPath } from './constants.ts';

try {
  if (core.getBooleanInput('restore')) {
    const key = await cache.restoreCache([diagramsPath], cachePrefix, [cachePrefix]);

    if (key) {
      core.saveState('matched-key', key);
    }
  }
} catch (error) {
  core.setFailed(error instanceof Error ? error.message : String(error));
}
