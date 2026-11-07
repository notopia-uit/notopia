import { NodeSDK } from '@opentelemetry/sdk-node';

export const otelSdk = new NodeSDK();
otelSdk.start();

process.once('SIGTERM', () => {
  void otelSdk
    .shutdown()
    .catch(() => {
      process.stderr.write('Failed to shut down telemetry\n');
      process.exitCode = 1;
    })
    .finally(() => process.exit(process.exitCode || 0));
});
