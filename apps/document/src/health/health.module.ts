import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { TerminusModule } from '@nestjs/terminus';
import Redis from 'ioredis';

import { REDIS_CONFIG, RedisConfig, SCALING_CONFIG } from '../config';
import { DatabaseModule } from '../database';
import { HEALTH_REDIS, HealthController } from './health.controller';

@Module({
  imports: [TerminusModule, DatabaseModule],
  controllers: [HealthController],
  providers: [
    {
      provide: SCALING_CONFIG,
      inject: [ConfigService],
      useFactory: (config: ConfigService) => config.getOrThrow(SCALING_CONFIG),
    },
    {
      provide: HEALTH_REDIS,
      inject: [ConfigService],
      useFactory: (config: ConfigService) => ({
        async ping() {
          const url = config.getOrThrow<RedisConfig>(REDIS_CONFIG).url;
          if (!url) throw new Error('Redis URL is missing');
          const redis = new Redis(url, {
            lazyConnect: true,
            connectTimeout: 2000,
            maxRetriesPerRequest: 1,
            retryStrategy: () => null,
          });
          try {
            return await redis.ping();
          } finally {
            redis.disconnect();
          }
        },
      }),
    },
  ],
})
export class HealthModule {}
