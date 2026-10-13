import { registerAs } from '@nestjs/config';

import {
  AppConfig,
  appConfigSchema,
  KafkaConfig,
  kafkaConfigSchema,
  MeiliConfig,
  meiliConfigSchema,
  TombstoneConfig,
  tombstoneConfigSchema,
} from './config';

export const APP_CONFIG = Symbol('APP_CONFIG');

export const appConfig = registerAs(APP_CONFIG, () =>
  appConfigSchema.parse({
    env: (process.env.NODE_ENV as AppConfig['env'] | undefined) ?? 'production',
    logLevel:
      (process.env.NOTOPIA_SEARCH_WORKER_LOG_LEVEL as AppConfig['logLevel'] | undefined) ?? 'warn',
    port: parseInt(process.env.NOTOPIA_SEARCH_WORKER_PORT ?? '8083', 10),
  } satisfies AppConfig)
);

export const KAFKA_CONFIG = Symbol('KAFKA_CONFIG');

export const kafkaConfig = registerAs(KAFKA_CONFIG, () =>
  kafkaConfigSchema.parse({
    clientId: process.env.NOTOPIA_SEARCH_WORKER_KAFKA_CLIENT_ID ?? 'search-worker',
    brokers: process.env
      .NOTOPIA_SEARCH_WORKER_KAFKA_BROKERS!.split(',')
      .map((v) => v.trim())
      .filter(Boolean),
    groupId: process.env.NOTOPIA_SEARCH_WORKER_KAFKA_GROUP_ID ?? 'search-worker',
  } satisfies KafkaConfig)
);

export const MEILI_CONFIG = Symbol('MEILI_CONFIG');

export const meiliConfig = registerAs(MEILI_CONFIG, () =>
  meiliConfigSchema.parse({
    host: process.env.NOTOPIA_SEARCH_WORKER_MEILI_HOST!,
    apiKey: process.env.NOTOPIA_SEARCH_WORKER_MEILI_API_KEY!,
  } satisfies MeiliConfig)
);

export const TOMBSTONE_CONFIG = Symbol('TOMBSTONE_CONFIG');

export const tombstoneConfig = registerAs(TOMBSTONE_CONFIG, () =>
  tombstoneConfigSchema.parse({
    redisUrl: process.env.NOTOPIA_SEARCH_WORKER_REDIS_URL || undefined,
    ttlSeconds: parseInt(process.env.NOTOPIA_SEARCH_WORKER_TOMBSTONE_TTL_SECONDS ?? '86400', 10),
    max: parseInt(process.env.NOTOPIA_SEARCH_WORKER_TOMBSTONE_MAX ?? '100000', 10),
  } satisfies TombstoneConfig)
);
