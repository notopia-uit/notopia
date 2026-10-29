import { simulateReadableStream } from 'ai';
import { MockLanguageModelV3 } from 'ai/test';
import { describe, expect, it } from 'vitest';

import { streamAssistantTask } from '../src/ai/assistant-stream';
import type { AssistantTask } from '../src/ai/assistant-task.types';

const noteId = '00000000-0000-4000-8000-000000000010';
const task: AssistantTask = {
  id: 'task',
  owner: { userId: 'user', workspaceId: 'workspace' },
  promptId: 'prompt',
  messages: [{ id: 'prompt', role: 'user', parts: [{ type: 'text', text: 'Edit my note' }] }],
  deadline: Date.now() + 60000,
  remainingSteps: 3,
  pendingCalls: [],
  state: 'running',
  revision: 0,
};
function model() {
  return new MockLanguageModelV3({
    doStream: {
      stream: simulateReadableStream({
        initialDelayInMs: null,
        chunkDelayInMs: null,
        chunks: [
          { type: 'stream-start', warnings: [] },
          { type: 'tool-input-start', id: 'open', toolName: 'openNoteForEditing' },
          { type: 'tool-input-delta', id: 'open', delta: JSON.stringify({ noteId }) },
          { type: 'tool-input-end', id: 'open' },
          {
            type: 'tool-call',
            toolCallId: 'open',
            toolName: 'openNoteForEditing',
            input: JSON.stringify({ noteId }),
          },
          {
            type: 'finish',
            finishReason: { unified: 'tool-calls', raw: 'tool_calls' },
            usage: {
              inputTokens: { total: 1, noCache: 1, cacheRead: 0, cacheWrite: 0 },
              outputTokens: { total: 1, text: 1, reasoning: 0 },
            },
          },
        ] as any[],
      }),
    },
  });
}
describe('assistant stream handoff', () => {
  it('does not advertise editing tools to a viewer', async () => {
    let advertised: unknown;
    const viewerModel = new MockLanguageModelV3({
      doStream: async (options) => {
        advertised = options.tools;
        return {
          stream: simulateReadableStream({
            initialDelayInMs: null,
            chunkDelayInMs: null,
            chunks: [
              { type: 'stream-start', warnings: [] },
              {
                type: 'finish',
                finishReason: { unified: 'stop', raw: 'stop' },
                usage: {
                  inputTokens: { total: 1, noCache: 1, cacheRead: 0, cacheWrite: 0 },
                  outputTokens: { total: 1, text: 1, reasoning: 0 },
                },
              },
            ],
          }),
        };
      },
    });
    const stream = await streamAssistantTask({
      task,
      model: viewerModel,
      tools: {},
      canEdit: false,
      maxOutputTokens: 100,
      signal: new AbortController().signal,
      authorize: async () => {
        throw new Error('Unexpected edit');
      },
      suspend: async () => {
        throw new Error('Unexpected handoff');
      },
      finish: async () => {},
      release: async () => {},
    });
    const chunks = [];
    for await (const chunk of stream) chunks.push(chunk);
    expect(chunks.some((chunk) => chunk.type === 'error')).toBe(false);
    expect(advertised).not.toEqual(
      expect.arrayContaining([expect.objectContaining({ name: 'openNoteForEditing' })])
    );
  });
  it('authorizes before exposing tool input and persists history before issuing a token', async () => {
    let authorized = false;
    let saved: AssistantTask['messages'] = [];
    const stream = await streamAssistantTask({
      task,
      model: model(),
      tools: {},
      maxOutputTokens: 100,
      signal: new AbortController().signal,
      authorize: async () => {
        authorized = true;
        return noteId;
      },
      suspend: async (messages, calls, target, steps) => {
        expect(authorized).toBe(true);
        expect(calls).toEqual([
          { toolCallId: 'open', toolName: 'openNoteForEditing', input: { noteId } },
        ]);
        expect(target).toBe(noteId);
        expect(steps).toBe(1);
        saved = messages;
        return { taskId: task.id, token: 'token', expiresAt: task.deadline };
      },
      finish: async () => {},
      release: async () => {},
    });
    const chunks = [];
    for await (const chunk of stream) {
      if (chunk.type.startsWith('tool-input')) expect(authorized).toBe(true);
      if (chunk.type === 'data-assistant-continuation')
        expect(saved.at(-1)?.role).toBe('assistant');
      chunks.push(chunk);
    }
    expect(chunks.some((chunk) => chunk.type === 'data-assistant-continuation')).toBe(true);
  });
  it('does not expose unauthorized target calls or a continuation', async () => {
    let outcome = '';
    const stream = await streamAssistantTask({
      task,
      model: model(),
      tools: {},
      maxOutputTokens: 100,
      signal: new AbortController().signal,
      authorize: async () => {
        throw new Error('Access denied');
      },
      suspend: async () => {
        throw new Error('Must not suspend');
      },
      finish: async (value) => {
        outcome = value;
      },
      release: async () => {},
    });
    const chunks = [];
    for await (const chunk of stream) chunks.push(chunk);
    expect(chunks.some((chunk) => chunk.type.startsWith('tool-input'))).toBe(false);
    expect(chunks.some((chunk) => chunk.type === 'data-assistant-continuation')).toBe(false);
    expect(outcome).toBe('failed');
  });
});
