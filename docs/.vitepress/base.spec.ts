import { loadConfigFromFile } from 'vite';
import { afterEach, describe, expect, it, vi } from 'vitest';

afterEach(() => {
  vi.unstubAllEnvs();
  vi.resetModules();
});

describe.each([
  ['/notopia/', '/notopia/api/openapi-scalar/'],
  ['/notopia/pr/123/', '/notopia/pr/123/api/openapi-scalar/'],
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
    expect(docs.ignoreDeadLinks).toEqual([`${apiBase}index.html`]);
    expect(docs.themeConfig?.nav).toContainEqual(
      expect.objectContaining({ text: 'Scalar API', link: '/api/openapi-scalar/index.html' })
    );
    expect(api?.config.base).toBe(apiBase);
  });
});

it('retains the standalone Scalar dev base', async () => {
  vi.stubEnv('DOCS_BASE', '');
  const api = await loadConfigFromFile(
    { command: 'serve', mode: 'development', isPreview: false },
    '../apps/api-web/vite.config.mts'
  );
  expect(api?.config.base).toBe('/notopia/api/');
});
