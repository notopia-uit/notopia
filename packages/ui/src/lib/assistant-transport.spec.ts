import { describe, expect, it, vi } from 'vitest';

import { createAssistantTransport } from './assistant-transport';

vi.mock('./get-access-token-client-side', () => ({
  fetchAccessTokenClientSide: async () => 'token',
}));

describe('assistant transport', () => {
  it('sends continuation results and live context without replacement history', async () => {
    let sent: Record<string, unknown> = {};
    vi.stubGlobal('fetch', async (_url: unknown, init: RequestInit) => {
      sent = JSON.parse(init.body as string);
      return new Response('data: [DONE]\n\n', { headers: { 'content-type': 'text/event-stream' } });
    });
    const transport = createAssistantTransport('/ai', 'workspace');
    await transport.sendMessages({
      chatId: 'chat',
      trigger: 'submit-message',
      messageId: undefined,
      messages: [
        { id: 'client-history', role: 'user', parts: [{ type: 'text', text: 'Not a new prompt' }] },
      ],
      abortSignal: undefined,
      body: {
        action: 'continue',
        taskId: 'task',
        token: 'one-shot',
        results: [],
        noteId: 'note',
        documentState: { blocks: [] },
        toolDefinitions: {},
      },
    });
    expect(sent).toEqual({
      action: 'continue',
      workspaceId: 'workspace',
      taskId: 'task',
      token: 'one-shot',
      results: [],
      noteId: 'note',
      documentState: { blocks: [] },
      toolDefinitions: {},
    });
    vi.unstubAllGlobals();
  });
});
