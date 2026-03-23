import { ConfigService } from '@nestjs/config';
import { NestFactory } from '@nestjs/core';
import { WsAdapter } from '@nestjs/platform-ws';
import { Logger } from 'nestjs-pino';

import { AppModule } from './app.module';
import { GlobalExceptionFilter } from './common/http-exception.filter';
import { AppConfig } from './config/config';
import { APP_CONFIG } from './config/config.factory';
import { otelSdk } from './otel';

otelSdk.start();

async function bootstrap() {
  const app = await NestFactory.create(AppModule, { bufferLogs: true });
  app.useWebSocketAdapter(new WsAdapter(app));
  const logger = app.get(Logger);
  const configService = app.get(ConfigService);
  app.useLogger(logger);
  app.useGlobalFilters(new GlobalExceptionFilter(logger));
  const port = configService.get<AppConfig>(APP_CONFIG)!.port;
  await app.listen(port);
}

bootstrap();
