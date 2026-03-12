import { Client, createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-node';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { AuthorizationService } from '@notopia-uit/pb/authorization';

import { ServicesConfig } from '../config/config';

export const AUTHORIZATION_SERVICE = Symbol('AUTHORIZATION_SERVICE');

@Module({
  providers: [
    {
      provide: AUTHORIZATION_SERVICE,
      useFactory: (
        configService: ConfigService
      ): Client<typeof AuthorizationService> => {
        const servicesCfg = configService.get<ServicesConfig>('services')!;
        const transport = createConnectTransport({
          baseUrl: servicesCfg.authorizationUrl,
          httpVersion: '1.1',
        });
        return createClient(AuthorizationService, transport);
      },
      inject: [ConfigService],
    },
  ],
  exports: [AUTHORIZATION_SERVICE],
})
export class AuthorizationModule {}
