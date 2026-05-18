import { HttpException, Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { createSchema } from '@notopia-uit/lib-server/block-note';
import { Meilisearch } from 'meilisearch';
import { OpenTelemetryModule } from 'nestjs-otel';
import { LoggerModule } from 'nestjs-pino';
import pretty from 'pino-pretty';
import { BLOCKNOTE_SCHEMA } from 'token';

import { AppController } from './app.controller';
import { AppService } from './app.service';
import { AppConfig, MeiliConfig } from './config';
import { APP_CONFIG, MEILI_CONFIG, appConfig, kafkaConfig, meiliConfig } from './config.factory';
import { validate } from './env.validation';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      validate,
      load: [appConfig, kafkaConfig, meiliConfig],
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
        const level = appCfg.logLevel;
        const stream = pretty({ colorize: true, ignore: 'pid,hostname' });
        return {
          pinoHttp: {
            level,
            stream,
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
  ],
  providers: [
    {
      provide: Meilisearch,
      useFactory: (configService: ConfigService) => {
        const config = configService.get<MeiliConfig>(MEILI_CONFIG);
        if (!config) {
          throw new Error('MEILI_CONFIG not found');
        }
        return new Meilisearch({
          host: config.host,
          apiKey: config.apiKey,
        });
      },
      inject: [ConfigService],
    },
    {
      provide: BLOCKNOTE_SCHEMA,
      useFactory: createSchema,
    },
    AppService,
  ],
  controllers: [AppController],
})
export class AppModule {}
