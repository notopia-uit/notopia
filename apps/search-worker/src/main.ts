// oxfmt-ignore
import './otel';

import { ConfigService } from '@nestjs/config';
import { NestFactory } from '@nestjs/core';
import { Logger } from 'nestjs-pino';

import { AppModule } from './app.module';
import { AppConfig } from './config';
import { APP_CONFIG } from './config.factory';
import { getKafkaConfig } from './kafka.config';

async function bootstrap() {
  const app = await NestFactory.create(AppModule, { bufferLogs: true });
  app.connectMicroservice({
    useFactory: getKafkaConfig,
    inject: [ConfigService],
  });
  const configService = app.get(ConfigService);
  const appConfig = configService.get<AppConfig>(APP_CONFIG);
  if (!appConfig) {
    throw new Error('APP_CONFIG not found');
  }
  const port = appConfig.port;
  const logger = app.get(Logger);
  app.useLogger(logger);
  app.enableShutdownHooks();
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
