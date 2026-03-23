export interface AppConfig {
  env: 'test' | 'production' | 'development';
  logLevel: 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal';
  port: number;
}

export interface KafkaConfig {
  clientId: string;
  brokers: string[];
  groupId: string;
}

export interface MeiliConfig {
  host: string;
  apiKey: string;
}
