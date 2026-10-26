import { createHash, randomUUID } from 'node:crypto';

import { Inject, Injectable, Logger, OnModuleDestroy } from '@nestjs/common';
import type { ConfigType } from '@nestjs/config';
import { Redis } from 'ioredis';

import { AiError } from '#/common/ai-error';
import { usageConfig } from '#/config.factory';

import { RELEASE, RESERVE, START, THROTTLE } from './usage.scripts';

export function usagePeriod(now = new Date()) {
  const resetAt = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 1));
  return {
    month: now.toISOString().slice(0, 7),
    resetAt: resetAt.toISOString(),
    expiresAt: Math.floor(resetAt.getTime() / 1000) + 7 * 86400,
  };
}

@Injectable()
export class UsageService implements OnModuleDestroy {
  private readonly logger = new Logger(UsageService.name);
  private readonly redis?: Redis;

  constructor(@Inject(usageConfig.KEY) private readonly config: ConfigType<typeof usageConfig>) {
    if (config.redisUrl) {
      this.redis = new Redis(config.redisUrl, {
        connectTimeout: 3000,
        commandTimeout: 5000,
        maxRetriesPerRequest: 1,
        enableOfflineQueue: false,
        autoResendUnfulfilledCommands: false,
      });
      this.redis.on('error', (error) => this.logger.error({ err: error }, 'AI Redis unavailable'));
    }
  }

  onModuleDestroy() {
    this.redis?.disconnect();
  }

  private async eval(script: string, keys: string[], args: (string | number)[]) {
    try {
      if (!this.redis) throw new Error('NOTOPIA_AI_REDIS_URL is not configured');
      return await this.redis.eval(script, keys.length, ...keys, ...args);
    } catch (error) {
      this.logger.error({ err: error }, 'AI usage accounting unavailable');
      throw new AiError(
        503,
        'aiUsageUnavailable',
        'AI usage accounting is unavailable. Please retry later.'
      );
    }
  }

  async throttle(userId: string) {
    const user = createHash('sha256').update(userId).digest('hex');
    const accepted = await this.eval(
      THROTTLE,
      [`ai:attempts:${user}`],
      [this.config.requestsPerMinute]
    );
    if (accepted !== 1)
      throw new AiError(429, 'aiRateLimited', 'Too many AI requests. Please retry later.');
  }

  async reserve(workspaceId: string, userId: string, messageId: string, credits: number) {
    const period = usagePeriod();
    if (credits === 0) {
      throw new AiError(402, 'aiCreditsExhausted', 'This workspace has no AI credits available.', {
        limit: credits,
        resetAt: period.resetAt,
      });
    }
    const digest = (value: string) => createHash('sha256').update(value).digest('hex');
    const user = digest(userId);
    const id = randomUUID();
    const keys = [
      `ai:usage:${workspaceId}:${period.month}`,
      `ai:request:${workspaceId}:${period.month}:${digest(JSON.stringify([userId, messageId]))}`,
      `ai:streams:${user}`,
      `ai:rate:${user}`,
    ];
    const result = await this.eval(RESERVE, keys, [
      Date.now(),
      credits,
      period.expiresAt,
      this.config.concurrentStreams,
      this.config.requestsPerMinute,
      this.config.timeoutMs + 5000,
      id,
    ]);
    if (result === 'duplicate')
      throw new AiError(
        409,
        'duplicateAiRequest',
        'This prompt has already been submitted. Send a new message to generate again.'
      );
    if (result === 'rate' || result === 'concurrency')
      throw new AiError(429, 'aiRateLimited', 'Too many AI requests. Please retry later.');
    if (result === 'quota')
      throw new AiError(402, 'aiCreditsExhausted', 'Monthly AI credits are exhausted.', {
        limit: credits,
        resetAt: period.resetAt,
      });
    if (result !== 'ok')
      throw new AiError(503, 'aiUsageUnavailable', 'Could not reserve AI credits.');
    let released = false;
    return {
      start: async () => {
        const started = await this.eval(START, [keys[1]], [id]);
        if (started !== 1)
          throw new AiError(409, 'duplicateAiRequest', 'This generation cannot be started again.');
      },
      release: async (refund = false) => {
        if (released) return;
        await this.eval(RELEASE, keys.slice(0, 3), [id, refund ? '1' : '0']);
        released = true;
      },
    };
  }
}
