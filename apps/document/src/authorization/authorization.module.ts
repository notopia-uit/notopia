import { readFile } from 'fs/promises';

import { loadFileDescriptorSetFromBuffer } from '@grpc/proto-loader';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { AUTHORIZATION_PACKAGE_NAME } from '@notopia-uit/pb/authorization';

import { ServicesConfig } from '#/config';
import { SERVICE_CONFIG } from '#/config';

import { NoteModule } from '../note/note.module';
import { AuthorizationService } from './authorization.service';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: AUTHORIZATION_PACKAGE_NAME,
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
              package: AUTHORIZATION_PACKAGE_NAME,
              packageDefinition: packageDefinition,
              url: servicesCfg.authorizationUrl,
              gracefulShutdown: true,
            },
          };
        },
      },
    ]),
    NoteModule,
  ],
  providers: [AuthorizationService],
  exports: [AuthorizationService],
})
export class AuthorizationModule {}
