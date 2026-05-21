'use client';

import { S3Client, PutObjectCommand, GetObjectCommand } from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';
import { createContext, useContext, useMemo, ReactNode } from 'react';

interface S3StorageContextType {
  uploadFile: (file: File) => Promise<string>;
  resolveFileUrl: (url: string) => Promise<string>;
}

const S3StorageContext = createContext<S3StorageContextType | null>(null);

interface S3StorageProviderProps {
  children: ReactNode;
  endpoint: string;
  bucket: string;
  region: string;
  accessKeyId: string;
  secretAccessKey: string;
}

export function S3StorageProvider({
  children,
  endpoint,
  bucket,
  region,
  accessKeyId,
  secretAccessKey,
}: S3StorageProviderProps) {
  const s3Client = useMemo(() => {
    if (!endpoint || !bucket || !region || !accessKeyId || !secretAccessKey) {
      console.warn('S3Storage provider: missing required configuration');
      return null;
    }

    return new S3Client({
      region,
      endpoint,
      credentials: {
        accessKeyId,
        secretAccessKey,
      },
      forcePathStyle: true,
    });
  }, [endpoint, bucket, region, accessKeyId, secretAccessKey]);

  const uploadFile = useMemo(
    () => async (file: File): Promise<string> => {
      if (!s3Client) {
        throw new Error('S3 client not initialized');
      }

      const key = `${Date.now()}-${file.name}`;
      const command = new PutObjectCommand({
        Bucket: bucket,
        Key: key,
        ContentType: file.type || 'application/octet-stream',
      });

      const signedUrl = await getSignedUrl(s3Client, command, { expiresIn: 3600 });

      const response = await fetch(signedUrl, {
        method: 'PUT',
        body: file,
        headers: {
          'Content-Type': file.type || 'application/octet-stream',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to upload file: ${response.statusText}`);
      }

      return `s3://${bucket}/${key}`;
    },
    [s3Client, bucket]
  );

  const resolveFileUrl = useMemo(
    () => async (url: string): Promise<string> => {
      if (!s3Client || !url.startsWith('s3://')) {
        return url;
      }

      const [, , bucketName, ...keyParts] = url.split('/');
      const key = keyParts.join('/');

      if (bucketName !== bucket) {
        console.warn(`Bucket mismatch: ${bucketName} !== ${bucket}`);
        return url;
      }

      const command = new GetObjectCommand({
        Bucket: bucket,
        Key: key,
      });

      const signedUrl = await getSignedUrl(s3Client, command, { expiresIn: 3600 });
      return signedUrl;
    },
    [s3Client, bucket]
  );

  const value: S3StorageContextType = {
    uploadFile,
    resolveFileUrl,
  };

  return (
    <S3StorageContext.Provider value={value}>
      {children}
    </S3StorageContext.Provider>
  );
}

export function useS3Storage() {
  const context = useContext(S3StorageContext);
  if (!context) {
    throw new Error('useS3Storage must be used within S3StorageProvider');
  }
  return context;
}
