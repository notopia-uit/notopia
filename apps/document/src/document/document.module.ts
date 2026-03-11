import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { DocumentController } from './document.controller';
import { DocumentEntity } from './document.entity';
import { DocumentRepository } from './document.repository';
import { DocumentService } from './document.service';

@Module({
  imports: [TypeOrmModule.forFeature([DocumentEntity])],
  providers: [
    DocumentRepository,
    DocumentService,
    DocumentController,
    {
      provide: ServerBlockNoteEditor,
      useFactory: () => ServerBlockNoteEditor.create(),
    },
  ],
  exports: [DocumentService, DocumentRepository, DocumentController],
})
export class DocumentModule {}
