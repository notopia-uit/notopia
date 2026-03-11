import { credentials } from '@grpc/grpc-js';
import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { OTLPMetricExporter } from '@opentelemetry/exporter-metrics-otlp-grpc';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-grpc';
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http';
import { PinoInstrumentation } from '@opentelemetry/instrumentation-pino';
import { Resource } from '@opentelemetry/resources';
import { PeriodicExportingMetricReader } from '@opentelemetry/sdk-metrics';
import { NodeSDK } from '@opentelemetry/sdk-node';
import {
  ParentBasedSampler,
  TraceIdRatioBasedSampler,
} from '@opentelemetry/sdk-trace-node';
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions';

import type { AppConfig } from '../config/config';

@Injectable()
export class OtelService {
  public readonly sdk: NodeSDK;

  constructor(private configService: ConfigService) {
    this.sdk = this.buildSdk();
  }

  private buildSdk(): NodeSDK {
    const appConfig = this.configService.get<AppConfig>('app');
    const otelConfig = appConfig?.otel;

    const resource = new Resource({
      [ATTR_SERVICE_NAME]: 'document',
    });

    const sdkOptions: ConstructorParameters<typeof NodeSDK>[0] = {
      resource,
      instrumentations: [new HttpInstrumentation(), new PinoInstrumentation()],
    };

    if (otelConfig?.trace.enabled && otelConfig.trace.grpc.endpoint) {
      sdkOptions.traceExporter = new OTLPTraceExporter({
        url: otelConfig.trace.grpc.endpoint,
        credentials: otelConfig.trace.grpc.insecure
          ? credentials.createInsecure()
          : credentials.createSsl(),
      });
      sdkOptions.sampler = new ParentBasedSampler({
        root: new TraceIdRatioBasedSampler(otelConfig.trace.sampleRate),
      });
    }

    if (otelConfig?.meter.enabled && otelConfig.meter.grpc.endpoint) {
      sdkOptions.metricReader = new PeriodicExportingMetricReader({
        exporter: new OTLPMetricExporter({
          url: otelConfig.meter.grpc.endpoint,
          credentials: otelConfig.meter.grpc.insecure
            ? credentials.createInsecure()
            : credentials.createSsl(),
        }),
        exportIntervalMillis: otelConfig.meter.exportInterval,
      });
    }

    return new NodeSDK(sdkOptions);
  }
}

export const otelSdk = new NodeSDK({
  resource: new Resource({
    [ATTR_SERVICE_NAME]: 'document',
  }),
  instrumentations: [
    getNodeAutoInstrumentations(),
    new HttpInstrumentation(),
    new PinoInstrumentation(),
  ],
});
