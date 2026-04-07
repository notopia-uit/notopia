import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { NodeSDK } from '@opentelemetry/sdk-node';

if (process.env.OTEL_SDK_DISABLED === undefined) {
  process.env.OTEL_SDK_DISABLED = 'true';
  process.env.OTEL_NODE_DISABLED_INSTRUMENTATIONS = 'all';
} else {
  process.env.OTEL_NODE_ENABLED_INSTRUMENTATIONS = 'kafkajs,pino,runtime-node';
}

export const otelSdk = new NodeSDK({
  instrumentations: [getNodeAutoInstrumentations()],
});

process.on('SIGTERM', () => {
  otelSdk
    .shutdown()
    .then(() => console.log('Otel SDK shut down successfully'))
    .catch((error) => console.error('Error shutting down Otel SDK', error))
    .finally(() => process.exit(0));
});
