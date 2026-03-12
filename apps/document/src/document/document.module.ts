import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { forwardRef, Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { HocuspocusModule } from '../hocuspocus/hocuspocus.module';
import { DocumentController } from './document.controller';
import { DocumentEntity } from './document.entity';
import { DocumentRepository } from './document.repository';
import { DocumentService } from './document.service';

@Module({
  imports: [
    TypeOrmModule.forFeature([DocumentEntity]),
    forwardRef(() => HocuspocusModule),
  ],
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
