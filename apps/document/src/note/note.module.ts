import { createClient } from '@connectrpc/connect';
import { createGrpcTransport } from '@connectrpc/connect-node';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { NoteService as NoteServiceDefinition } from '@notopia-uit/pb/note';

import { ServicesConfig } from '../config/config';
import { SERVICE_CONFIG } from '../config/config.factory';
import { NoteClient, NoteService } from './note.service';

const NOTE_CLIENT = Symbol('NOTE_SERVICE_CLIENT');

@Module({
  providers: [
    {
      provide: NOTE_CLIENT,
      useFactory: (configService: ConfigService): NoteClient => {
        const servicesCfg = configService.get<ServicesConfig>(SERVICE_CONFIG)!;
        const transport = createGrpcTransport({
          baseUrl: servicesCfg.noteUrl,
        });
        return createClient(NoteServiceDefinition, transport);
      },
      inject: [ConfigService],
    },
    {
      provide: NoteService,
      useFactory: (client: NoteClient) => new NoteService(client),
      inject: [NOTE_CLIENT],
    },
  ],
  exports: [NoteService],
})
export class NoteModule {}
