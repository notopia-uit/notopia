import { z } from 'zod';

export const envSchema = z.object({
  NODE_ENV: z.enum(['test', 'production', 'development']).default('production'),
  NOTOPIA_DOCUMENT_LOG_LEVEL: z
    .enum(['trace', 'debug', 'info', 'warn', 'error', 'fatal'])
    .default('warn'),
  NOTOPIA_DOCUMENT_PORT: z.coerce.number().default(8082),
  NOTOPIA_DOCUMENT_API_URL: z.string().default(''),
  NOTOPIA_DOCUMENT_DB_HOST: z.string().default('localhost'),
  NOTOPIA_DOCUMENT_DB_PORT: z.coerce.number().default(5434),
  NOTOPIA_DOCUMENT_DB_USER: z.string().default('postgres'),
  NOTOPIA_DOCUMENT_DB_PASSWORD: z.string().default(''),
  NOTOPIA_DOCUMENT_DB_NAME: z.string().default('document'),
  NOTOPIA_DOCUMENT_SERVICES_NOTE_GRPC_URL: z.string().default(''),
  NOTOPIA_DOCUMENT_SERVICES_AUTHORIZATION_GRPC_URL: z.string().default(''),
  NOTOPIA_DOCUMENT_S3_ENDPOINT: z.string().default(''),
  NOTOPIA_DOCUMENT_S3_REGION: z.string().default('us-east-1'),
  NOTOPIA_DOCUMENT_S3_ACCESS_KEY_ID: z.string().default(''),
  NOTOPIA_DOCUMENT_S3_SECRET_ACCESS_KEY: z.string().default(''),
  NOTOPIA_DOCUMENT_S3_BUCKET_NAME: z.string().default('document'),
  NOTOPIA_DOCUMENT_KAFKA_CLIENT_ID: z.string().default('document'),
  NOTOPIA_DOCUMENT_KAFKA_BROKERS: z.string().default('localhost:19092'),
  NOTOPIA_DOCUMENT_KAFKA_GROUP_ID: z.string().default('document'),
  NOTOPIA_DOCUMENT_AUTHENTICATION_JWKS_URLS: z.string().default(''),
  NOTOPIA_DOCUMENT_AUTHENTICATION_ISSUERS: z.string().optional(),
  NOTOPIA_DOCUMENT_AUTHENTICATION_AUDIENCES: z.string().optional(),
});

export type Env = z.infer<typeof envSchema>;

export function validate(config: Record<string, unknown>) {
  return envSchema.parse(config);
}
