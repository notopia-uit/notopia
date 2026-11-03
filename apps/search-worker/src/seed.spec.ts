import { afterEach, describe, expect, it, vi } from 'vitest';

import { ConfigMeilisearch } from '../seed/seed';

describe('seed configuration', () => {
  afterEach(() => {
    vi.unstubAllEnvs();
    vi.unstubAllGlobals();
  });

  it.each([
    'NOTOPIA_SEARCH_WORKER_MEILI_HOST',
    'NOTOPIA_SEARCH_WORKER_MEILI_API_KEY',
    'NOTOPIA_SEARCH_WORKER_MEILI_NOTE_SEARCH_KEY_UID',
  ])('rejects missing %s before contacting Meilisearch', async (missing) => {
    vi.stubEnv('NOTOPIA_SEARCH_WORKER_MEILI_HOST', 'http://localhost:7700');
    vi.stubEnv('NOTOPIA_SEARCH_WORKER_MEILI_API_KEY', 'secret');
    vi.stubEnv(
      'NOTOPIA_SEARCH_WORKER_MEILI_NOTE_SEARCH_KEY_UID',
      '00000000-0000-4000-8000-000000000001'
    );
    vi.stubEnv(missing, '');
    await expect(ConfigMeilisearch()).rejects.toThrow('are required');
  });

  it('creates the configured key only once across repeated runs', async () => {
    const uid = '00000000-0000-4000-8000-000000000001';
    vi.stubEnv('NOTOPIA_SEARCH_WORKER_MEILI_HOST', 'http://localhost:7700');
    vi.stubEnv('NOTOPIA_SEARCH_WORKER_MEILI_API_KEY', 'secret');
    vi.stubEnv('NOTOPIA_SEARCH_WORKER_MEILI_NOTE_SEARCH_KEY_UID', uid);
    let created = 0;
    vi.stubGlobal('fetch', (input: string, init?: { method?: string; body?: string }) => {
      const url = String(input);
      let data: unknown;
      let status = 200;
      if (url.endsWith('/indexes/notes/settings')) {
        data = {
          searchableAttributes: ['name', 'folderName', 'plainTextContent', 'tags'],
          filterableAttributes: ['workspaceId', 'tags'],
          sortableAttributes: ['name'],
        };
      } else if (url.endsWith('/indexes/notes')) {
        data = { uid: 'notes', primaryKey: 'id' };
      } else if (url.endsWith(`/keys/${uid}`)) {
        status = created ? 200 : 404;
        data = created
          ? { uid, key: 'private-search-key' }
          : { message: 'not found', code: 'api_key_not_found' };
      } else if (url.endsWith('/keys') && init?.method === 'POST') {
        expect(JSON.parse(init.body ?? '').uid).toBe(uid);
        created++;
        data = { uid, key: 'private-search-key' };
      } else {
        throw new Error(`Unexpected request ${url}`);
      }
      return Promise.resolve(
        new Response(JSON.stringify(data), {
          status,
          headers: { 'content-type': 'application/json' },
        })
      );
    });
    await ConfigMeilisearch();
    await ConfigMeilisearch();
    expect(created).toBe(1);
  });
});
