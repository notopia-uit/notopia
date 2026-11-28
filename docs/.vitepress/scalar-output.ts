import { cp, stat } from 'node:fs/promises';
import { join } from 'node:path';

export async function copyScalarOutput(sourceDir: string, siteOutDir: string): Promise<void> {
  const entry = join(sourceDir, 'index.html');
  if (!(await stat(entry, { throwIfNoEntry: false }))?.isFile()) {
    throw new Error(`Missing Scalar entry point at ${entry}; run api-web:build first.`);
  }
  await cp(sourceDir, join(siteOutDir, 'api/openapi-scalar'), { recursive: true });
}
