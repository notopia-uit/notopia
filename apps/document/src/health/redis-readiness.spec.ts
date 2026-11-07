import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { Test } from '@nestjs/testing';
import { DataSource } from 'typeorm';
import { expect, it, vi } from 'vitest';

import { REDIS_CONFIG } from '../config';
import { DatabaseModule } from '../database';
import { HEALTH_REDIS } from './health.controller';
import { HealthModule } from './health.module';

const redis = vi.hoisted(() => ({
  ping: vi.fn(() => new Promise<never>(() => {})),
  disconnect: vi.fn<() => void>(),
}));
vi.mock('ioredis', () => ({
  default: class {
    ping = redis.ping;
    disconnect = redis.disconnect;
  },
}));

@Module({
  providers: [{ provide: DataSource, useValue: {} }],
  exports: [DataSource],
})
class FakeDatabaseModule {}

it('times out and disconnects when Redis does not answer', async () => {
  const module = await Test.createTestingModule({
    imports: [ConfigModule.forRoot({ isGlobal: true }), HealthModule],
  })
    .overrideModule(DatabaseModule)
    .useModule(FakeDatabaseModule)
    .overrideProvider(ConfigService)
    .useValue({
      getOrThrow: (key: symbol) =>
        key === REDIS_CONFIG ? { url: 'redis://localhost' } : { enabled: true },
    })
    .compile();
  vi.useFakeTimers();
  try {
    const probe = module.get<{ ping(): Promise<unknown> }>(HEALTH_REDIS).ping();
    const result = expect(probe).rejects.toThrow('Redis health check timed out');
    await vi.advanceTimersByTimeAsync(2000);
    await result;
    expect(redis.ping).toHaveBeenCalledOnce();
    expect(redis.disconnect).toHaveBeenCalledOnce();
  } finally {
    vi.useRealTimers();
    await module.close();
  }
});
