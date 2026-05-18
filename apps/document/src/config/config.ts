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
  jwksUrls: z.array(z.string()),
  issuers: z.array(z.string()).optional(),
  audiences: z.array(z.string()).optional(),
});
export type AuthenticationConfig = z.infer<typeof authenticationConfigSchema>;
