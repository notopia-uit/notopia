#!/usr/bin/env node

import { spawn } from 'node:child_process';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');

/**
 * @typedef {Object} UploadEntry
 * @property {string} file - Path to the coverage report.
 * @property {string} flag - Codecov flag name (from .codecov.yml).
 */

/** @type {UploadEntry[]} */
const entries = [
  { file: resolve(root, 'apps/document/test-report.junit.xml'), flag: 'document' },
  { file: resolve(root, 'cmd/authorization/coverage.out'), flag: 'authorization' },
];

async function upload({ file, flag }) {
  return new Promise((resolve, reject) => {
    const proc = spawn(
      'codecov',
      ['--disable-telem', 'upload-process', '--disable-search', '--file', file, '--flag', flag],
      { stdio: 'inherit' }
    );
    proc.on('exit', (code) => {
      if (code === 0) resolve();
      else reject(new Error(`codecov exited with code ${code} for ${file}`));
    });
    proc.on('error', reject);
  });
}

try {
  await Promise.all(entries.map(upload));
  console.log('All coverage reports uploaded successfully.');
} catch (err) {
  console.error('Coverage upload failed:', err.message);
  process.exit(1);
}
