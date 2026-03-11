import { credentials } from '@grpc/grpc-js';
import { OTLPMetricExporter } from '@opentelemetry/exporter-metrics-otlp-grpc';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-grpc';
import { PinoInstrumentation } from '@opentelemetry/instrumentation-pino';
import { Resource } from '@opentelemetry/resources';
import { PeriodicExportingMetricReader } from '@opentelemetry/sdk-metrics';
import { NodeSDK } from '@opentelemetry/sdk-node';
import {
  ParentBasedSampler,
  TraceIdRatioBasedSampler,
} from '@opentelemetry/sdk-trace-node';
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions';

function buildSdk(): NodeSDK {
  const resource = new Resource({
    [ATTR_SERVICE_NAME]: 'document',
  });

  const sdkOptions: ConstructorParameters<typeof NodeSDK>[0] = {
    resource,
    instrumentations: [new PinoInstrumentation()],
  };

  const traceEnabled = process.env['OTEL_TRACE_ENABLED'] === 'true';
  const traceEndpoint = process.env['OTEL_TRACE_GRPC_ENDPOINT'] ?? '';

  if (traceEnabled && traceEndpoint) {
    const insecure = process.env['OTEL_TRACE_GRPC_INSECURE'] === 'true';
    const sampleRate = parseFloat(
      process.env['OTEL_TRACE_SAMPLE_RATE'] ?? '1.0'
    );

    sdkOptions.traceExporter = new OTLPTraceExporter({
      url: traceEndpoint,
      credentials: insecure
        ? credentials.createInsecure()
        : credentials.createSsl(),
    });
    sdkOptions.sampler = new ParentBasedSampler({
      root: new TraceIdRatioBasedSampler(sampleRate),
    });
  }

  const meterEnabled = process.env['OTEL_METER_ENABLED'] === 'true';
  const meterEndpoint = process.env['OTEL_METER_GRPC_ENDPOINT'] ?? '';

  if (meterEnabled && meterEndpoint) {
    const insecure = process.env['OTEL_METER_GRPC_INSECURE'] === 'true';
    const exportInterval = parseInt(
      process.env['OTEL_METER_EXPORT_INTERVAL'] ?? '60000',
      10
    );

    sdkOptions.metricReader = new PeriodicExportingMetricReader({
      exporter: new OTLPMetricExporter({
        url: meterEndpoint,
        credentials: insecure
          ? credentials.createInsecure()
          : credentials.createSsl(),
      }),
      exportIntervalMillis: exportInterval,
    });
  }

  return new NodeSDK(sdkOptions);
}

export const otelSdk = buildSdk();
