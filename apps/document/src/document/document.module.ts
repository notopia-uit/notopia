import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { AuthorizationModule } from '#/authorization';
import { BlockNoteModule } from '#/blocknote';
import { DatabaseModule } from '#/database';
import { KafkaModule } from '#/kafka';
import { StorageModule } from '#/storage';

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
