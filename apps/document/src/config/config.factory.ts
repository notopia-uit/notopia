import { registerAs } from '@nestjs/config';

import {
  AppConfig,
  AuthenticationConfig,
  DatabaseConfig,
  KafkaConfig,
  S3Config,
  ServicesConfig,
} from './config';

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
    noteUrl: process.env.NOTOPIA_DOCUMENT_SERVICES_NOTE_GPRC_URL ?? '',
    authorizationUrl: process.env.NOTOPIA_DOCUMENT_SERVICES_AUTHORIZATION_GRPC_URL ?? '',
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

export const KAFKA_CONFIG = Symbol('KAFKA_CONFIG');

export const kafkaConfig = registerAs(
  KAFKA_CONFIG,
  (): KafkaConfig => ({
    clientId: process.env.NOTOPIA_DOCUMENT_KAFKA_CLIENT_ID ?? 'document',
    brokers: (process.env.NOTOPIA_DOCUMENT_KAFKA_BROKERS ?? 'localhost:19092').split(','),
    groupId: process.env.NOTOPIA_DOCUMENT_KAFKA_GROUP_ID ?? 'document',
  })
);

export const AUTHENTICATION_CONFIG = Symbol('AUTHENTICATION_CONFIG');

export const authenticationConfig = registerAs(
  AUTHENTICATION_CONFIG,
  (): AuthenticationConfig => ({
    jwksUrls: process.env.NOTOPIA_DOCUMENT_AUTHENTICATION_JWKS_URLS?.split(',') ?? [],
    issuers: process.env.NOTOPIA_DOCUMENT_AUTHENTICATION_ISSUERS?.split(','),
    audiences: process.env.NOTOPIA_DOCUMENT_AUTHENTICATION_AUDIENCES?.split(','),
  })
);
