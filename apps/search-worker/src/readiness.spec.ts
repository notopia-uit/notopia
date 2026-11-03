import { HealthCheckService, HealthIndicatorService } from '@nestjs/terminus';
import { describe, expect, it, vi } from 'vitest';

import { HealthController, KafkaConsumerState } from './health';

describe('health', () => {
  it('keeps liveness independent of Kafka and Meilisearch', async () => {
    const state = new KafkaConsumerState();
    const checker = {
      check: async (checks: (() => unknown)[]) => {
        const results = await Promise.all(checks.map((check) => check()));
        if (JSON.stringify(results).includes('down')) throw new Error('unhealthy');
        return results;
      },
    } as unknown as HealthCheckService;
    const meili = { health: vi.fn().mockResolvedValue({ status: 'available' }) };
    const health = new HealthController(
      state,
      meili as never,
      checker,
      new HealthIndicatorService()
    );
    expect(health.live()).toEqual({ status: 'ok' });
    await expect(health.ready()).rejects.toThrow();
    state.connected = true;
    await expect(health.ready()).resolves.toBeDefined();
    meili.health.mockRejectedValue(new Error('down'));
    await expect(health.ready()).rejects.toThrow();
  });
});
