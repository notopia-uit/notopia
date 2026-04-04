import { DatabaseConfig } from '../config/config.ts';
import { DocumentEntity } from '../document/document.entity.ts';
import { RevisionEntity } from '../revision/revision.entity.ts';
import { DataSource, DataSourceOptions } from 'typeorm';

export const createDatasourceOptions = async (
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

export const createDatasource = async (
  databaseCfg: DatabaseConfig,
  synchronize: boolean
) => {
  const options = await createDatasourceOptions(databaseCfg, synchronize);
  const dataSource = new DataSource(options);
  await dataSource.initialize();
  return dataSource;
};
