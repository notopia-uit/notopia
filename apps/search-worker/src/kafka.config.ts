import { ConfigService } from '@nestjs/config';
import { KafkaOptions, Transport } from '@nestjs/microservices';

import { KafkaConfig } from './config';
import { KAFKA_CONFIG } from './config.factory';

export const getKafkaConfig = (configService: ConfigService): KafkaOptions => {
  const config = configService.get<KafkaConfig>(KAFKA_CONFIG)!;
  return {
    transport: Transport.KAFKA,
    options: {
      client: {
        clientId: config.clientId,
        brokers: config.brokers,
      },
      consumer: {
        groupId: config.groupId,
      },
    },
  };
};
