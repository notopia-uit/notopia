import { S3Client } from '@aws-sdk/client-s3';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';

import { S3Config } from '../config/config';

@Module({
  providers: [
    {
      provide: S3Client,
      useFactory: (configService: ConfigService) => {
        const s3Config = configService.get<S3Config>('s3')!;
        return new S3Client({
          region: s3Config.region,
          forcePathStyle: true,
          endpoint: s3Config.endpoint,
          credentials: {
            accessKeyId: s3Config.accessKeyId,
            secretAccessKey: s3Config.secretAccessKey,
          },
        });
      },
      inject: [ConfigService],
    },
  ],
  exports: [S3Client],
})
export class S3Module {}
