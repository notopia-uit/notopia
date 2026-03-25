import { AppConfig, DatabaseConfig } from '../config/config';
import { APP_CONFIG, DATABASE_CONFIG } from '../config/config.factory';
import { createDatasourceOptions } from './database.provider';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';

@Module({
  imports: [
    TypeOrmModule.forRootAsync({
      inject: [ConfigService],
      useFactory: async (configService: ConfigService) => {
        const appConfig = configService.get<AppConfig>(APP_CONFIG)!;
        const databaseConfig =
          configService.get<DatabaseConfig>(DATABASE_CONFIG)!;
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
