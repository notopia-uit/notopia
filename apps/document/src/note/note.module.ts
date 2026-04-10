import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { NOTE_PACKAGE_NAME } from '@notopia-uit/pb/note';
import { join } from 'path';

import { ServicesConfig } from '#/config/config';
import { SERVICE_CONFIG } from '#/config/config.factory';

import { NoteService } from './note.service';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: NOTE_PACKAGE_NAME,
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
              package: NOTE_PACKAGE_NAME,
              protoPath: join(__dirname, '../../../proto/note/note.proto'),
              loader: {
                includeDirs: [join(__dirname, '../../../proto')],
              },
              url: servicesCfg.noteUrl,
            },
          };
        },
      },
    ]),
  ],
  providers: [NoteService],
  exports: [NoteService],
})
export class NoteModule {}
