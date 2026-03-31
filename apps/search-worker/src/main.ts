import { AppModule } from './app.module';
import { AppConfig } from './config';
import { APP_CONFIG } from './config.factory';
import { getKafkaConfig } from './kafka.config';
import { otelSdk } from './otel';
import { ConfigService } from '@nestjs/config';
import { NestFactory } from '@nestjs/core';
import { Logger } from 'nestjs-pino';

otelSdk.start();

async function bootstrap() {
  const app = await NestFactory.create(AppModule);
  app.connectMicroservice({
    useFactory: getKafkaConfig,
    inject: [ConfigService],
  });
  const configService = app.get(ConfigService);
  const port = configService.get<AppConfig>(APP_CONFIG)!.port;
  const logger = app.get(Logger);
  app.useLogger(logger);
  await app.startAllMicroservices();
  await app.listen(port);
}

bootstrap();
