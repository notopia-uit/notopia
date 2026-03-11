import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ApiModule } from '@notopia-uit/api-document-nestjs-server';

import { DocumentService } from '../application/document.service';
import { RevisionService } from '../application/revision.service';
import { DocumentEntity } from '../domain/document.entity';
import { RevisionEntity } from '../domain/revision.entity';
import { DocumentRepository } from '../infrastructure/database/document.repository';
import { RevisionRepository } from '../infrastructure/database/revision.repository';
import { DocumentApiImpl } from './document.api.impl';
import { RevisionApiImpl } from './revision.api.impl';

@Module({
  imports: [
    TypeOrmModule.forFeature([DocumentEntity, RevisionEntity]),
    ApiModule.forRoot({
      apiImplementations: {
        documentApi: DocumentApiImpl,
        revisionApi: RevisionApiImpl,
      },
    }),
  ],
  providers: [
    DocumentRepository,
    RevisionRepository,
    DocumentService,
    RevisionService,
    DocumentApiImpl,
    RevisionApiImpl,
  ],
})
export class DocumentModule {}
