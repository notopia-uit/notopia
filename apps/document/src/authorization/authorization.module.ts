import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-node';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { AuthorizationService as AuthorizationServiceDefinition } from '@notopia-uit/pb/authorization';

import { ServicesConfig } from '../config/config';
import { SERVICE_CONFIG } from '../config/config.factory';
import {
  AuthorizationClient,
  AuthorizationService,
} from './authorization.service';

export const AUTHORIZATION_SERVICE = Symbol('AUTHORIZATION_SERVICE');
const AUTHORIZATION_CLIENT = Symbol('AUTHORIZATION_CLIENT');

@Module({
  providers: [
    {
      provide: AUTHORIZATION_CLIENT,
      useFactory: (configService: ConfigService): AuthorizationClient => {
        const servicesCfg = configService.get<ServicesConfig>(SERVICE_CONFIG)!;
        const transport = createConnectTransport({
          baseUrl: servicesCfg.authorizationUrl,
          httpVersion: '1.1',
        });
        return createClient(AuthorizationServiceDefinition, transport);
      },
      inject: [ConfigService],
    },
    {
      provide: AuthorizationService,
      useFactory: (client: AuthorizationClient) =>
        new AuthorizationService(client),
      inject: [AUTHORIZATION_CLIENT],
    },
  ],
  exports: [AuthorizationService],
})
export class AuthorizationModule {}
