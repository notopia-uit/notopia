import { readFile } from 'fs/promises';

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
        useFactory: async (configService: ConfigService) => {
          const servicesCfg = configService.get<ServicesConfig>(SERVICE_CONFIG);
          if (!servicesCfg) {
            throw new Error('SERVICE_CONFIG not found');
          }
          const binaryBuffer = await readFile('proto/build.bin');
          const packageDefinition = loadFileDescriptorSetFromBuffer(binaryBuffer, {
            keepCase: false,
          });
          return {
            transport: Transport.GRPC,
            options: {
              package: NOTE_PACKAGE_NAME,
              packageDefinition: packageDefinition,
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
