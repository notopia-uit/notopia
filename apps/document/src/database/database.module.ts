import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';

import { AppConfig, DatabaseConfig } from '../config/config';
import { createDatasourceOptions } from './database.provider';

@Module({
  imports: [
    TypeOrmModule.forRootAsync({
      inject: [ConfigService],
      useFactory: async (configService: ConfigService) => {
        const appConfig = configService.get<AppConfig>('app')!;
        const databaseConfig = configService.get<DatabaseConfig>('database')!;
        return await createDatasourceOptions(
          databaseConfig,
          appConfig.env !== 'production'
        );
      },
    }),
  ],
  exports: [TypeOrmModule],
})
export class DatabaseModule {}
