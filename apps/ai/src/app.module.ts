// oxfmt-ignore
import './otel';
// oxfmt-ignore
import 'reflect-metadata';

import { HttpException, Module, OnModuleInit } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { OpenTelemetryModule } from 'nestjs-otel';
import { Logger, LoggerModule } from 'nestjs-pino';
import pretty from 'pino-pretty';

import { AiModule } from './ai/ai.module';
import type { AppConfig, LlmConfig } from './config';
import { APP_CONFIG, LLM_CONFIG, appConfig, llmConfig, servicesConfig } from './config.factory';
import { validate } from './env.validation';
import { HealthController } from './health.controller';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      validate,
      load: [appConfig, llmConfig, servicesConfig],
    }),
    OpenTelemetryModule.forRoot({
      metrics: {
        hostMetrics: true,
      },
    }),
    LoggerModule.forRootAsync({
      imports: [ConfigModule],
      inject: [ConfigService],
      useFactory: (configService: ConfigService) => {
        const appCfg = configService.get<AppConfig>(APP_CONFIG);
        if (!appCfg) {
          throw new Error('APP_CONFIG not found');
        }
        return {
          pinoHttp: {
            level: appCfg.logLevel,
            stream:
              appCfg.env !== 'production'
                ? pretty({ colorize: true, ignore: 'pid,hostname' })
                : undefined,
            serializers: {
              err: (err: unknown) => {
                if (err instanceof HttpException) {
                  return {
                    type: err.name,
                    status: err.getStatus(),
                    message: err.message,
                    stack: err.stack,
                    cause: err.cause,
                  };
                }
                if (err instanceof Error) {
                  return {
                    type: err.name,
                    message: err.message,
                    stack: err.stack,
                    cause: (err as Error & { cause?: unknown }).cause,
                  };
                }
                return { message: String(err) };
              },
            },
          },
        };
      },
    }),
    AiModule,
  ],
  controllers: [HealthController],
})
export class AppModule implements OnModuleInit {
  constructor(
    private readonly configService: ConfigService,
    private readonly logger: Logger
  ) {}

  onModuleInit() {
    this.logger.log(
      {
        app: this.configService.get(APP_CONFIG),
        llm: {
          ...this.configService.get(LLM_CONFIG),
          apiKey: '***',
        } satisfies Partial<LlmConfig>,
      },
      'Application configuration loaded'
    );
  }
}
