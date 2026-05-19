import { z } from 'zod';

export const envSchema = z.object({
  NODE_ENV: z.enum(['test', 'production', 'development']).default('production'),
  NOTOPIA_SEARCH_WORKER_LOG_LEVEL: z
    .enum(['trace', 'debug', 'info', 'warn', 'error', 'fatal'])
    .default('warn'),
  NOTOPIA_SEARCH_WORKER_PORT: z.coerce.number().default(8084),
  NOTOPIA_SEARCH_WORKER_KAFKA_CLIENT_ID: z.string().default('search-worker'),
  NOTOPIA_SEARCH_WORKER_KAFKA_BROKERS: z.string().default('localhost:19092'),
  NOTOPIA_SEARCH_WORKER_KAFKA_GROUP_ID: z.string().default('search-worker'),
  NOTOPIA_SEARCH_WORKER_MEILI_HOST: z.string().default(''),
  NOTOPIA_SEARCH_WORKER_MEILI_API_KEY: z.string().default(''),
});

export type Env = z.infer<typeof envSchema>;

export function validate(config: Record<string, unknown>) {
  return envSchema.parse(config);
}
