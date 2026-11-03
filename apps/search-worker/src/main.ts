// oxfmt-ignore
import './otel';

import { ConfigService } from '@nestjs/config';
import { NestFactory } from '@nestjs/core';
import { Logger } from 'nestjs-pino';

import { AppModule } from './app.module';
import { AppConfig } from './config';
import { APP_CONFIG } from './config.factory';
import { KafkaConsumerState } from './health';
import { getKafkaConfig } from './kafka.config';
import { TrackedKafkaServer } from './kafka.server';

async function bootstrap() {
  const app = await NestFactory.create(AppModule, { bufferLogs: true });
  const configService = app.get(ConfigService);
  const state = app.get(KafkaConsumerState);
  app.connectMicroservice({
    strategy: new TrackedKafkaServer(getKafkaConfig(configService).options!, state),
  });
  const appConfig = configService.get<AppConfig>(APP_CONFIG);
  if (!appConfig) {
    throw new Error('APP_CONFIG not found');
  }
  const port = appConfig.port;
  const logger = app.get(Logger);
  app.useLogger(logger);
  app.enableShutdownHooks();
  await app.listen(port);
  void app.startAllMicroservices().catch((error: unknown) => {
    logger.error(error, 'Failed to start Kafka microservice');
  });

  if (module.hot) {
    module.hot.accept();
    module.hot.dispose(() => app.close());
  }
}

bootstrap().catch((err) => {
  console.error('Failed to start application', err);
  process.exit(1);
});
