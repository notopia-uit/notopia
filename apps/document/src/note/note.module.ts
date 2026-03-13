import { Client, createClient } from '@connectrpc/connect';
import { createGrpcTransport } from '@connectrpc/connect-node';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
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
        const transport = createGrpcTransport({
          baseUrl: servicesCfg.noteUrl,
        });
        return createClient(NoteService, transport);
      },
      inject: [ConfigService],
    },
  ],
  exports: [NOTE_SERVICE],
})
export class NoteModule {}
