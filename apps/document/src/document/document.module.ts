import { Module } from '@nestjs/common';
import { ClientsModule } from '@nestjs/microservices';
import { TypeOrmModule } from '@nestjs/typeorm';

import { DatabaseModule } from '#/database/database.module';
import { AuthorizationModule } from '#/authorization/authorization.module';
import { BlockNoteModule } from '#/blocknote/blocknote.module';
import { StorageModule } from '#/storage/storage.module';
import { KAFKA_CLIENT } from '#/common/token';
import { getKafkaConfig } from '#/config/kafka.config';

import { DocumentEntity } from './document.entity';
import { DocumentService } from './document.service';

@Module({
  imports: [
    TypeOrmModule.forFeature([DocumentEntity]),
    ClientsModule.registerAsync([{ name: KAFKA_CLIENT, useFactory: getKafkaConfig }]),
    DatabaseModule,
    AuthorizationModule,
    BlockNoteModule,
    StorageModule,
  ],
  providers: [DocumentService],
  exports: [DocumentService],
})
export class DocumentModule {}