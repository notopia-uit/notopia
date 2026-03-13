import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { AuthorizationModule } from '../authorization/authorization.module';
import { S3Module } from '../s3/s3.module';
import { DocumentApi } from './document.api';
import { DocumentEntity } from './document.entity';
import { DocumentRepository } from './document.repository';
import { DocumentService } from './document.service';

@Module({
  imports: [
    TypeOrmModule.forFeature([DocumentEntity]),
    S3Module,
    AuthorizationModule,
  ],
  providers: [
    DocumentRepository,
    DocumentService,
    DocumentApi,
    {
      provide: ServerBlockNoteEditor,
      useFactory: () => ServerBlockNoteEditor.create(),
    },
  ],
  exports: [DocumentService, DocumentRepository, DocumentApi],
})
export class DocumentModule {}
