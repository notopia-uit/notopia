import { describe, expect, it } from 'vitest';

import { parseChatRequest } from '../src/ai/chat-request.dto';

const prompt = { id: 'prompt', role: 'user', parts: [{ type: 'text', text: 'Find my notes' }] };

describe('AI request validation', () => {
  it('accepts existing assistant attachments and BlockNote document state', async () => {
    const body = await parseChatRequest({
      messages: [
        {
          ...prompt,
          parts: [
            ...prompt.parts,
            { type: 'data-attached-note', data: { noteId: '123', name: 'Notes' } },
          ],
          metadata: { documentState: { blocks: [] } },
        },
      ],
      toolDefinitions: {},
    });
    expect(body.messages).toHaveLength(1);
  });
  it.each([
    { messages: [] },
    {
      messages: [
        {
          ...prompt,
          parts: [
            ...prompt.parts,
            { type: 'file', mediaType: 'image/png', url: 'http://localhost/private' },
          ],
        },
      ],
    },
    { messages: [{ ...prompt, role: 'system' }] },
    { messages: [{ ...prompt, parts: [{ type: 'unknown' }] }] },
    { messages: [prompt], workspaceId: 'bad-id' },
    { messages: [prompt], toolDefinitions: { searchNotes: {} } },
    { messages: [{ ...prompt, role: 'assistant' }] },
  ])('rejects invalid input %#', async (body) => {
    await expect(parseChatRequest(body)).rejects.toMatchObject({ status: 400 });
  });
  it('rejects oversized context', async () => {
    await expect(
      parseChatRequest({
        messages: [{ ...prompt, parts: [{ type: 'text', text: 'a'.repeat(512001) }] }],
      })
    ).rejects.toMatchObject({ status: 413 });
  });
});
