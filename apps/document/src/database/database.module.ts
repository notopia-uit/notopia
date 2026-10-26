import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';

import { DATABASE_CONFIG, DatabaseConfig } from '#/config';

import { createDatasourceOptions } from './database.provider';

@Module({
  imports: [
    TypeOrmModule.forRootAsync({
      inject: [ConfigService],
      useFactory: (configService: ConfigService) => {
        const databaseCfg = configService.get<DatabaseConfig>(DATABASE_CONFIG);
        if (!databaseCfg) {
          throw new Error('DATABASE_CONFIG not found');
        }
        return createDatasourceOptions({ databaseCfg });
      },
    }),
  ],
  exports: [TypeOrmModule],
})
export class DatabaseModule {}
