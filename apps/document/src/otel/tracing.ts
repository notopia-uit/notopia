import { credentials } from '@grpc/grpc-js';
import { OTLPMetricExporter } from '@opentelemetry/exporter-metrics-otlp-grpc';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-grpc';
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http';
import { PinoInstrumentation } from '@opentelemetry/instrumentation-pino';
import { Resource } from '@opentelemetry/resources';
import { PeriodicExportingMetricReader } from '@opentelemetry/sdk-metrics';
import { NodeSDK } from '@opentelemetry/sdk-node';
import {
  ParentBasedSampler,
  TraceIdRatioBased,
} from '@opentelemetry/sdk-trace-node';
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions';

function buildSdk(): NodeSDK {
  const otelEnabled = process.env.OTEL_ENABLED === 'true';

  const resource = new Resource({
    [ATTR_SERVICE_NAME]: process.env.OTEL_SERVICE_NAME ?? 'document',
  });

  const traceEnabled = otelEnabled && process.env.OTEL_TRACE_ENABLED === 'true';
  const traceEndpoint = process.env.OTEL_TRACE_GRPC_ENDPOINT ?? '';
  const traceInsecure = process.env.OTEL_TRACE_GRPC_INSECURE === 'true';
  const sampleRate = parseFloat(process.env.OTEL_TRACE_SAMPLE_RATE ?? '1.0');

  const meterEnabled = otelEnabled && process.env.OTEL_METER_ENABLED === 'true';
  const meterEndpoint = process.env.OTEL_METER_GRPC_ENDPOINT ?? '';
  const meterInsecure = process.env.OTEL_METER_GRPC_INSECURE === 'true';
  const exportInterval = parseInt(
    process.env.OTEL_METER_EXPORT_INTERVAL ?? '60000',
    10
  );

  const sdkOptions: ConstructorParameters<typeof NodeSDK>[0] = {
    resource,
    instrumentations: [new HttpInstrumentation(), new PinoInstrumentation()],
  };

  if (traceEnabled && traceEndpoint) {
    sdkOptions.traceExporter = new OTLPTraceExporter({
      url: traceEndpoint,
      credentials: traceInsecure
        ? credentials.createInsecure()
        : credentials.createSsl(),
    });
    sdkOptions.sampler = new ParentBasedSampler({
      root: new TraceIdRatioBased(sampleRate),
    });
  }

  if (meterEnabled && meterEndpoint) {
    sdkOptions.metricReader = new PeriodicExportingMetricReader({
      exporter: new OTLPMetricExporter({
        url: meterEndpoint,
        credentials: meterInsecure
          ? credentials.createInsecure()
          : credentials.createSsl(),
      }),
      exportIntervalMillis: exportInterval,
    });
  }

  return new NodeSDK(sdkOptions);
}

export const otelSdk = buildSdk();
