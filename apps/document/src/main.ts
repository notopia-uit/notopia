// oxfmt-ignore
import './otel';
// oxfmt-ignore
import '@notopia-uit/lib/yjs';
import 'reflect-metadata';
import { ConfigService } from '@nestjs/config';
import { NestFactory } from '@nestjs/core';
import { WsAdapter } from '@nestjs/platform-ws';
import { Logger } from 'nestjs-pino';

import { getKafkaConfig } from '#/config/kafka.config';

import { AppModule } from './app.module';
import { GlobalExceptionFilter } from './common/http-exception.filter';
import { AppConfig } from './config/config';
import { APP_CONFIG } from './config/config.factory';

declare const module: any;

async function bootstrap() {
  const app = await NestFactory.create(AppModule, { bufferLogs: true });
  app.useWebSocketAdapter(new WsAdapter(app));
  app.connectMicroservice({
    useFactory: getKafkaConfig,
    inject: [ConfigService],
  });
  const logger = app.get(Logger);
  const configService = app.get(ConfigService);
  app.useLogger(logger);
  app.useGlobalFilters(new GlobalExceptionFilter(logger));
  const appConfig = configService.get<AppConfig>(APP_CONFIG);
  if (!appConfig) {
    throw new Error('APP_CONFIG not found');
  }
  const port = appConfig.port;
  await app.startAllMicroservices();
  await app.listen(port);

  if (module.hot) {
    module.hot.accept();
    module.hot.dispose(() => app.close());
  }
}

bootstrap().catch((err) => {
  console.error('Failed to start application', err);
  process.exit(1);
});
