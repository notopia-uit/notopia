import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { OpenTelemetryModule } from 'nestjs-otel';
import { LoggerModule } from 'nestjs-pino';

import { AppConfig } from '../config/config';
import { appConfig } from '../config/config.factory';
import { DocumentModule } from '../document/document.module';
import { DatabaseModule } from '../infrastructure/database/database.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      load: [appConfig],
    }),
    LoggerModule.forRootAsync({
      imports: [ConfigModule],
      inject: [ConfigService],
      useFactory: (configService: ConfigService<AppConfig, true>) => {
        const level =
          configService.get('otel', { infer: true }).log.level ?? 'info';
        return {
          pinoHttp: { level },
        };
      },
    }),
    OpenTelemetryModule.forRoot({
      metrics: {
        hostMetrics: true,
        apiMetrics: {
          enable: true,
        },
      },
    }),
    DatabaseModule,
    DocumentModule,
  ],
})
export class AppModule {}
