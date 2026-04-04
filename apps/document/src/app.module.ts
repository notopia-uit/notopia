import { AuthorizationModule } from './authorization/authorization.module';
import { BlockNoteModule } from './blocknote/blocknote.module';
import { HttpUserGuard } from './common/user.guard';
import { AppConfig } from './config/config';
import {
  APP_CONFIG,
  appConfig,
  databaseConfig,
  s3Config,
  servicesConfig,
} from './config/config.factory';
import { DatabaseModule } from './database/database.module';
import { DocumentApi } from './document/document.api';
import { DocumentEntity } from './document/document.entity';
import { DocumentRepository } from './document/document.repository';
import { DocumentService } from './document/document.service';
import { NoteModule } from './note/note.module';
import { RevisionApi } from './revision/revision.api';
import { RevisionEntity } from './revision/revision.entity';
import { RevisionRepository } from './revision/revision.repository';
import { RevisionService } from './revision/revision.service';
import { StorageModule } from './storage/storage.module';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { APP_GUARD } from '@nestjs/core';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ApiModule } from '@notopia-uit/api-document-nestjs-server';
import { OpenTelemetryModule } from 'nestjs-otel';
import { LoggerModule } from 'nestjs-pino';
import pretty from 'pino-pretty';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      load: [appConfig, databaseConfig, servicesConfig, s3Config],
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
        const appCfg = configService.get<AppConfig>(APP_CONFIG)!;
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
          DocumentRepository,
          DocumentService,
          RevisionRepository,
          RevisionService,
        ],
      }),
      {
        imports: [
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
export class AppModule {}
