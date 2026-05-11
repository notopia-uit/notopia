import { Global, Module } from '@nestjs/common';
import { ClientsModule } from '@nestjs/microservices';

import { KAFKA_CLIENT } from '#/common/token';
import { getKafkaConfig } from '#/config/kafka.config';

@Global()
@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: KAFKA_CLIENT,
        useFactory: getKafkaConfig,
      },
    ]),
  ],
  exports: [KAFKA_CLIENT],
})
export class KafkaModule {}