import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import {
  BILLING_PACKAGE_NAME,
  BILLING_SERVICE_NAME,
  BillingServiceService,
} from '@notopia-uit/pb/billing';

import type { ServicesConfig } from '#/config';
import { SERVICES_CONFIG } from '#/config.factory';

import { BillingService } from './billing.service';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: BILLING_PACKAGE_NAME,
        imports: [ConfigModule],
        inject: [ConfigService],
        useFactory: (configService: ConfigService) => {
          const services = configService.get<ServicesConfig>(SERVICES_CONFIG);
          if (!services) {
            throw new Error('SERVICES_CONFIG not found');
          }
          return {
            transport: Transport.GRPC,
            options: {
              package: BILLING_PACKAGE_NAME,
              packageDefinition: {
                [`${BILLING_PACKAGE_NAME}.${BILLING_SERVICE_NAME}`]: BillingServiceService,
              },
              url: services.billingGrpcUrl,
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
