// oxfmt-ignore
import { otelSdk } from './otel';
// oxfmt-ignore
import 'reflect-metadata';

import pino from 'pino';
import { DataSource } from 'typeorm';

import { migrations } from '../database/migrations';
import { getDatabaseConfig } from './config/config.factory';
import { createDatasourceOptions } from './database/database.provider';
import { runMigration } from './migration-runner';

const logger = pino();
async function main() {
  try {
    const datasource = new DataSource({
      ...createDatasourceOptions({ databaseCfg: getDatabaseConfig() }),
      migrations,
    });
    process.exitCode = await runMigration(datasource, otelSdk, logger);
  } catch (error) {
    logger.error({ err: error }, 'Migration setup failed');
    process.exitCode = 1;
    try {
      await otelSdk.shutdown();
    } catch (shutdownError) {
      logger.error({ err: shutdownError }, 'Telemetry shutdown failed');
    }
  }
}
void main();
