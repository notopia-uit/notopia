import { createHash, randomUUID } from 'node:crypto';

import { Redis } from 'ioredis';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import { AssistantTaskService } from '../src/usage/assistant-task.service';
import { UsageService } from '../src/usage/usage.service';

describe.skipIf(!process.env.AI_TEST_REDIS_URL)('assistant task admission', () => {
  const owner = { userId: randomUUID(), workspaceId: randomUUID() };
  const config = {
    redisUrl: process.env.AI_TEST_REDIS_URL!,
    timeoutMs: 120000,
    maxSteps: 10,
    concurrentStreams: 1,
    requestsPerMinute: 20,
  };
  let usage: UsageService;
  let tasks: AssistantTaskService;
  let redis: Redis;
  const owned: string[] = [];
  beforeAll(async () => {
    usage = new UsageService(config);
    tasks = new AssistantTaskService(usage);
    redis = new Redis(config.redisUrl);
    await new Promise<void>((resolve) => redis.once('ready', resolve));
    await vi.waitFor(() => usage.pingRedis());
  });
  afterAll(async () => {
    if (!redis) return;
    if (owned.length) await redis.del(...owned.map((id) => `ai:task:${id}`));
    const user = createHash('sha256').update(owner.userId).digest('hex');
    const accounting = await redis.keys(`ai:*:${owner.workspaceId}:*`);
    await redis.del(`ai:streams:${user}`, `ai:rate:${user}`, ...accounting);
    await redis.quit();
    usage.onModuleDestroy();
  });
  const create = async (timeoutMs = 120000) => {
    const task = await tasks.create(owner, randomUUID(), [], { timeoutMs, maxSteps: 10 });
    owned.push(task.id);
    return task;
  };
  const pending = [
    { toolCallId: 'open-1', toolName: 'openNoteForEditing', input: { noteId: owner.workspaceId } },
  ];
  it('claimIsSingleUse: parallel claims execute once', async () => {
    const task = await create();
    const ticket = await tasks.suspend(task.id, task.revision, [], pending, owner.workspaceId, 1);
    const results = await Promise.allSettled([
      tasks.claim(owner, task.id, ticket.token),
      tasks.claim(owner, task.id, ticket.token),
    ]);
    expect(results.filter((result) => result.status === 'fulfilled')).toHaveLength(1);
    const result = results.find((result) => result.status === 'fulfilled');
    expect(result?.value).toMatchObject({
      remainingSteps: 9,
      targetNoteId: owner.workspaceId,
      deadline: task.deadline,
    });
    expect(await redis.get(`ai:task:${task.id}`)).not.toContain(ticket.token);
  });
  it('ownerCannotBeChanged and forged tokens do not consume the real token', async () => {
    const task = await create();
    const ticket = await tasks.suspend(task.id, 0, [], pending, undefined, 1);
    await expect(
      tasks.claim({ ...owner, userId: 'other' }, task.id, ticket.token)
    ).rejects.toThrow();
    await expect(
      tasks.claim({ ...owner, workspaceId: 'other' }, task.id, ticket.token)
    ).rejects.toThrow();
    await expect(tasks.claim(owner, task.id, 'forged')).rejects.toThrow();
    await expect(tasks.claim(owner, task.id, ticket.token)).resolves.toMatchObject({
      state: 'running',
    });
  });
  it('expiredTaskCannotResume and suspension never extends expiry', async () => {
    const task = await create();
    const expiry = await redis.pexpiretime(`ai:task:${task.id}`);
    const ticket = await tasks.suspend(task.id, 0, [], pending, undefined, 1);
    expect(await redis.pexpiretime(`ai:task:${task.id}`)).toBe(expiry);
    await redis.pexpireat(`ai:task:${task.id}`, Date.now() - 1);
    await expect(tasks.claim(owner, task.id, ticket.token)).rejects.toThrow();
  });
  it.each(['completed', 'cancelled', 'failed'] as const)(
    'terminalTaskCannotResume: %s',
    async (outcome) => {
      const task = await create();
      const ticket = await tasks.suspend(task.id, 0, [], pending, undefined, 1);
      await tasks.finish(owner, task.id, outcome);
      await expect(tasks.claim(owner, task.id, ticket.token)).rejects.toThrow();
    }
  );
  it('preserves empty arrays inside recorded tool inputs and metadata', async () => {
    const task = await create();
    const messages = [
      {
        id: 'prompt',
        role: 'user' as const,
        parts: [{ type: 'text' as const, text: 'edit' }],
        metadata: { blocks: [] },
      },
    ];
    const calls = [
      { toolCallId: 'edit', toolName: 'applyDocumentOperations', input: { operations: [] } },
    ];
    const ticket = await tasks.suspend(task.id, 0, messages, calls, undefined, 1);
    const claimed = await tasks.claim(owner, task.id, ticket.token);
    expect(claimed.messages).toEqual(messages);
    expect(claimed.pendingCalls).toEqual(calls);
  });
  it('suspendCannotOverwriteCancelledTask', async () => {
    const task = await create();
    await tasks.finish(owner, task.id, 'cancelled');
    await expect(tasks.suspend(task.id, 0, [], pending, undefined, 1)).rejects.toThrow();
  });
  it('rejects exhausted steps and oversized transcripts', async () => {
    const task = await create();
    await expect(tasks.suspend(task.id, 0, [], pending, undefined, 10)).rejects.toThrow();
    await expect(
      tasks.suspend(
        task.id,
        0,
        [{ id: 'x', role: 'user', parts: [{ type: 'text', text: 'x'.repeat(512000) }] }],
        pending,
        undefined,
        1
      )
    ).rejects.toThrow();
  });
  it('continuationLeaseDoesNotDebit and shares initial admission capacity', async () => {
    const reservation = await usage.reserve(owner.workspaceId, owner.userId, 'prompt', 2);
    await reservation.start();
    await expect(usage.acquireContinuationLease(owner.userId)).rejects.toThrow();
    await reservation.release();
    const before = await redis.keys(`ai:usage:${owner.workspaceId}:*`);
    const lease = await usage.acquireContinuationLease(owner.userId);
    await expect(usage.reserve(owner.workspaceId, owner.userId, 'second', 2)).rejects.toThrow();
    expect(await redis.get(before[0])).toBe('1');
    await lease.release();
    await lease.release();
    const next = await usage.acquireContinuationLease(owner.userId);
    await next.release();
  });
  it('continuations and initial reservations share the rate limit', async () => {
    const user = createHash('sha256').update(owner.userId).digest('hex');
    await redis.set(`ai:rate:${user}`, 19, 'EX', 60);
    const lease = await usage.acquireContinuationLease(owner.userId);
    await lease.release();
    await expect(usage.acquireContinuationLease(owner.userId)).rejects.toThrow();
    await expect(
      usage.reserve(owner.workspaceId, owner.userId, 'rate-rejected', 2)
    ).rejects.toThrow();
  });
});
