import { getDatabaseConfig } from '../config/config.factory';
import { createDatasource } from './database.provider';

const config = getDatabaseConfig();
const datasource = createDatasource(config, true);
export default datasource;
