import { Inject, Module, OnModuleInit } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { APP_GUARD } from '@nestjs/core';
import { ClientKafka, ClientsModule } from '@nestjs/microservices';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ApiModule } from '@notopia-uit/api-document-nestjs-server';
import { OpenTelemetryModule } from 'nestjs-otel';
import { LoggerModule } from 'nestjs-pino';
import pretty from 'pino-pretty';

import { getKafkaConfig } from '#/config/kafka.config';

import { AuthorizationModule } from './authorization/authorization.module';
import { BlockNoteModule } from './blocknote/blocknote.module';
import { KAFKA_CLIENT } from './common/token';
import { HttpUserGuard } from './common/user.guard';
import { AppConfig } from './config/config';
import {
  APP_CONFIG,
  appConfig,
  databaseConfig,
  kafkaConfig,
  s3Config,
  servicesConfig,
} from './config/config.factory';
import { DatabaseModule } from './database/database.module';
import { DocumentApi } from './document/document.api';
import { DocumentEntity } from './document/document.entity';
import { DocumentService } from './document/document.service';
import { NoteModule } from './note/note.module';
import { RevisionApi } from './revision/revision.api';
import { RevisionEntity } from './revision/revision.entity';
import { RevisionService } from './revision/revision.service';
import { StorageModule } from './storage/storage.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      load: [appConfig, databaseConfig, servicesConfig, s3Config, kafkaConfig],
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
          },
        };
      },
    }),
    ClientsModule.registerAsync([
      {
        name: KAFKA_CLIENT,
        useFactory: getKafkaConfig,
        inject: [ConfigService],
      },
    ]),
    StorageModule,
    AuthorizationModule,
    NoteModule,
    DatabaseModule,
    BlockNoteModule,
    Object.assign(
      ApiModule.forRoot({
        apiImplementations: {
          documentApi: DocumentApi,
          revisionApi: RevisionApi,
        },
        providers: [
          DocumentService,
          RevisionService,
          {
            provide: KAFKA_CLIENT,
            useFactory: getKafkaConfig,
            inject: [ConfigService],
          },
        ],
      }),
      {
        imports: [
          ConfigModule,
          TypeOrmModule.forFeature([DocumentEntity, RevisionEntity]),
          DatabaseModule,
          AuthorizationModule,
          NoteModule,
          StorageModule,
          BlockNoteModule,
        ],
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
