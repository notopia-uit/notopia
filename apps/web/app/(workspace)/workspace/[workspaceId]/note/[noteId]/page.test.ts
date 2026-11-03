import { afterEach, expect, test, vi } from 'vitest';

vi.mock('@lib/get-access-token', () => ({
  fetchAccessTokenServerSide: () => Promise.resolve('token'),
}));
vi.mock('./editor-boundary', () => ({ EditorBoundary: () => null }));

import NotePage from './page';

afterEach(() => vi.unstubAllEnvs());

test('passes the configured document websocket URL unchanged', async () => {
  vi.stubEnv('DOCUMENT_WS_URL', 'wss://example.test/document/ws/document');

  const page = await NotePage({
    params: Promise.resolve({ noteId: 'note', workspaceId: 'workspace' }),
  });

  expect(page.props.children.props.url).toBe('wss://example.test/document/ws/document');
});
