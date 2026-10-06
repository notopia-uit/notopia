import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { NOTE_PACKAGE_NAME, NOTE_SERVICE_NAME, NoteServiceService } from '@notopia-uit/pb/note';

import type { ServicesConfig } from '#/config';
import { SERVICES_CONFIG } from '#/config.factory';

import { NoteClientService } from './note-client.service';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: NOTE_PACKAGE_NAME,
        imports: [ConfigModule],
        inject: [ConfigService],
        useFactory: (configService: ConfigService) => {
          const services = configService.get<ServicesConfig>(SERVICES_CONFIG);
          if (!services) {
            throw new Error('SERVICES_CONFIG not found');
          }
          return {
            transport: Transport.GRPC,
            options: {
              package: NOTE_PACKAGE_NAME,
              packageDefinition: {
                [`${NOTE_PACKAGE_NAME}.${NOTE_SERVICE_NAME}`]: NoteServiceService,
              },
              url: services.noteGrpcUrl,
              gracefulShutdown: true,
            },
          };
        },
      },
    ]),
  ],
  providers: [NoteClientService],
  exports: [NoteClientService],
})
export class NoteModule {}
