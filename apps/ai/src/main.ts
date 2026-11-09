// oxfmt-ignore
import './otel';
// oxfmt-ignore
import 'reflect-metadata';

import { ConfigService } from '@nestjs/config';
import { NestFactory } from '@nestjs/core';
import type { NestExpressApplication } from '@nestjs/platform-express';
import { Logger } from 'nestjs-pino';

import { AppModule } from './app.module';
import { AiExceptionFilter } from './common/ai-error';
import { AppConfig } from './config';
import { APP_CONFIG } from './config.factory';

async function bootstrap() {
  const app = await NestFactory.create<NestExpressApplication>(AppModule, {
    bufferLogs: true,
    bodyParser: false,
  });
  const logger = app.get(Logger);
  const configService = app.get(ConfigService);
  app.useLogger(logger);
  app.useBodyParser('json', { limit: 512000 });
  app.useGlobalFilters(new AiExceptionFilter());
  const appConfig = configService.get<AppConfig>(APP_CONFIG);
  if (!appConfig) {
    throw new Error('APP_CONFIG not found');
  }
  const port = appConfig.port;
  app.enableShutdownHooks();
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
