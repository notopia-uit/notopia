import { TerminusModule } from '@nestjs/terminus';
import { Test } from '@nestjs/testing';
import { DataSource } from 'typeorm';
import { expect, it, vi } from 'vitest';

import { SCALING_CONFIG } from '../config';
import { HEALTH_REDIS, HealthController } from './health.controller';

it('serves Terminus health status over HTTP', async () => {
  const database = {
    query: vi.fn().mockResolvedValue([]),
    showMigrations: vi.fn().mockResolvedValue(true),
  };
  const redis = { ping: vi.fn().mockResolvedValue('PONG') };
  const module = await Test.createTestingModule({
    imports: [TerminusModule],
    controllers: [HealthController],
    providers: [
      { provide: DataSource, useValue: database },
      { provide: SCALING_CONFIG, useValue: { enabled: true } },
      { provide: HEALTH_REDIS, useValue: redis },
    ],
  }).compile();
  const app = module.createNestApplication();
  await app.listen(0, '127.0.0.1');
  try {
    const url = await app.getUrl();
    expect((await fetch(`${url}/document/health/live`)).status).toBe(200);
    expect((await fetch(`${url}/document/health/ready`)).status).toBe(200);
    expect((await fetch(`${url}/document/health/startup`)).status).toBe(503);
    redis.ping.mockRejectedValue(new Error('offline'));
    expect((await fetch(`${url}/document/health/ready`)).status).toBe(503);
    redis.ping.mockResolvedValue('PONG');
    database.query.mockRejectedValue(new Error('offline'));
    expect((await fetch(`${url}/document/health/ready`)).status).toBe(503);
    expect((await fetch(`${url}/document/health/live`)).status).toBe(200);
  } finally {
    await app.close();
  }
});
