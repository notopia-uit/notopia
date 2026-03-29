import { ServicesConfig } from '../config/config';
import { SERVICE_CONFIG } from '../config/config.factory';
import { NoteModule } from '../note/note.module';
import { NoteService } from '../note/note.service';
import {
  AuthorizationClient,
  AuthorizationService,
} from './authorization.service';
import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-node';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { AuthorizationService as AuthorizationServiceDefinition } from '@notopia-uit/pb/authorization';

const AUTHORIZATION_CLIENT = Symbol('AUTHORIZATION_CLIENT');

@Module({
  imports: [NoteModule],
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
      useFactory: (client: AuthorizationClient, noteService: NoteService) =>
        new AuthorizationService(client, noteService),
      inject: [AUTHORIZATION_CLIENT, NoteService],
    },
  ],
  exports: [AuthorizationService],
})
export class AuthorizationModule {}
