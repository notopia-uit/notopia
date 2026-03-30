export interface AppConfig {
  env: 'test' | 'production' | 'development';
  logLevel: 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'fatal';
  port: number;
  apiUrl: string;
}

export interface DatabaseConfig {
  host: string;
  port: number;
  username: string;
  password: string;
  database: string;
}

export interface ServicesConfig {
  noteUrl: string;
  authorizationUrl: string;
}

export interface S3Config {
  endpoint: string;
  region: string;
  accessKeyId: string;
  secretAccessKey: string;
  bucketName: string;
}
