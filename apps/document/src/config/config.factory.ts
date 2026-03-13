import { registerAs } from '@nestjs/config';

import { AppConfig, DatabaseConfig, S3Config, ServicesConfig } from './config';

export const appConfig = registerAs<AppConfig>(
  'app',
  (): AppConfig => ({
    env: process.env.NODE_ENV ?? 'production',
    logLevel: process.env.LOG_LEVEL ?? 'warn',
  })
);

export const getDatabaseConfig = (): DatabaseConfig => ({
  host: process.env.DB_HOST ?? 'localhost',
  port: parseInt(process.env.DB_PORT ?? '5434', 10),
  username: process.env.DB_USER ?? 'postgres',
  password: process.env.DB_PASSWORD ?? '',
  database: process.env.DB_NAME ?? 'document',
});

export const databaseConfig = registerAs('database', getDatabaseConfig);

export const servicesConfig = registerAs(
  'services',
  (): ServicesConfig => ({
    noteUrl: process.env.SERVICES_NOTE_URL ?? 'http://localhost:18081',
    authorizationUrl:
      process.env.SERVICES_AUTHORIZATION_URL ?? 'http://localhost:18089',
  })
);

export const s3Config = registerAs(
  's3',
  (): S3Config => ({
    endpoint: process.env.S3_ENDPOINT ?? '',
    region: process.env.S3_REGION ?? 'us-east-1',
    accessKeyId: process.env.S3_ACCESS_KEY_ID ?? '',
    secretAccessKey: process.env.S3_SECRET_ACCESS_KEY ?? '',
    bucketName: process.env.S3_BUCKET_NAME ?? 'document',
  })
);
