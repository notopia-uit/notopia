declare namespace NodeJS {
  interface ProcessEnv {
    OTEL_ENABLED?: 'true' | 'false';
    OTEL_STDOUT?: 'true' | 'false';

    OTEL_TRACE_ENABLED?: 'true' | 'false';
    OTEL_TRACE_SAMPLE_RATE?: string; // float, e.g., '1.0'
    OTEL_TRACE_GRPC_ENDPOINT?: string;
    OTEL_TRACE_GRPC_INSECURE?: 'true' | 'false';
    OTEL_TRACE_STDOUT?: 'true' | 'false';

    OTEL_LOG_ENABLED?: 'true' | 'false';
    OTEL_LOG_LEVEL?: 'debug' | 'info' | 'warn' | 'error';
    OTEL_LOG_GRPC_ENDPOINT?: string;
    OTEL_LOG_GRPC_INSECURE?: 'true' | 'false';
    OTEL_LOG_STDOUT?: 'true' | 'false';

    OTEL_METER_ENABLED?: 'true' | 'false';
    OTEL_METER_GRPC_ENDPOINT?: string;
    OTEL_METER_GRPC_INSECURE?: 'true' | 'false';
    OTEL_METER_STDOUT?: 'true' | 'false';
    OTEL_METER_EXPORT_INTERVAL?: string; // integer, e.g., '60000'

    DB_HOST?: string;
    DB_PORT?: string;
    DB_USER?: string;
    DB_PASSWORD?: string;
    DB_NAME?: string;

    NODE_ENV?: 'development' | 'production' | 'test';
    PORT?: string;
  }
}
