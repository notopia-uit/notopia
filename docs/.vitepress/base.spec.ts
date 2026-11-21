import { loadConfigFromFile } from 'vite';
import { afterEach, describe, expect, it, vi } from 'vitest';

afterEach(() => {
  vi.unstubAllEnvs();
  vi.resetModules();
});

describe.each([
  ['/notopia/', '/notopia/api/'],
  ['/notopia/pr/123/', '/notopia/pr/123/api/'],
])('site base %s', (base, apiBase) => {
  it('configures documentation and API asset paths', async () => {
    vi.stubEnv('DOCS_BASE', base);
    vi.stubEnv('DOCS_KROKI_URL', 'http://localhost:8000');
    const docs = (await import('./config.mjs')).default;
    const api = await loadConfigFromFile(
      { command: 'build', mode: 'production' },
      '../apps/api-web/vite.config.mts'
    );

    expect(docs.base).toBe(base);
    expect(docs.ignoreDeadLinks).toContain(`${apiBase}index.html`);
    expect(docs.themeConfig?.nav).toContainEqual(
      expect.objectContaining({ text: 'Scalar API', link: '/api/index.html' })
    );
    expect(api?.config.base).toBe(apiBase);
  });
});
