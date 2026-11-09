import { describe, expect, it } from 'vitest';

import { parseAiRequest, parseChatRequest } from '../src/ai/chat-request.dto';

const prompt = { id: 'prompt', role: 'user', parts: [{ type: 'text', text: 'Find my notes' }] };

const continuation = {
  action: 'continue',
  workspaceId: '00000000-0000-4000-8000-000000000110',
  taskId: '00000000-0000-4000-8000-000000000111',
  token: 'a'.repeat(43),
  results: [
    {
      toolCallId: 'open-1',
      toolName: 'openNoteForEditing',
      output: { noteId: '00000000-0000-4000-8000-000000000112', ready: true },
    },
  ],
};

describe('assistant request protocol', () => {
  it('accepts a continuation without replacement history', async () => {
    expect(await parseAiRequest(continuation)).toEqual(continuation);
  });
  it('accepts owner-scoped cancellation without a prompt', async () => {
    const body = {
      action: 'cancel',
      workspaceId: continuation.workspaceId,
      taskId: continuation.taskId,
    };
    expect(await parseAiRequest(body)).toEqual(body);
  });
  it('preserves explicit assistant opt-in', async () => {
    expect(await parseAiRequest({ messages: [prompt], assistantMode: true })).toMatchObject({
      assistantMode: true,
    });
  });
  it.each([
    { ...continuation, messages: [prompt] },
    { ...continuation, prompt: 'replace the task' },
    { ...continuation, results: [] },
    { ...continuation, results: [...continuation.results, ...continuation.results] },
    { ...continuation, toolDefinitions: { openNoteForEditing: {} } },
    { action: 'unknown', messages: [prompt] },
  ])('rejects malformed or injected continuation %#', async (body) => {
    await expect(parseAiRequest(body)).rejects.toMatchObject({ status: 400 });
  });
  it('bounds fresh document state before resuming', async () => {
    await expect(
      parseAiRequest({ ...continuation, documentState: { content: 'a'.repeat(512001) } })
    ).rejects.toMatchObject({ status: 413 });
  });
});

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
