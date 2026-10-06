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
  NOTOPIA_AI_MEILI_HOST: z.string().default(''),
  NOTOPIA_AI_MEILI_API_KEY: z.string().default(''),
});

export type Env = z.infer<typeof envSchema>;

export function validate(config: Record<string, unknown>) {
  return envSchema.parse(config);
}
