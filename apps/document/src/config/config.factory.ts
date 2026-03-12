import { registerAs } from '@nestjs/config';

import { AppConfig, DatabaseConfig } from './config';

export const appConfig = registerAs<AppConfig>(
  'app',
  (): AppConfig => ({
    env: process.env.NODE_ENV ?? 'production',
    logLevel: process.env.LOG_LEVEL ?? 'warn',
  })
);

export const databaseConfig = registerAs(
  'database',
  (): DatabaseConfig => ({
    host: process.env.DB_HOST ?? 'localhost',
    port: parseInt(process.env.DB_PORT ?? '5434', 10),
    username: process.env.DB_USER ?? 'postgres',
    password: process.env.DB_PASSWORD ?? '',
    database: process.env.DB_NAME ?? 'document',
  })
);
