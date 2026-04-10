import { DataSource, DataSourceOptions } from 'typeorm';

import { DatabaseConfig } from '../config/config';
import { DocumentEntity } from '../document/document.entity';
import { RevisionEntity } from '../revision/revision.entity';

export const createDatasourceOptions = (
  databaseCfg: DatabaseConfig,
  synchronize: boolean
) => {
  return {
    type: 'postgres',
    host: databaseCfg.host,
    port: databaseCfg.port,
    username: databaseCfg.username,
    password: databaseCfg.password,
    database: databaseCfg.database,
    entities: [DocumentEntity, RevisionEntity],
    synchronize,
    logging: true,
    logger: 'simple-console',
  } satisfies DataSourceOptions;
};

export const createDatasource = (
  databaseCfg: DatabaseConfig,
  synchronize: boolean
) => {
  const options = createDatasourceOptions(databaseCfg, synchronize);
  const dataSource = new DataSource(options);
  return dataSource;
};
