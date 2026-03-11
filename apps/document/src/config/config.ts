export interface RemoteConfig {
  endpoint: string;
  insecure: boolean;
}

export interface TraceConfig {
  enabled: boolean;
  sampleRate: number;
  grpc: RemoteConfig;
  stdout: boolean;
}

export interface LogConfig {
  enabled: boolean;
  level: string;
  grpc: RemoteConfig;
  stdout: boolean;
}

export interface MeterConfig {
  enabled: boolean;
  grpc: RemoteConfig;
  stdout: boolean;
  exportInterval: number;
}

export interface OtelConfig {
  enabled: boolean;
  stdout: boolean;
  trace: TraceConfig;
  log: LogConfig;
  meter: MeterConfig;
}

export interface DatabaseConfig {
  host: string;
  port: number;
  username: string;
  password: string;
  database: string;
}

export interface AppConfig {
  env: string;
  database: DatabaseConfig;
  otel: OtelConfig;
}
