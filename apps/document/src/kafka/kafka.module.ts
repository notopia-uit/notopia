import { Global, Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { ClientProxyFactory } from '@nestjs/microservices';

import { KAFKA_CLIENT } from '#/common/token';
import { getKafkaConfig } from '#/config/kafka.config';

@Global()
@Module({
  providers: [
    {
      provide: KAFKA_CLIENT,
      useFactory: (configService: ConfigService) =>
        ClientProxyFactory.create(getKafkaConfig(configService)),
      inject: [ConfigService],
    },
  ],
  exports: [KAFKA_CLIENT],
})
export class KafkaModule {}