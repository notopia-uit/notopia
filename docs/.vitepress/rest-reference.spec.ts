import spec from '@notopia-uit/api/openapi' with { type: 'json' };
import type { DefaultTheme } from 'vitepress';
import { expect, it, vi } from 'vitest';

import { getRestReference } from './rest-reference';

it('groups one operation under each declared tag without duplicating routes', () => {
  const plan = { ...spec.paths['/billing/plans'].get, tags: ['Plan', 'Billing'] };
  const reference = getRestReference({
    ...spec,
    tags: [
      { ...spec.tags[0], name: 'Billing' },
      { ...spec.tags[0], name: 'Plan' },
    ],
    paths: { '/billing/plans': { get: plan } } as typeof spec.paths,
  });
  const operation = {
    operationId: 'listPlans',
    title: 'List plans',
    method: 'GET',
    path: '/billing/plans',
    href: '/api/rest/operations/listPlans',
  };
  expect(reference.operations).toEqual([operation]);
  expect(reference.groups).toEqual([
    { tag: 'Billing', operations: [operation] },
    { tag: 'Plan', operations: [operation] },
  ]);
});

it('maps the joined spec to a stable operation route', () => {
  const { groups } = getRestReference(spec);
  expect(groups.find(({ tag }) => tag === 'Note')?.operations).toContainEqual(
    expect.objectContaining({ operationId: 'createNote', href: '/api/rest/operations/createNote' })
  );
});

it('groups untagged operations under Other', () => {
  const reference = getRestReference({
    ...spec,
    paths: { '/plans': { get: { operationId: 'listPlans' } } } as unknown as typeof spec.paths,
  });
  expect(reference.groups).toEqual([
    { tag: 'Other', operations: [expect.objectContaining({ operationId: 'listPlans' })] },
  ]);
});

it.each([
  [
    { '/a': { get: { operationId: 'duplicate' } }, '/b': { post: { operationId: 'duplicate' } } },
    /duplicate/i,
  ],
  [{ '/a': { post: {} } }, /POST \/a/],
  [{ '/a': { get: { operationId: '../unsafe' } } }, /\.\.\/unsafe/],
])('rejects invalid operation IDs: %o', (paths, error) => {
  expect(() => getRestReference({ ...spec, paths: paths as unknown as typeof spec.paths })).toThrow(
    error
  );
});

it('uses tag navigation only for REST pages', async () => {
  vi.stubEnv('DOCS_KROKI_URL', 'http://localhost:8000');
  try {
    const { default: config } = await import('./config.mjs');
    const sidebar = config.themeConfig?.sidebar as DefaultTheme.SidebarMulti;
    const overview = sidebar['/api/rest.md'] as DefaultTheme.SidebarItem[];
    expect(overview.find(({ text }) => text === 'Note')?.items).toContainEqual(
      expect.objectContaining({ link: '/api/rest/operations/createNote' })
    );
    expect(sidebar['/api/rest/']).toEqual(overview);
    const general = sidebar['/'] as DefaultTheme.SidebarItem[];
    expect(general).not.toEqual(overview);
    expect(general.length).toBeGreaterThan(0);
    expect(JSON.stringify(sidebar)).not.toContain('[operationId]');
  } finally {
    vi.unstubAllEnvs();
  }
});
