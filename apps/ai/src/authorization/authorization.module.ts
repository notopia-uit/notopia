import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import {
  AUTHORIZATION_PACKAGE_NAME,
  AUTHORIZATION_SERVICE_NAME,
  AuthorizationServiceService,
} from '@notopia-uit/pb/authorization';

import type { ServicesConfig } from '#/config';
import { SERVICES_CONFIG } from '#/config.factory';

import { AuthorizationService } from './authorization.service';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: AUTHORIZATION_PACKAGE_NAME,
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
              package: AUTHORIZATION_PACKAGE_NAME,
              packageDefinition: {
                [`${AUTHORIZATION_PACKAGE_NAME}.${AUTHORIZATION_SERVICE_NAME}`]:
                  AuthorizationServiceService,
              },
              url: services.authorizationGrpcUrl,
              gracefulShutdown: true,
            },
          };
        },
      },
    ]),
  ],
  providers: [AuthorizationService],
  exports: [AuthorizationService],
})
export class AuthorizationModule {}
