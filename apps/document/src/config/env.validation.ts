import { z } from 'zod';

const booleanString = z.enum(['true', 'false']);

export const envSchema = z
  .object({
    NODE_ENV: z.enum(['test', 'production', 'development']).default('production'),
    NOTOPIA_DOCUMENT_LOG_LEVEL: z
      .enum(['trace', 'debug', 'info', 'warn', 'error', 'fatal'])
      .default('warn'),
    NOTOPIA_DOCUMENT_PORT: z.coerce.number().int().min(1).max(65535).default(8082),
    NOTOPIA_DOCUMENT_API_URL: z.string().default(''),
    NOTOPIA_DOCUMENT_DB_HOST: z.string().default('localhost'),
    NOTOPIA_DOCUMENT_DB_PORT: z.coerce.number().int().min(1).max(65535).default(5434),
    NOTOPIA_DOCUMENT_DB_USER: z.string().default('postgres'),
    NOTOPIA_DOCUMENT_DB_PASSWORD: z.string().default(''),
    NOTOPIA_DOCUMENT_DB_NAME: z.string().default('document'),
    NOTOPIA_DOCUMENT_SERVICES_NOTE_GRPC_URL: z.string().default(''),
    NOTOPIA_DOCUMENT_SERVICES_AUTHORIZATION_GRPC_URL: z.string().default(''),
    NOTOPIA_DOCUMENT_SERVICES_BILLING_GRPC_URL: z.string().default(''),
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
    NOTOPIA_DOCUMENT_HOCUSPOCUS_DEBOUNCE_MS: z.coerce.number().int().positive().default(2000),
    NOTOPIA_DOCUMENT_HOCUSPOCUS_MAX_DEBOUNCE_MS: z.coerce.number().int().positive().default(10000),
    NOTOPIA_DOCUMENT_AUTO_COMMIT_CHECK_FLOOR_SECONDS: z.coerce
      .number()
      .int()
      .nonnegative()
      .default(60),
    NOTOPIA_DOCUMENT_IDENTITY_ADMIN_GROUP: z.string().default('notopia-admins'),
    NOTOPIA_DOCUMENT_IDENTITY_SERVICE_ACCOUNT_USER_ID: z.string().min(1),
    NOTOPIA_DOCUMENT_CRON_SWEEPER_NAME: z.string().min(1).default('document:sweeper'),
    NOTOPIA_DOCUMENT_CRON_SWEEPER_SCHEDULE: z.string().min(1).default('*/10 * * * *'),
    NOTOPIA_DOCUMENT_CRON_SWEEPER_ENABLED: booleanString.default('true'),
    NOTOPIA_DOCUMENT_CRON_SWEEPER_TIMEOUT_SECONDS: z.coerce.number().int().positive().default(300),
    NOTOPIA_DOCUMENT_CRON_SWEEPER_IDLE_SECONDS: z.coerce.number().int().positive().default(600),
    NOTOPIA_DOCUMENT_CRON_REVISION_PURGE_NAME: z.string().min(1).default('document:revision-purge'),
    NOTOPIA_DOCUMENT_CRON_REVISION_PURGE_SCHEDULE: z.string().min(1).default('30 3 * * *'),
    NOTOPIA_DOCUMENT_CRON_REVISION_PURGE_ENABLED: booleanString.default('true'),
    NOTOPIA_DOCUMENT_CRON_REVISION_PURGE_TIMEOUT_SECONDS: z.coerce
      .number()
      .int()
      .positive()
      .default(1800),
    NOTOPIA_DOCUMENT_REDIS_URL: z.preprocess(
      (v) => (v === '' ? undefined : v),
      z.string().url().optional()
    ),
    NOTOPIA_DOCUMENT_SCALING_ENABLED: booleanString.default('false'),
  })
  .superRefine((env, ctx) => {
    if (env.NOTOPIA_DOCUMENT_SCALING_ENABLED === 'true' && !env.NOTOPIA_DOCUMENT_REDIS_URL) {
      ctx.addIssue({
        code: 'custom',
        path: ['NOTOPIA_DOCUMENT_REDIS_URL'],
        message:
          'NOTOPIA_DOCUMENT_REDIS_URL is required when NOTOPIA_DOCUMENT_SCALING_ENABLED=true',
      });
    }
  });

export type Env = z.infer<typeof envSchema>;

export function validate(config: Record<string, unknown>) {
  return envSchema.parse(config);
}
