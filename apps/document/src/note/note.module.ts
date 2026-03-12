import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-node';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { NoteService } from '@notopia-uit/pb/note';

import { ServicesConfig } from '../config/config';

export const NOTE_SERVICE = Symbol('NOTE_SERVICE');

@Module({
  providers: [
    {
      provide: NOTE_SERVICE,
      useFactory: (configService: ConfigService) => {
        const servicesCfg = configService.get<ServicesConfig>('services')!;
        const transport = createConnectTransport({
          baseUrl: servicesCfg.noteUrl,
          httpVersion: '1.1',
        });
        return createClient(NoteService, transport);
      },
      inject: [ConfigService],
    },
  ],
  exports: [NOTE_SERVICE],
})
export class NoteModule {}
