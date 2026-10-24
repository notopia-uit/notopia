import path from 'node:path';

import { DataSource } from 'typeorm';

import { getDatabaseConfig } from '#/config/config.factory';
import { createDatasourceOptions } from '#/database/database.provider';

const option = {
  ...createDatasourceOptions({
    databaseCfg: getDatabaseConfig(),
    logging: ['error', 'warn'],
  }),
  migrations: [path.join(import.meta.dirname, 'migrations', '*.ts')],
};

const datasource = new DataSource(option);
export default datasource;
