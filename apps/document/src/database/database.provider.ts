import { DataSourceOptions } from 'typeorm';
import { type BaseDataSourceOptions } from 'typeorm/data-source/BaseDataSourceOptions.js';

import { DatabaseConfig } from '#/config';

import { DocumentEntity } from '../document/document.entity';
import { RevisionEntity } from '../revision/revision.entity';

export const createDatasourceOptions = ({
  databaseCfg,
  logger = 'simple-console',
  logging = ['error'],
}: {
  databaseCfg: DatabaseConfig;
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
    synchronize: false,
    logging,
    logger: logger,
  } satisfies DataSourceOptions;
};
