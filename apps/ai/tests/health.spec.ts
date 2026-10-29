import type { Response } from 'express';
import { describe, expect, it, vi } from 'vitest';

import { HealthController } from '../src/health.controller';
import type { UsageService } from '../src/usage/usage.service';

function controller(pingRedis: () => Promise<void>) {
  return new HealthController({ pingRedis } as unknown as UsageService);
}

describe('AI health', () => {
  it('reports ready when Redis answers', async () => {
    const status = vi.fn();
    const health = controller(async () => {});

    await expect(health.ready({ status } as unknown as Response)).resolves.toEqual({
      status: 'up',
      details: { redis: { status: 'up' } },
    });
    expect(status).not.toHaveBeenCalled();
  });

  it('reports not ready when Redis fails', async () => {
    const status = vi.fn();
    const health = controller(async () => {
      throw new Error('Redis unavailable');
    });

    await expect(health.ready({ status } as unknown as Response)).resolves.toEqual({
      status: 'down',
      details: { redis: { status: 'down' } },
    });
    expect(status).toHaveBeenCalledWith(503);
  });
});
