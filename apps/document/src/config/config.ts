export interface DatabaseConfig {
  host: string;
  port: number;
  username: string;
  password: string;
  database: string;
}

export interface AppConfig {
  env: 'test' | 'production' | 'development';
  logLevel: 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal';
}
