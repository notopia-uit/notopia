import { afterEach, describe, expect, it, vi } from 'vitest';

import { createAiTransport, fetchAi } from './ai-transport';

vi.mock('./get-access-token-client-side', () => ({
  fetchAccessTokenClientSide: async () => 'test-token',
}));
afterEach(() => vi.unstubAllGlobals());

describe('AI transport', () => {
  it('preserves the stream and merges workspace context with BlockNote tool definitions', async () => {
    const response = new Response('data: [DONE]\n\n', {
      headers: { 'Content-Type': 'text/event-stream' },
    });
    const fetch = vi.fn().mockResolvedValue(response);
    vi.stubGlobal('fetch', fetch);
    const transport = createAiTransport('https://api.example.test/ai', { noteId: 'linked-note' });
    const stream = await transport.sendMessages({
      trigger: 'submit-message',
      chatId: 'chat',
      messages: [{ id: 'user', role: 'user', parts: [{ type: 'text', text: 'Edit this' }] }],
      body: { toolDefinitions: { applyDocumentOperations: {} } },
    });
    const [url, init] = fetch.mock.calls[0];
    expect(url).toBe('https://api.example.test/ai');
    expect(JSON.parse(init.body)).toMatchObject({
      noteId: 'linked-note',
      toolDefinitions: { applyDocumentOperations: {} },
    });
    expect(new Headers(init.headers).get('Authorization')).toBe('Bearer test-token');
    expect((await stream.getReader().read()).done).toBe(true);
  });
  it('sends the explicit workspace even when the user has several', async () => {
    const fetch = vi.fn().mockResolvedValue(new Response('data: [DONE]\n\n'));
    vi.stubGlobal('fetch', fetch);
    await createAiTransport('/ai', { workspaceId: 'selected-workspace' }).sendMessages({
      trigger: 'submit-message',
      chatId: 'chat',
      messages: [],
    });
    expect(JSON.parse(fetch.mock.calls[0][1].body).workspaceId).toBe('selected-workspace');
  });
  it.each([
    [401, 'session expired'],
    [402, 'no AI credits'],
    [403, 'permission'],
    [429, 'Wait a minute'],
    [503, 'temporarily unavailable'],
    [504, 'timed out'],
  ])('normalizes HTTP %i for both chat interfaces', async (status, expected) => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          Response.json({ code: 'serverCode', message: 'raw message' }, { status })
        )
    );
    await expect(fetchAi('/ai')).rejects.toMatchObject({
      statusCode: status,
      code: 'serverCode',
      message: expect.stringContaining(expected),
    });
  });
  it('shows a quota reset date without leaking the JSON body', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          Response.json(
            { code: 'planLimitExceeded', resetAt: '2026-11-01T00:00:00Z' },
            { status: 402 }
          )
        )
    );
    await expect(fetchAi('/ai')).rejects.toMatchObject({ message: expect.stringContaining('UTC') });
  });
  it('does not render an HTML gateway error', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(new Response('<html>private gateway details</html>', { status: 502 }))
    );
    await expect(fetchAi('/ai')).rejects.toMatchObject({
      statusCode: 502,
      message: 'The AI request failed. Please try again.',
    });
  });
});
