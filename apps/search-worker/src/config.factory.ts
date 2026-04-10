import { registerAs } from '@nestjs/config';

import { AppConfig, KafkaConfig, MeiliConfig } from './config';

export const APP_CONFIG = Symbol('APP_CONFIG');

export const appConfig = registerAs<AppConfig>(
  APP_CONFIG,
  (): AppConfig => ({
    env: process.env.NODE_ENV ?? 'production',
    logLevel: process.env.NOTOPIA_SEARCH_WORKER_LOG_LEVEL ?? 'warn',
    port: parseInt(process.env.NOTOPIA_SEARCH_WORKER_PORT ?? '8084', 10),
  })
);

export const KAFKA_CONFIG = Symbol('KAFKA_CONFIG');

export const kafkaConfig = registerAs(
  KAFKA_CONFIG,
  (): KafkaConfig => ({
    clientId:
      process.env.NOTOPIA_SEARCH_WORKER_KAFKA_CLIENT_ID ?? 'search-worker',
    brokers: (
      process.env.NOTOPIA_SEARCH_WORKER_KAFKA_BROKERS ?? 'localhost:19092'
    ).split(','),
    groupId:
      process.env.NOTOPIA_SEARCH_WORKER_KAFKA_GROUP_ID ?? 'search-worker',
  })
);

export const MEILI_CONFIG = Symbol('MEILI_CONFIG');

export const meiliConfig = registerAs(
  MEILI_CONFIG,
  (): MeiliConfig => ({
    host: process.env.NOTOPIA_SEARCH_WORKER_MEILI_HOST ?? '',
    apiKey: process.env.NOTOPIA_SEARCH_WORKER_MEILI_API_KEY ?? '',
  })
);
