import { S3Config } from '../config/config';
import { S3_CONFIG } from '../config/config.factory';
import { StorageService } from './storage.service';
import { S3Client } from '@aws-sdk/client-s3';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';

@Module({
  providers: [
    {
      provide: S3Client,
      useFactory: (configService: ConfigService) => {
        const s3Config = configService.get<S3Config>(S3_CONFIG);
        if (!s3Config) {
          throw new Error('S3_CONFIG not found');
        }
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
    StorageService,
  ],
  exports: [StorageService],
})
export class StorageModule {}
