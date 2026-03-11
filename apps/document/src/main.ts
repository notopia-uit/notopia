import { NestFactory } from '@nestjs/core';
import { Logger } from 'nestjs-pino';

import { AppModule } from './app/app.module';
import { otelSdk } from './otel/tracing';

otelSdk.start();

async function bootstrap() {
  const app = await NestFactory.create(AppModule, { bufferLogs: true });
  app.useLogger(app.get(Logger));
  const port = process.env.PORT ?? 3000;
  await app.listen(port);
}

bootstrap();
