import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';

import { AppConfig, DatabaseConfig } from '../config/config';
import { APP_CONFIG, DATABASE_CONFIG } from '../config/config.factory';

import { createDatasourceOptions } from './database.provider';

@Module({
  imports: [
    TypeOrmModule.forRootAsync({
      inject: [ConfigService],
      useFactory: (configService: ConfigService) => {
        const appConfig = configService.get<AppConfig>(APP_CONFIG);
        if (!appConfig) {
          throw new Error('APP_CONFIG not found');
        }
        const databaseCfg = configService.get<DatabaseConfig>(DATABASE_CONFIG);
        if (!databaseCfg) {
          throw new Error('DATABASE_CONFIG not found');
        }
        return createDatasourceOptions({
          databaseCfg,
          synchronize: appConfig.env !== 'production',
        });
      },
    }),
  ],
  exports: [TypeOrmModule],
})
export class DatabaseModule {}
