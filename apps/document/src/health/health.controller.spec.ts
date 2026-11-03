import { TerminusModule } from '@nestjs/terminus';
import { Test } from '@nestjs/testing';
import { DataSource } from 'typeorm';
import { describe, expect, it, vi } from 'vitest';

import { SCALING_CONFIG } from '../config';
import { HEALTH_REDIS, HealthController } from './health.controller';

describe('health checks', () => {
  const setup = (scaling = false) => {
    const dataSource = {
      query: vi.fn().mockResolvedValue([]),
      showMigrations: vi.fn().mockResolvedValue(false),
    };
    const redis = { ping: vi.fn().mockResolvedValue('PONG') };
    const health = {
      check: async (checks: Array<() => Promise<Record<string, { status: string }>>>) => {
        const results = await Promise.all(checks.map((check) => check()));
        if (
          results.some((result) => Object.values(result).some((value) => value.status === 'down'))
        )
          throw new Error('unhealthy');
        return results;
      },
    };
    const controller = new HealthController(
      health as never,
      dataSource as never,
      { enabled: scaling },
      redis as never
    );
    return { controller, dataSource, redis };
  };

  it('reports live without querying dependencies', async () => {
    const { controller, dataSource } = setup();
    await expect(controller.live()).resolves.toBeDefined();
    expect(dataSource.query).not.toHaveBeenCalled();
  });

  it('fails readiness when postgres fails', async () => {
    const { controller, dataSource } = setup();
    dataSource.query.mockRejectedValue(new Error('postgres unavailable'));
    await expect(controller.ready()).rejects.toThrow();
  });

  it('fails readiness when scaling redis fails', async () => {
    const { controller, redis } = setup(true);
    redis.ping.mockRejectedValue(new Error('redis unavailable'));
    await expect(controller.ready()).rejects.toThrow();
  });

  it('fails startup when migrations are pending', async () => {
    const { controller, dataSource } = setup();
    dataSource.showMigrations.mockResolvedValue(true);
    await expect(controller.startup()).rejects.toThrow();
  });
});

it('serves Terminus health status over HTTP', async () => {
  const database = {
    query: vi.fn().mockResolvedValue([]),
    showMigrations: vi.fn().mockResolvedValue(true),
  };
  const module = await Test.createTestingModule({
    imports: [TerminusModule],
    controllers: [HealthController],
    providers: [
      { provide: DataSource, useValue: database },
      { provide: SCALING_CONFIG, useValue: { enabled: false } },
      { provide: HEALTH_REDIS, useValue: { ping: vi.fn() } },
    ],
  }).compile();
  const app = module.createNestApplication();
  await app.listen(0, '127.0.0.1');
  try {
    const url = await app.getUrl();
    expect((await fetch(`${url}/document/health/live`)).status).toBe(200);
    expect((await fetch(`${url}/document/health/ready`)).status).toBe(200);
    expect((await fetch(`${url}/document/health/startup`)).status).toBe(503);
    database.query.mockRejectedValue(new Error('offline'));
    expect((await fetch(`${url}/document/health/ready`)).status).toBe(503);
  } finally {
    await app.close();
  }
});
