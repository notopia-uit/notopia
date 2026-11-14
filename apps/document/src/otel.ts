import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { NodeSDK } from '@opentelemetry/sdk-node';

const sdkDisabled = process.env.OTEL_SDK_DISABLED === 'true';
if (!sdkDisabled && !process.env.OTEL_NODE_ENABLED_INSTRUMENTATIONS) {
  process.env.OTEL_NODE_ENABLED_INSTRUMENTATIONS = 'grpc,http,kafkajs,pino,pg,runtime-node';
}

export const otelSdk = new NodeSDK({
  instrumentations: sdkDisabled ? [] : [getNodeAutoInstrumentations()],
});

otelSdk.start();

process.on('SIGTERM', () => {
  otelSdk
    .shutdown()
    .then(() => console.log('Otel SDK shut down successfully'))
    .catch((error) => console.error('Error shutting down Otel SDK', error))
    .finally(() => process.exit(0));
});
