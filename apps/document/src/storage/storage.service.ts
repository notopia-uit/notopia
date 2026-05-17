import { PutObjectCommand, S3Client } from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';
import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Traceable } from 'nestjs-otel';

import { S3Config, S3_CONFIG } from '#/config';

@Injectable()
@Traceable()
export class StorageService {
  private readonly logger = new Logger(StorageService.name);
  private readonly bucketName: string;
  private readonly s3Endpoint: string;
  private static readonly s3UrlExpirationSeconds = 3600;

  constructor(
    configService: ConfigService,
    private readonly s3Client: S3Client
  ) {
    const s3Config = configService.get<S3Config>(S3_CONFIG);
    if (!s3Config) {
      throw new Error('S3_CONFIG not found');
    }
    this.bucketName = s3Config.bucketName;
    this.s3Endpoint = s3Config.endpoint;
  }

  async generateAttachmentPresignedUploadUrl(key: string) {
    this.logger.debug({ key }, 'generateAttachmentPresignedUploadUrl');
    const command = new PutObjectCommand({
      Bucket: this.bucketName,
      Key: key,
    });

    const uploadUrl = await getSignedUrl(this.s3Client, command, {
      expiresIn: StorageService.s3UrlExpirationSeconds,
    });
    const publicUrl = `${this.s3Endpoint}/${this.bucketName}/${key}`;
    this.logger.debug({ key }, 'generateAttachmentPresignedUploadUrl: done');
    return { uploadUrl, publicUrl };
  }
}
