import { DataSourceOptions } from 'typeorm';
import { type BaseDataSourceOptions } from 'typeorm/data-source/BaseDataSourceOptions.js';

import { DatabaseConfig } from '#/config';
import { DocumentEntity } from '#/document';
import { RevisionEntity } from '#/revision';

export const createDatasourceOptions = ({
  databaseCfg,
  synchronize,
  logger = 'simple-console',
  logging = ['error'],
}: {
  databaseCfg: DatabaseConfig;
  synchronize: boolean;
  logger?: BaseDataSourceOptions['logger'];
  logging?: BaseDataSourceOptions['logging'];
}) => {
  return {
    type: 'postgres',
    host: databaseCfg.host,
    port: databaseCfg.port,
    username: databaseCfg.username,
    password: databaseCfg.password,
    database: databaseCfg.database,
    entities: [DocumentEntity, RevisionEntity],
    synchronize,
    logging,
    logger: logger,
  } satisfies DataSourceOptions;
};
