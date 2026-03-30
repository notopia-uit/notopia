import { AppConfig, DatabaseConfig, S3Config, ServicesConfig } from './config';
import { registerAs } from '@nestjs/config';

export const APP_CONFIG = Symbol('APP_CONFIG');

export const appConfig = registerAs<AppConfig>(
  APP_CONFIG,
  (): AppConfig => ({
    env: process.env.NODE_ENV ?? 'production',
    logLevel: process.env.NOTOPIA_DOCUMENT_LOG_LEVEL ?? 'warn',
    port: parseInt(process.env.NOTOPIA_DOCUMENT_PORT ?? '8082', 10),
    apiUrl: process.env.NOTOPIA_DOCUMENT_API_URL ?? '',
  })
);

export const getDatabaseConfig = (): DatabaseConfig => ({
  host: process.env.NOTOPIA_DOCUMENT_DB_HOST ?? 'localhost',
  port: parseInt(process.env.NOTOPIA_DOCUMENT_DB_PORT ?? '5434', 10),
  username: process.env.NOTOPIA_DOCUMENT_DB_USER ?? 'postgres',
  password: process.env.NOTOPIA_DOCUMENT_DB_PASSWORD ?? '',
  database: process.env.NOTOPIA_DOCUMENT_DB_NAME ?? 'document',
});

export const DATABASE_CONFIG = Symbol('DATABASE_CONFIG');

export const databaseConfig = registerAs(DATABASE_CONFIG, getDatabaseConfig);

export const SERVICE_CONFIG = Symbol('SERVICE_CONFIG');

export const servicesConfig = registerAs(
  SERVICE_CONFIG,
  (): ServicesConfig => ({
    noteUrl: process.env.NOTOPIA_DOCUMENT_SERVICES_NOTE_CONNECTRPC_URL ?? '',
    authorizationUrl:
      process.env.NOTOPIA_DOCUMENT_SERVICES_AUTHORIZATION_CONNECTRPC_URL ?? '',
  })
);

export const S3_CONFIG = Symbol('S3_CONFIG');

export const s3Config = registerAs(
  S3_CONFIG,
  (): S3Config => ({
    endpoint: process.env.NOTOPIA_DOCUMENT_S3_ENDPOINT ?? '',
    region: process.env.NOTOPIA_DOCUMENT_S3_REGION ?? 'us-east-1',
    accessKeyId: process.env.NOTOPIA_DOCUMENT_S3_ACCESS_KEY_ID ?? '',
    secretAccessKey: process.env.NOTOPIA_DOCUMENT_S3_SECRET_ACCESS_KEY ?? '',
    bucketName: process.env.NOTOPIA_DOCUMENT_S3_BUCKET_NAME ?? 'document',
  })
);
