import { z } from 'zod';

export const appConfigSchema = z.object({
  env: z.enum(['test', 'production', 'development']),
  logLevel: z.enum(['trace', 'debug', 'info', 'warn', 'error', 'fatal']),
  port: z.number(),
});
export type AppConfig = z.infer<typeof appConfigSchema>;

export const kafkaConfigSchema = z.object({
  clientId: z.string(),
  brokers: z.array(z.string()),
  groupId: z.string(),
});
export type KafkaConfig = z.infer<typeof kafkaConfigSchema>;

export const meiliConfigSchema = z.object({
  host: z.string(),
  apiKey: z.string(),
});
export type MeiliConfig = z.infer<typeof meiliConfigSchema>;
