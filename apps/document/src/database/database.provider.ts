import { DatabaseConfig } from '../config/config';
import { DocumentEntity } from '../document/document.entity';
import { RevisionEntity } from '../revision/revision.entity';
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
