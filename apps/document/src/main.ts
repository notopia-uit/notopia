import { NestFactory } from '@nestjs/core';
import { Logger } from 'nestjs-pino';

import { AppModule } from './app.module';
import { GlobalExceptionFilter } from './common/http-exception.filter';
import { otelSdk } from './otel/tracing';

otelSdk.start();

async function bootstrap() {
  const app = await NestFactory.create(AppModule, { bufferLogs: true });
  const logger = app.get(Logger);
  app.useLogger(logger);

  app.useGlobalFilters(new GlobalExceptionFilter(logger));

  const port = process.env.PORT ?? 8082;
  await app.listen(port);
}

bootstrap();
