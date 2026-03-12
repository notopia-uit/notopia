import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { OpenTelemetryModule } from 'nestjs-otel';
import { LoggerModule } from 'nestjs-pino';
import pretty from 'pino-pretty';

import { ApiHttpModule } from './api.module';
import { AuthorizationModule } from './authorization/authorization.module';
import { AppConfig } from './config/config';
import { appConfig, databaseConfig } from './config/config.factory';
import { DatabaseModule } from './database/database.module';
import { DocumentModule } from './document/document.module';
import { NoteModule } from './note/note.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      load: [appConfig, databaseConfig],
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
    AuthorizationModule,
    NoteModule,
    DatabaseModule,
    DocumentModule,
    ApiHttpModule,
  ],
})
export class AppModule {}
