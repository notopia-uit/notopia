import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';

import { AppConfig, DatabaseConfig } from '../config/config';
import { DocumentEntity } from '../document/document.entity';
import { RevisionEntity } from '../revision/revision.entity';

@Module({
  imports: [
    TypeOrmModule.forRootAsync({
      inject: [ConfigService],
      useFactory: (configService: ConfigService) => {
        const appConfig = configService.get<AppConfig>('app');
        const databaseConfig = configService.get<DatabaseConfig>('database');
        if (!appConfig) {
          throw new Error('App configuration not found');
        }
        if (!databaseConfig) {
          throw new Error('Database configuration not found');
        }
        return {
          type: 'postgres',
          host: databaseConfig.host,
          port: databaseConfig.port,
          username: databaseConfig.username,
          password: databaseConfig.password,
          database: databaseConfig.database,
          entities: [DocumentEntity, RevisionEntity],
          synchronize: appConfig.env === 'development',
        };
      },
    }),
  ],
  exports: [TypeOrmModule],
})
export class DatabaseModule {}
