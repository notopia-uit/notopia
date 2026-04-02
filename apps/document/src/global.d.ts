import { User } from './common/user';

declare namespace NodeJS {
  interface ProcessEnv {
    OTEL_SDK_DISABLED?: 'true' | 'false'; // Standard: Master switch
    OTEL_SERVICE_NAME?: string;
    OTEL_RESOURCE_ATTRIBUTES?: string; // 'env=prod,version=1.0.0'
    OTEL_TRACES_EXPORTER?: 'otlp' | 'console' | 'none';
    OTEL_TRACES_SAMPLER?:
      | 'always_on'
      | 'always_off'
      | 'traceidratio'
      | 'parentbased_traceidratio';
    OTEL_TRACES_SAMPLER_ARG?: string; // (0.0 - 1.0)
    OTEL_EXPORTER_OTLP_TRACES_ENDPOINT?: string; // Standard
    OTEL_EXPORTER_OTLP_TRACES_PROTOCOL?: 'grpc' | 'http/protobuf' | 'http/json';
    OTEL_METRICS_EXPORTER?: 'otlp' | 'console' | 'none';
    OTEL_METRIC_EXPORT_INTERVAL?: string; // (ms)
    OTEL_EXPORTER_OTLP_METRICS_ENDPOINT?: string; // Standard
    OTEL_EXPORTER_OTLP_METRICS_PROTOCOL?: 'grpc' | 'http/protobuf';
    OTEL_LOGS_EXPORTER?: 'otlp' | 'console' | 'none'; // Standard
    OTEL_LOG_LEVEL?: 'debug' | 'info' | 'warn' | 'error';
    OTEL_EXPORTER_OTLP_LOGS_ENDPOINT?: string;
    OTEL_EXPORTER_OTLP_ENDPOINT?: string; // 'http://localhost:4317'
    OTEL_EXPORTER_OTLP_PROTOCOL?: 'grpc' | 'http/protobuf';
    OTEL_EXPORTER_OTLP_HEADERS?: string; // e.g., 'api-key=123,auth=xyz'
    OTEL_EXPORTER_OTLP_TIMEOUT?: string; // (ms)

    NODE_ENV?: 'development' | 'production' | 'test';

    NOTOPIA_DOCUMENT_PORT?: string;
    NOTOPIA_DOCUMENT_LOG_LEVEL?:
      | 'trace'
      | 'debug'
      | 'info'
      | 'warn'
      | 'error'
      | 'fatal';
    NOTOPIA_DOCUMENT_API_URL?: string; // public serve, for it own, and other services reference

    NOTOPIA_DOCUMENT_DB_HOST?: string;
    NOTOPIA_DOCUMENT_DB_PORT?: string;
    NOTOPIA_DOCUMENT_DB_USER?: string;
    NOTOPIA_DOCUMENT_DB_PASSWORD?: string;
    NOTOPIA_DOCUMENT_DB_NAME?: string;

    NOTOPIA_DOCUMENT_SERVICES_NOTE_GPRC_URL?: string;
    NOTOPIA_DOCUMENT_SERVICES_AUTHORIZATION_GRPC_URL?: string;
  }
}

declare global {
  namespace Express {
    interface Request {
      user?: User;
    }
  }
}
