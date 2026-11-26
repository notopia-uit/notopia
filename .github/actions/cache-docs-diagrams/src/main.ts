import * as cache from '@actions/cache';
import * as core from '@actions/core';

try {
  if (core.getBooleanInput('restore')) {
    const key = await cache.restoreCache(['docs/src/public/diagrams/'], 'docs-diagrams-', [
      'docs-diagrams-',
    ]);

    if (key) {
      core.saveState('matched-key', key);
    }
  }
} catch (error) {
  core.setFailed(error instanceof Error ? error.message : String(error));
}
