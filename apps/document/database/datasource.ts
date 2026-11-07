import { DataSource } from 'typeorm';

import { getDatabaseConfig } from '#/config/config.factory';
import { createDatasourceOptions } from '#/database/database.provider';

import { migrations } from './migrations';

const option = {
  ...createDatasourceOptions({
    databaseCfg: getDatabaseConfig(),
    logging: ['error', 'warn'],
  }),
  migrations,
};

const datasource = new DataSource(option);
export default datasource;
