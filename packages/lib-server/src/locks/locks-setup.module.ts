import { type DynamicModule, Inject, Module, type OnApplicationShutdown } from '@nestjs/common';
import { type Duration, LocksModule, LocksStorage } from '@nestjs/locks';
import Redis from 'ioredis';

import { RedisLockStore } from './redis-lock.store';

const LOCKS_REDIS = Symbol('LOCKS_REDIS');

export interface LocksSetupOptions {
  redisUrl?: string;
  ttl?: Duration;
}

@Module({})
export class LocksSetupModule implements OnApplicationShutdown {
  constructor(@Inject(LOCKS_REDIS) private readonly redis: Redis | null) {}

  static forRoot({ redisUrl, ttl = '30s' }: LocksSetupOptions): DynamicModule {
    if (!redisUrl) {
      return {
        module: LocksSetupModule,
        imports: [LocksModule.forRoot({ ttl, allowInMemoryStorage: true })],
        providers: [{ provide: LOCKS_REDIS, useValue: null }],
      };
    }
    return {
      module: LocksSetupModule,
      imports: [LocksModule.forRoot({ ttl })],
      providers: [
        { provide: LOCKS_REDIS, useFactory: () => new Redis(redisUrl) },
        {
          provide: RedisLockStore,
          inject: [LOCKS_REDIS, LocksStorage],
          useFactory: (redis: Redis, storage: LocksStorage) => new RedisLockStore(redis, storage),
        },
      ],
    };
  }

  async onApplicationShutdown() {
    await this.redis?.quit();
  }
}
