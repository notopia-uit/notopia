import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { AuthorizationModule } from '../authorization/authorization.module';
import { BlockNoteModule } from '../blocknote/blocknote.module';
import { DatabaseModule } from '../database/database.module';
import { KafkaModule } from '../kafka/kafka.module';
import { StorageModule } from '../storage/storage.module';

import { DocumentEntity } from './document.entity';
import { DocumentService } from './document.service';

@Module({
  imports: [
    TypeOrmModule.forFeature([DocumentEntity]),
    KafkaModule,
    DatabaseModule,
    AuthorizationModule,
    BlockNoteModule,
    StorageModule,
  ],
  providers: [DocumentService],
  exports: [DocumentService],
})
export class DocumentModule {}
