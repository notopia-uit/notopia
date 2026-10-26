import { z } from 'zod';

export const envSchema = z.object({
  NODE_ENV: z.enum(['test', 'production', 'development']).default('production'),
  NOTOPIA_AI_LOG_LEVEL: z
    .enum(['trace', 'debug', 'info', 'warn', 'error', 'fatal'])
    .default('warn'),
  NOTOPIA_AI_PORT: z.coerce.number().default(8085),
  NOTOPIA_AI_LLM_BASE_URL: z.string().default(''),
  NOTOPIA_AI_LLM_API_KEY: z.string().default(''),
  NOTOPIA_AI_LLM_MODEL: z.string().default('default'),
  NOTOPIA_AI_SERVICES_NOTE_GRPC_URL: z.string().default(''),
  NOTOPIA_AI_SERVICES_AUTHORIZATION_GRPC_URL: z.string().default(''),
  NOTOPIA_AI_SERVICES_BILLING_GRPC_URL: z.string().min(1),
  NOTOPIA_AI_REDIS_URL: z.string().url().optional(),
  NOTOPIA_AI_REQUESTS_PER_MINUTE: z.coerce.number().int().positive().default(20),
  NOTOPIA_AI_CONCURRENT_STREAMS: z.coerce.number().int().positive().default(2),
  NOTOPIA_AI_TIMEOUT_MS: z.coerce.number().int().min(1000).max(600000).default(120000),
  NOTOPIA_AI_MAX_OUTPUT_TOKENS: z.coerce.number().int().positive().default(4096),
  NOTOPIA_AI_MAX_STEPS: z.coerce.number().int().min(1).max(20).default(10),
  NOTOPIA_AI_MEILI_HOST: z.string().default(''),
  NOTOPIA_AI_MEILI_API_KEY: z.string().default(''),
});

export type Env = z.infer<typeof envSchema>;

export function validate(config: Record<string, unknown>) {
  return envSchema.parse(config);
}
