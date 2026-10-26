import { simulateReadableStream } from 'ai';
import { MockLanguageModelV3 } from 'ai/test';
import { describe, expect, it, vi } from 'vitest';

import { AiService } from '../src/ai/ai.service';
import { AiError } from '../src/common/ai-error';

const provider = vi.hoisted(() => ({ model: undefined as any }));
vi.mock('@ai-sdk/openai-compatible', () => ({
  createOpenAICompatible: () => () => provider.model,
}));

const workspaceId = '00000000-0000-4000-8000-000000000001';
const config = {
  timeoutMs: 1000,
  requestsPerMinute: 20,
  concurrentStreams: 2,
  maxSteps: 3,
  maxOutputTokens: 100,
};
const body = {
  workspaceId,
  messages: [
    { id: 'message', role: 'user' as const, parts: [{ type: 'text' as const, text: 'Help me' }] },
  ],
};
const tokenUsage = {
  inputTokens: { total: 1, noCache: 1, cacheRead: 0, cacheWrite: 0 },
  outputTokens: { total: 1, text: 1, reasoning: 0 },
};
const textResult = () => ({
  stream: simulateReadableStream({
    initialDelayInMs: null,
    chunkDelayInMs: null,
    chunks: [
      { type: 'stream-start', warnings: [] },
      { type: 'text-start', id: 'text' },
      { type: 'text-delta', id: 'text', delta: 'Hello' },
      { type: 'text-end', id: 'text' },
      { type: 'finish', finishReason: { unified: 'stop', raw: 'stop' }, usage: tokenUsage },
    ] as any[],
  }),
});
function setup() {
  const reservation = {
    start: vi.fn().mockResolvedValue(undefined),
    release: vi.fn().mockResolvedValue(undefined),
  };
  const usage = {
    throttle: vi.fn().mockResolvedValue(undefined),
    reserve: vi.fn().mockResolvedValue(reservation),
  };
  const authorization = {
    permissions: vi.fn().mockResolvedValue({ canRead: true, canWrite: true }),
    canRead: vi.fn().mockResolvedValue(true),
    getUserWorkspaces: vi.fn().mockResolvedValue([{ workspaceId }]),
  };
  const billing = {
    monthlyCredits: vi.fn().mockResolvedValue(200),
    requiredPlan: vi.fn().mockResolvedValue('plus'),
  };
  const service = new AiService(
    { baseUrl: 'http://unused.test', apiKey: 'unused', model: 'mock' },
    config,
    {} as any,
    {} as any,
    authorization as any,
    billing as any,
    usage as any
  );
  return { service, reservation, usage, authorization, billing };
}
async function collect(stream: ReadableStream) {
  const chunks: any[] = [];
  for await (const chunk of stream) chunks.push(chunk);
  return chunks;
}
describe('AI stream lifecycle', () => {
  it('charges once across server tool steps and releases the stream lease', async () => {
    provider.model = new MockLanguageModelV3({
      doStream: [
        {
          stream: simulateReadableStream({
            initialDelayInMs: null,
            chunkDelayInMs: null,
            chunks: [
              { type: 'stream-start', warnings: [] },
              { type: 'tool-call', toolCallId: 'call', toolName: 'listWorkspaces', input: '{}' },
              {
                type: 'finish',
                finishReason: { unified: 'tool-calls', raw: 'tool_calls' },
                usage: tokenUsage,
              },
            ] as any[],
          }),
        },
        textResult(),
      ],
    });
    const { service, usage, reservation } = setup();
    const chunks = await collect(
      await service.streamBlockNoteChat(body, 'user', new AbortController().signal)
    );
    expect(chunks.some((chunk) => chunk.type === 'tool-output-available')).toBe(true);
    expect(chunks.some((chunk) => chunk.type === 'text-delta' && chunk.delta === 'Hello')).toBe(
      true
    );
    expect(provider.model.doStreamCalls).toHaveLength(2);
    expect(usage.reserve).toHaveBeenCalledTimes(1);
    expect(reservation.start).toHaveBeenCalledTimes(1);
    expect(reservation.release).toHaveBeenCalledWith(false);
  });
  it('masks provider errors and keeps the credit for a started generation', async () => {
    provider.model = new MockLanguageModelV3({
      doStream: async () => {
        throw new Error('secret provider credentials');
      },
    });
    const { service, reservation } = setup();
    const chunks = await collect(
      await service.streamBlockNoteChat(body, 'user', new AbortController().signal)
    );
    const errors = chunks.filter((chunk) => chunk.type === 'error');
    expect(errors).not.toHaveLength(0);
    expect(JSON.stringify(errors)).not.toContain('secret');
    expect(reservation.release).toHaveBeenCalledWith(false);
  });
  it('refunds if the client disconnects before a generation starts', async () => {
    const { service, reservation, usage } = setup();
    const abort = new AbortController();
    usage.reserve.mockImplementation(async () => {
      abort.abort();
      return reservation;
    });
    await expect(service.streamBlockNoteChat(body, 'user', abort.signal)).rejects.toBeDefined();
    expect(reservation.start).not.toHaveBeenCalled();
    expect(reservation.release).toHaveBeenCalledWith(true);
  });
  it('propagates a timeout to the provider and releases the active stream', async () => {
    provider.model = new MockLanguageModelV3({
      doStream: async ({ abortSignal }) => ({
        stream: new ReadableStream({
          start(controller) {
            controller.enqueue({ type: 'stream-start', warnings: [] });
            abortSignal!.addEventListener('abort', () => controller.error(abortSignal!.reason), {
              once: true,
            });
          },
        }),
      }),
    });
    const { service, reservation } = setup();
    const abort = new AbortController();
    const chunks = collect(await service.streamBlockNoteChat(body, 'user', abort.signal));
    await vi.waitFor(() => expect(provider.model.doStreamCalls).toHaveLength(1));
    abort.abort(new AiError(504, 'aiRequestTimeout', 'The AI request timed out.'));
    const result = await chunks;
    expect(result.some((chunk) => chunk.type === 'error')).toBe(true);
    expect(provider.model.doStreamCalls[0].abortSignal.aborted).toBe(true);
    expect(reservation.release).toHaveBeenCalledWith(false);
  });
  it('denies unauthorized requests before billing and accounting', async () => {
    const { service, billing, usage, authorization } = setup();
    authorization.permissions.mockResolvedValue({ canRead: false, canWrite: false });
    await expect(
      service.streamBlockNoteChat(body, 'user', new AbortController().signal)
    ).rejects.toMatchObject({ status: 403 });
    expect(billing.monthlyCredits).not.toHaveBeenCalled();
    expect(usage.reserve).not.toHaveBeenCalled();
  });
});
