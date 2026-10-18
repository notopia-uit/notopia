import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import {
  BILLING_PACKAGE_NAME,
  BILLING_SERVICE_NAME,
  BillingServiceService,
} from '@notopia-uit/pb/billing';

import { SERVICE_CONFIG, ServicesConfig } from '#/config';

import { BillingService } from './billing.service';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: BILLING_PACKAGE_NAME,
        imports: [ConfigModule],
        inject: [ConfigService],
        useFactory: (configService: ConfigService) => {
          const servicesCfg = configService.get<ServicesConfig>(SERVICE_CONFIG);
          if (!servicesCfg) {
            throw new Error('SERVICE_CONFIG not found');
          }
          return {
            transport: Transport.GRPC,
            options: {
              package: BILLING_PACKAGE_NAME,
              packageDefinition: {
                [`${BILLING_PACKAGE_NAME}.${BILLING_SERVICE_NAME}`]: BillingServiceService,
              },
              url: servicesCfg.billingUrl,
              gracefulShutdown: true,
            },
          };
        },
      },
    ]),
  ],
  providers: [BillingService],
  exports: [BillingService],
})
export class BillingModule {}
