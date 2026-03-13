import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ApiModule } from '@notopia-uit/api-document-nestjs-server';
import { OpenTelemetryModule } from 'nestjs-otel';
import { LoggerModule } from 'nestjs-pino';
import pretty from 'pino-pretty';

import { AuthorizationModule } from './authorization/authorization.module';
import { AppConfig } from './config/config';
import {
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
import { S3Module } from './s3/s3.module';

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
        const appCfg = configService.get<AppConfig>('app')!;
        const level = appCfg.logLevel;
        const stream = pretty({ colorize: true, ignore: 'pid,hostname' });
        return {
          pinoHttp: {
            level,
            stream,
          },
        };
      },
    }),
    S3Module,
    AuthorizationModule,
    NoteModule,
    DatabaseModule,
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
          {
            provide: ServerBlockNoteEditor,
            useFactory: () => ServerBlockNoteEditor.create(),
          },
        ],
      }),
      {
        imports: [
          TypeOrmModule.forFeature([DocumentEntity, RevisionEntity]),
          DatabaseModule,
          AuthorizationModule,
          NoteModule,
          S3Module,
        ],
      }
    ),
  ],
})
export class AppModule {}
