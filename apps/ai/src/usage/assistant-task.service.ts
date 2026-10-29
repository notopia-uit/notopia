import { createHash, randomBytes, randomUUID } from 'node:crypto';

import { Injectable } from '@nestjs/common';
import type { UIMessage } from 'ai';

import type {
  AssistantTask,
  PendingClientCall,
  TaskOutcome,
  TaskOwner,
} from '#/ai/assistant-task.types';
import { AiError } from '#/common/ai-error';

import { CLAIM_TASK, CREATE_TASK, FINISH_TASK, SUSPEND_TASK } from './assistant-task.scripts';
import { UsageService } from './usage.service';

const hash = (token: string) => createHash('sha256').update(token).digest('hex');

@Injectable()
export class AssistantTaskService {
  constructor(private readonly usage: UsageService) {}

  private encode(value: unknown) {
    const encoded = JSON.stringify(value);
    if (Buffer.byteLength(encoded) > 512000) {
      throw new AiError(413, 'requestTooLarge', 'Assistant task context is too large.');
    }
    return encoded;
  }

  private invalid(): never {
    throw new AiError(
      409,
      'invalidAssistantTask',
      'This assistant task cannot continue. Send a new prompt.'
    );
  }

  async create(
    owner: TaskOwner,
    promptId: string,
    messages: UIMessage[],
    limits: { timeoutMs: number; maxSteps: number }
  ): Promise<AssistantTask> {
    const task: AssistantTask = {
      id: randomUUID(),
      owner,
      promptId,
      messages,
      deadline: Date.now() + limits.timeoutMs,
      remainingSteps: limits.maxSteps,
      pendingCalls: [],
      state: 'running',
      revision: 0,
    };
    const created = await this.usage.eval(
      CREATE_TASK,
      [`ai:task:${task.id}`],
      [
        this.encode({ ...task, messages: JSON.stringify(messages), pendingCalls: '[]' }),
        task.deadline,
      ]
    );
    if (created !== 1) this.invalid();
    return task;
  }

  async suspend(
    taskId: string,
    revision: number,
    messages: UIMessage[],
    pendingCalls: PendingClientCall[],
    targetNoteId: string | undefined,
    stepsUsed: number
  ) {
    if (!Number.isSafeInteger(stepsUsed) || !pendingCalls.length) this.invalid();
    const token = randomBytes(32).toString('base64url');
    const expiresAt = await this.usage.eval(
      SUSPEND_TASK,
      [`ai:task:${taskId}`],
      [
        revision,
        this.encode({
          messages: JSON.stringify(messages),
          pendingCalls: JSON.stringify(pendingCalls),
          targetNoteId,
          stepsUsed,
        }),
        hash(token),
      ]
    );
    if (typeof expiresAt !== 'string') this.invalid();
    return { taskId, token, expiresAt: Number(expiresAt) };
  }

  async claim(owner: TaskOwner, taskId: string, token: string): Promise<AssistantTask> {
    const raw = await this.usage.eval(
      CLAIM_TASK,
      [`ai:task:${taskId}`],
      [owner.userId, owner.workspaceId, hash(token)]
    );
    if (typeof raw !== 'string') this.invalid();
    const task = JSON.parse(raw) as Omit<AssistantTask, 'messages' | 'pendingCalls'> & {
      messages: string;
      pendingCalls: string;
    };
    // Keep nested JSON opaque to Lua, which changes empty arrays into objects.
    return {
      ...task,
      messages: JSON.parse(task.messages),
      pendingCalls: JSON.parse(task.pendingCalls),
    };
  }

  async finish(owner: TaskOwner, taskId: string, outcome: TaskOutcome): Promise<void> {
    const result = await this.usage.eval(
      FINISH_TASK,
      [`ai:task:${taskId}`],
      [owner.userId, owner.workspaceId, outcome]
    );
    if (result !== 1) this.invalid();
  }
}
