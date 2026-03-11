import { registerAs } from '@nestjs/config';

import { AppConfig } from './config';

export const appConfig = registerAs(
  'app',
  (): AppConfig => ({
    env: process.env.NODE_ENV ?? 'production',
    database: {
      host: process.env.DB_HOST ?? 'localhost',
      port: parseInt(process.env.DB_PORT ?? '5432', 10),
      username: process.env.DB_USER ?? 'postgres',
      password: process.env.DB_PASSWORD ?? '',
      database: process.env.DB_NAME ?? 'document',
    },
    otel: {
      enabled: process.env.OTEL_ENABLED === 'true',
      stdout: process.env.OTEL_STDOUT === 'true',
      trace: {
        enabled: process.env.OTEL_TRACE_ENABLED === 'true',
        sampleRate: parseFloat(process.env.OTEL_TRACE_SAMPLE_RATE ?? '1.0'),
        grpc: {
          endpoint: process.env.OTEL_TRACE_GRPC_ENDPOINT ?? '',
          insecure: process.env.OTEL_TRACE_GRPC_INSECURE === 'true',
        },
        stdout: process.env.OTEL_TRACE_STDOUT === 'true',
      },
      log: {
        enabled: process.env.OTEL_LOG_ENABLED === 'true',
        level: process.env.OTEL_LOG_LEVEL ?? 'info',
        grpc: {
          endpoint: process.env.OTEL_LOG_GRPC_ENDPOINT ?? '',
          insecure: process.env.OTEL_LOG_GRPC_INSECURE === 'true',
        },
        stdout: process.env.OTEL_LOG_STDOUT === 'true',
      },
      meter: {
        enabled: process.env.OTEL_METER_ENABLED === 'true',
        grpc: {
          endpoint: process.env.OTEL_METER_GRPC_ENDPOINT ?? '',
          insecure: process.env.OTEL_METER_GRPC_INSECURE === 'true',
        },
        stdout: process.env.OTEL_METER_STDOUT === 'true',
        exportInterval: parseInt(
          process.env.OTEL_METER_EXPORT_INTERVAL ?? '60000',
          10
        ),
      },
    },
  })
);
