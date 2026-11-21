import { registerOTel } from '@vercel/otel';

function registerTelemetry(options: Parameters<typeof registerOTel>[0]) {
  const disabled = process.env.OTEL_SDK_DISABLED;
  if (disabled?.toLowerCase() === 'false') {
    delete process.env.OTEL_SDK_DISABLED;
  }

  try {
    registerOTel(options);
  } finally {
    if (disabled !== undefined) {
      process.env.OTEL_SDK_DISABLED = disabled;
    }
  }
}

export async function register() {
  if (process.env.OTEL_SDK_DISABLED?.toLowerCase() === 'true') {
    return;
  }

  if (process.env.NEXT_RUNTIME === 'nodejs') {
    const logsProtocol =
      process.env.OTEL_EXPORTER_OTLP_LOGS_PROTOCOL ?? process.env.OTEL_EXPORTER_OTLP_PROTOCOL;
    const tracesProtocol =
      process.env.OTEL_EXPORTER_OTLP_TRACES_PROTOCOL ?? process.env.OTEL_EXPORTER_OTLP_PROTOCOL;
    if (logsProtocol && logsProtocol !== 'http/protobuf') {
      throw new Error(`Unsupported OTLP logs protocol for web: ${logsProtocol}`);
    }
    if (tracesProtocol && !['http/protobuf', 'http/json'].includes(tracesProtocol)) {
      throw new Error(`Unsupported OTLP traces protocol for web: ${tracesProtocol}`);
    }

    const { OTLPLogExporter } = await import('@opentelemetry/exporter-logs-otlp-proto');
    const { BatchLogRecordProcessor } = await import('@opentelemetry/sdk-logs');
    registerTelemetry({
      serviceName: 'web',
      logRecordProcessors: [new BatchLogRecordProcessor({ exporter: new OTLPLogExporter() })],
    });
    return;
  }

  registerTelemetry({ serviceName: 'web' });
}
