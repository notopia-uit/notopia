import {
  type DynamicModule,
  Inject,
  Logger,
  Module,
  type OnApplicationShutdown,
} from '@nestjs/common';
import { type Duration, LocksModule, LocksStorage } from '@nestjs/locks';
import Redis from 'ioredis';

import { RedisLockStore } from './redis-lock.store';

const logger = new Logger('LocksSetupModule');

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
        providers: [
          {
            provide: LOCKS_REDIS,
            useFactory: () => {
              logger.warn('locks: no redis url, using in-memory store (single instance only)');
              return null;
            },
          },
        ],
      };
    }
    return {
      module: LocksSetupModule,
      imports: [LocksModule.forRoot({ ttl })],
      providers: [
        {
          provide: LOCKS_REDIS,
          useFactory: () => {
            const client = new Redis(redisUrl);
            client.on('error', (err) => logger.error(`locks: redis error: ${err.message}`));
            return client;
          },
        },
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
