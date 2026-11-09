import './otel';
import { trace } from '@opentelemetry/api';
import pino from 'pino';

import { otelSdk } from './otel';
import { ConfigMeilisearch } from './seed';

const logger = pino();

async function main() {
  try {
    await trace.getTracer('search-worker').startActiveSpan('cli.seed', async (span) => {
      try {
        await ConfigMeilisearch();
      } finally {
        span.end();
      }
    });
  } catch {
    logger.error('Failed to configure Meilisearch');
    process.exitCode = 1;
  } finally {
    try {
      await otelSdk.shutdown();
    } catch {
      logger.error('Failed to shut down telemetry');
      process.exitCode = 1;
    }
  }
}

void main();
