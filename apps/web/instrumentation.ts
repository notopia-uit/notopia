import { registerOTel } from '@vercel/otel';

export async function register() {
  if (process.env.NEXT_RUNTIME === 'nodejs') {
    const { OTLPLogExporter } = await import('@opentelemetry/exporter-logs-otlp-proto');
    const { BatchLogRecordProcessor } = await import('@opentelemetry/sdk-logs');
    registerOTel({
      serviceName: 'web',
      logRecordProcessors: [new BatchLogRecordProcessor(new OTLPLogExporter())],
    });
    return;
  }

  registerOTel({ serviceName: 'web' });
}
