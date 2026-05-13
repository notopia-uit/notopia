import { join } from 'path';

import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { AUTHORIZATION_PACKAGE_NAME } from '@notopia-uit/pb/authorization';

import { ServicesConfig } from '#/config/config';
import { SERVICE_CONFIG } from '#/config/config.factory';
import { NoteModule } from '#/note/note.module';

import { AuthorizationService } from './authorization.service';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: AUTHORIZATION_PACKAGE_NAME,
        imports: [ConfigModule],
        inject: [ConfigService],
        useFactory: (configService: ConfigService) => {
          const servicesCfg = configService.get<ServicesConfig>(SERVICE_CONFIG);
          if (!servicesCfg) {
            throw new Error('SERVICE_CONFIG not found');
          }
          const protoPath = join(process.cwd(), 'proto/authorization/authorization.proto');
          const includeDirs = [join(process.cwd(), 'proto')];
          return {
            transport: Transport.GRPC,
            options: {
              package: AUTHORIZATION_PACKAGE_NAME,
              protoPath,
              loader: {
                includeDirs,
              },
              url: servicesCfg.authorizationUrl,
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
