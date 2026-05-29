import { HttpException, Inject, Module, OnModuleInit } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { APP_GUARD } from '@nestjs/core';
import { ClientKafka } from '@nestjs/microservices';
import { ApiModule } from '@notopia-uit/api-document-nestjs-server';
import { OpenTelemetryModule } from 'nestjs-otel';
import { LoggerModule } from 'nestjs-pino';
import pretty from 'pino-pretty';

import { AuthenticationModule } from './authentication';
import { AuthorizationModule } from './authorization';
import { BlockNoteModule } from './blocknote';
import { HttpUserGuard } from './common';
import {
  APP_CONFIG,
  AppConfig,
  appConfig,
  authenticationConfig,
  databaseConfig,
  kafkaConfig,
  s3Config,
  servicesConfig,
  validate,
} from './config';
import { DatabaseModule } from './database';
import { DocumentApi, DocumentModule } from './document';
import { HocuspocusModule } from './hocuspocus';
import { KAFKA_CLIENT, KafkaModule } from './kafka';
import { NoteModule } from './note';
import { RevisionApi, RevisionModule } from './revision';
import { StorageModule } from './storage';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      validate,
      load: [
        appConfig,
        databaseConfig,
        servicesConfig,
        s3Config,
        kafkaConfig,
        authenticationConfig,
      ],
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

    StorageModule,
    KafkaModule,
    AuthenticationModule,
    AuthorizationModule,
    NoteModule,
    DatabaseModule,
    BlockNoteModule,
    HocuspocusModule,
    DocumentModule,
    RevisionModule,
    Object.assign(
      ApiModule.forRoot({
        apiImplementations: {
          documentApi: DocumentApi,
          revisionApi: RevisionApi,
        },
      }),
      {
        imports: [ConfigModule, DocumentModule, HocuspocusModule, RevisionModule],
      }
    ),
  ],
  providers: [
    {
      provide: APP_GUARD,
      useClass: HttpUserGuard,
    },
  ],
})
export class AppModule implements OnModuleInit {
  // We should refactor into bus or so, this is a mess for app module
  constructor(@Inject(KAFKA_CLIENT) private readonly kafkaClient: ClientKafka) {}

  async onModuleInit() {
    await this.kafkaClient.connect();
  }
}
