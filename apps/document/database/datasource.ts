import { DataSource } from 'typeorm';

import { getDatabaseConfig } from '#/config/config.factory';
import { createDatasourceOptions } from '#/database/database.provider';

const option = createDatasourceOptions({
  databaseCfg: getDatabaseConfig(),
  synchronize: true,
  logging: ['error', 'warn'],
});

const datasource = new DataSource(option);
export default datasource;
