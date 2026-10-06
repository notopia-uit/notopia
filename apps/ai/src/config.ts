import { z } from 'zod';

export const appConfigSchema = z.object({
  env: z.enum(['test', 'production', 'development']),
  logLevel: z.enum(['trace', 'debug', 'info', 'warn', 'error', 'fatal']),
  port: z.number().int().min(1).max(65535),
});
export type AppConfig = z.infer<typeof appConfigSchema>;

export const llmConfigSchema = z.object({
  baseUrl: z.string().url('baseUrl must be a full URL including http:// or https://'),
  apiKey: z.string(),
  model: z.string().min(1),
});
export type LlmConfig = z.infer<typeof llmConfigSchema>;

export const servicesConfigSchema = z.object({
  noteGrpcUrl: z.string().min(1, 'noteGrpcUrl must not be empty'),
  authorizationGrpcUrl: z.string().min(1, 'authorizationGrpcUrl must not be empty'),
  meiliHost: z.string().url('meiliHost must be a full URL including http:// or https://'),
  meiliApiKey: z.string(),
});
export type ServicesConfig = z.infer<typeof servicesConfigSchema>;
