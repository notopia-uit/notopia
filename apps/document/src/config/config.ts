import { z } from 'zod';

export const appConfigSchema = z.object({
  env: z.enum(['test', 'production', 'development']),
  logLevel: z.enum(['trace', 'debug', 'info', 'warn', 'error', 'fatal']),
  port: z.number(),
  apiUrl: z.string(),
});
export type AppConfig = z.infer<typeof appConfigSchema>;

export const databaseConfigSchema = z.object({
  host: z.string(),
  port: z.number(),
  username: z.string(),
  password: z.string(),
  database: z.string(),
});
export type DatabaseConfig = z.infer<typeof databaseConfigSchema>;

export const servicesConfigSchema = z.object({
  noteUrl: z.string(),
  authorizationUrl: z.string(),
  billingUrl: z.string(),
});
export type ServicesConfig = z.infer<typeof servicesConfigSchema>;

export const s3ConfigSchema = z.object({
  endpoint: z.string(),
  region: z.string(),
  accessKeyId: z.string(),
  secretAccessKey: z.string(),
  bucketName: z.string(),
});
export type S3Config = z.infer<typeof s3ConfigSchema>;

export const kafkaConfigSchema = z.object({
  clientId: z.string(),
  brokers: z.array(z.string()),
  groupId: z.string(),
});
export type KafkaConfig = z.infer<typeof kafkaConfigSchema>;

export const authenticationConfigSchema = z.object({
  jwksUrls: z.array(z.string().url()).nonempty(),
  issuers: z.array(z.string()).optional(),
  audiences: z.array(z.string()).optional(),
});
export type AuthenticationConfig = z.infer<typeof authenticationConfigSchema>;

export const identityConfigSchema = z.object({
  adminGroup: z.string().min(1),
  serviceAccountUserId: z.string().min(1),
});
export type IdentityConfig = z.infer<typeof identityConfigSchema>;

export const hocuspocusConfigSchema = z.object({
  debounceMs: z.number().int().positive(),
  maxDebounceMs: z.number().int().positive(),
});
export type HocuspocusConfig = z.infer<typeof hocuspocusConfigSchema>;

export const autoCommitConfigSchema = z.object({
  checkFloorSeconds: z.number().int().nonnegative(),
});
export type AutoCommitConfig = z.infer<typeof autoCommitConfigSchema>;

export const cronJobConfigSchema = z.object({
  name: z.string().min(1),
  schedule: z.string().min(1),
  enabled: z.boolean(),
  timeoutSeconds: z.number().int().positive(),
});
export type CronJobConfig = z.infer<typeof cronJobConfigSchema>;

export const cronConfigSchema = z.object({
  sweeper: cronJobConfigSchema.extend({ idleSeconds: z.number().int().positive() }),
  revisionPurge: cronJobConfigSchema,
});
export type CronConfig = z.infer<typeof cronConfigSchema>;

export const redisConfigSchema = z.object({
  url: z.string().url().optional(),
});
export type RedisConfig = z.infer<typeof redisConfigSchema>;

export const scalingConfigSchema = z.object({
  enabled: z.boolean(),
});
export type ScalingConfig = z.infer<typeof scalingConfigSchema>;
