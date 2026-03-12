import { Client, createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-node';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { createStrictClient } from '@notopia-uit/lib/connectrpc';
import { NoteService } from '@notopia-uit/pb/note';

import { ServicesConfig } from '../config/config';

export const NOTE_SERVICE = Symbol('NOTE_SERVICE');

@Module({
  providers: [
    {
      provide: NOTE_SERVICE,
      useFactory: (
        configService: ConfigService
      ): Client<typeof NoteService> => {
        const servicesCfg = configService.get<ServicesConfig>('services')!;
        const transport = createConnectTransport({
          baseUrl: servicesCfg.noteUrl,
          httpVersion: '1.1',
        });
        const client = createStrictClient(NoteService, transport);
        client.checkNoteExistence;
      },
      inject: [ConfigService],
    },
  ],
  exports: [NOTE_SERVICE],
})
export class NoteModule {}
