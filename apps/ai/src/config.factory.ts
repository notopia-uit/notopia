import { registerAs } from '@nestjs/config';

import {
  AppConfig,
  appConfigSchema,
  LlmConfig,
  llmConfigSchema,
  ServicesConfig,
  servicesConfigSchema,
} from './config';

export const APP_CONFIG = Symbol('APP_CONFIG');

export const LLM_CONFIG = Symbol('LLM_CONFIG');

export const SERVICES_CONFIG = Symbol('SERVICES_CONFIG');

// NOTE: ConfigService.get() needs the raw Symbol token (internalConfig is keyed
// by it), while @Inject() needs the factory KEY (`CONFIGURATION(...)` string).
// So: configService.get(SERVICES_CONFIG) but @Inject(servicesConfig.KEY).

export const appConfig = registerAs(APP_CONFIG, () =>
  appConfigSchema.parse({
    env: (process.env.NODE_ENV as AppConfig['env'] | undefined) ?? 'production',
    logLevel: (process.env.NOTOPIA_AI_LOG_LEVEL as AppConfig['logLevel'] | undefined) ?? 'warn',
    port: parseInt(process.env.NOTOPIA_AI_PORT ?? '8085', 10),
  } satisfies AppConfig)
);

export const llmConfig = registerAs(LLM_CONFIG, () =>
  llmConfigSchema.parse({
    baseUrl: process.env.NOTOPIA_AI_LLM_BASE_URL!,
    apiKey: process.env.NOTOPIA_AI_LLM_API_KEY!,
    model: process.env.NOTOPIA_AI_LLM_MODEL ?? 'default',
  } satisfies LlmConfig)
);

export const servicesConfig = registerAs(SERVICES_CONFIG, () =>
  servicesConfigSchema.parse({
    noteGrpcUrl: process.env.NOTOPIA_AI_SERVICES_NOTE_GRPC_URL!,
    authorizationGrpcUrl: process.env.NOTOPIA_AI_SERVICES_AUTHORIZATION_GRPC_URL!,
    meiliHost: process.env.NOTOPIA_AI_MEILI_HOST!,
    meiliApiKey: process.env.NOTOPIA_AI_MEILI_API_KEY ?? '',
  } satisfies ServicesConfig)
);
