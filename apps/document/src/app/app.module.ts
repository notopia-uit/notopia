import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { OpenTelemetryModule } from 'nestjs-otel';
import { LoggerModule } from 'nestjs-pino';

import { AppConfig } from '../config/config';
import { appConfig } from '../config/config.factory';
import { DatabaseModule } from '../database/database.module';
import { HocuspocusModule } from '../hocuspocus/hocuspocus.module';
import { ApiHttpModule } from './api.module';

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
        const appCfg = configService.get<AppConfig>('app');
        const level = appCfg?.otel.log.level ?? 'info';
        return {
          pinoHttp: { level },
        };
      },
    }),
    DatabaseModule,
    HocuspocusModule,
    ApiHttpModule,
  ],
})
export class AppModule {}
