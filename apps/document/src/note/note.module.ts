import { ServicesConfig } from '../config/config';
import { SERVICE_CONFIG } from '../config/config.factory';
import { NoteService } from './note.service';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { NOTE_PACKAGE_NAME } from '@notopia-uit/pb/note';
import { join } from 'node:path';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: NOTE_PACKAGE_NAME,
        imports: [ConfigModule],
        useFactory: (configService: ConfigService) => {
          const servicesCfg =
            configService.get<ServicesConfig>(SERVICE_CONFIG)!;
          return {
            transport: Transport.GRPC,
            options: {
              package: NOTE_PACKAGE_NAME,
              protoPath: join(__dirname, '../../../../proto/note/note.proto'),
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
