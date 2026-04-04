import { getDatabaseConfig } from '../config/config.factory.ts';
import { createDatasource } from './database.provider.ts';

const config = getDatabaseConfig();
const datasource = await createDatasource(config, true);
export default datasource;
