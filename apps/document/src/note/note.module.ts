import { readFileSync } from 'fs';
import { join } from 'path';

import { loadFileDescriptorSetFromBuffer } from '@grpc/proto-loader';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { NOTE_PACKAGE_NAME } from '@notopia-uit/pb/note';

import { ServicesConfig, SERVICE_CONFIG } from '#/config';

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
              packageDefinition: loadFileDescriptorSetFromBuffer(
                readFileSync(join(__dirname, 'proto/build.bin'))
              ),
              url: servicesCfg.noteUrl,
              gracefulShutdown: true,
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
