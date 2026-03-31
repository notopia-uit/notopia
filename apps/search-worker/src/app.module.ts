import { AppController } from './app.controller';
import { AppService } from './app.service';
import { AppConfig, MeiliConfig } from './config';
import { APP_CONFIG, MEILI_CONFIG, appConfig } from './config.factory';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { MeiliSearch } from 'meilisearch';
import { OpenTelemetryModule } from 'nestjs-otel';
import { LoggerModule } from 'nestjs-pino';
import pretty from 'pino-pretty';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      load: [appConfig],
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
  ],
  providers: [
    {
      provide: MeiliSearch,
      useFactory: (configService: ConfigService) => {
        const config = configService.get<MeiliConfig>(MEILI_CONFIG)!;
        return new MeiliSearch({
          host: config.host,
          apiKey: config.apiKey,
        });
      },
      inject: [ConfigService],
    },
    AppService,
    AppController,
  ],
})
export class AppModule {}
