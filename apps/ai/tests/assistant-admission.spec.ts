import { simulateReadableStream } from 'ai';
import { MockLanguageModelV3 } from 'ai/test';
import { describe, expect, it, vi } from 'vitest';

import { AiService } from '../src/ai/ai.service';
import type { AssistantTask } from '../src/ai/assistant-task.types';

const provider = vi.hoisted(() => ({ model: undefined as any }));
vi.mock('@ai-sdk/openai-compatible', () => ({
  createOpenAICompatible: () => () => provider.model,
}));
const workspaceId = '00000000-0000-4000-8000-000000000001';
const noteId = '00000000-0000-4000-8000-000000000010';
const usageTokens = {
  inputTokens: { total: 1, noCache: 1, cacheRead: 0, cacheWrite: 0 },
  outputTokens: { total: 1, text: 1, reasoning: 0 },
};
const response = (open: boolean) => ({
  stream: simulateReadableStream({
    initialDelayInMs: null,
    chunkDelayInMs: null,
    chunks: [
      { type: 'stream-start', warnings: [] },
      ...(open
        ? [
            {
              type: 'tool-call',
              toolCallId: 'open',
              toolName: 'openNoteForEditing',
              input: JSON.stringify({ noteId }),
            },
          ]
        : [
            { type: 'text-start', id: 'answer' },
            { type: 'text-delta', id: 'answer', delta: 'Done' },
            { type: 'text-end', id: 'answer' },
          ]),
      {
        type: 'finish',
        finishReason: { unified: open ? 'tool-calls' : 'stop', raw: open ? 'tool_calls' : 'stop' },
        usage: usageTokens,
      },
    ] as any[],
  }),
});

describe('assistant admission across HTTP requests', () => {
  it('finishes an open continuation using its paid credit when no unused credits remain', async () => {
    provider.model = new MockLanguageModelV3({ doStream: [response(true), response(false)] });
    let credits = 1;
    let streams = 0;
    let saved: AssistantTask;
    const tasks = {
      create: async (
        owner: AssistantTask['owner'],
        promptId: string,
        messages: AssistantTask['messages']
      ) =>
        (saved = {
          id: 'task',
          owner,
          promptId,
          messages,
          deadline: Date.now() + 60000,
          remainingSteps: 3,
          pendingCalls: [],
          state: 'running',
          revision: 0,
        }),
      suspend: async (
        _id: string,
        _revision: number,
        messages: AssistantTask['messages'],
        pendingCalls: AssistantTask['pendingCalls'],
        targetNoteId: string,
        steps: number
      ) => {
        saved = {
          ...saved,
          messages,
          pendingCalls,
          targetNoteId,
          state: 'awaiting-client',
          remainingSteps: saved.remainingSteps - steps,
        };
        return { taskId: 'task', token: 'one-shot', expiresAt: saved.deadline };
      },
      claim: async () => {
        if (saved.state !== 'awaiting-client') throw new Error('Not pending');
        saved.state = 'running';
        return saved;
      },
      finish: async (_owner: unknown, _id: string, outcome: AssistantTask['state']) => {
        saved.state = outcome;
      },
    };
    const service = new AiService(
      { baseUrl: 'http://unused', apiKey: '', model: 'mock' },
      {
        timeoutMs: 60000,
        maxSteps: 3,
        maxOutputTokens: 100,
        requestsPerMinute: 20,
        concurrentStreams: 2,
      },
      {
        getNoteWorkspace: async () => ({ id: workspaceId }),
        getNote: async () => ({ id: noteId, workspaceId }),
      } as any,
      {} as any,
      {
        permissions: async () => ({ canRead: true, canWrite: true }),
        canRead: async () => true,
      } as any,
      { monthlyCredits: async () => credits } as any,
      {
        throttle: async () => {},
        reserve: async () => {
          if (credits !== 1) throw new Error('No credit');
          credits--;
          streams++;
          return {
            start: async () => {},
            release: async () => {
              streams--;
            },
          };
        },
        acquireContinuationLease: async () => {
          streams++;
          return {
            release: async () => {
              streams--;
            },
          };
        },
      } as any,
      tasks as any
    );
    const initial = await service.streamRequest(
      {
        assistantMode: true,
        workspaceId,
        messages: [{ id: 'prompt', role: 'user', parts: [{ type: 'text', text: 'Edit my note' }] }],
      },
      'user',
      new AbortController().signal
    );
    const first = [];
    for await (const chunk of initial!) first.push(chunk);
    expect(first.some((chunk) => chunk.type === 'data-assistant-continuation')).toBe(true);
    expect(credits).toBe(0);
    expect(streams).toBe(0);
    const continuation = await service.streamRequest(
      {
        action: 'continue',
        workspaceId,
        taskId: 'task',
        token: 'one-shot',
        noteId,
        results: [
          { toolName: 'openNoteForEditing', toolCallId: 'open', output: { noteId, ready: true } },
        ],
      },
      'user',
      new AbortController().signal
    );
    const second = [];
    for await (const chunk of continuation!) second.push(chunk);
    expect(second.some((chunk) => chunk.type === 'text-delta' && chunk.delta === 'Done')).toBe(
      true
    );
    expect(saved!.state).toBe('completed');
    expect(credits).toBe(0);
    expect(streams).toBe(0);
  });
});
