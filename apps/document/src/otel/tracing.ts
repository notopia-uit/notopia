import { PinoInstrumentation } from '@opentelemetry/instrumentation-pino';
import { NodeSDK } from '@opentelemetry/sdk-node';

function buildSdk(): NodeSDK {
  return new NodeSDK({
    instrumentations: [new PinoInstrumentation()],
  });
}

export const otelSdk = buildSdk();
