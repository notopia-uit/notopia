import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { ApiModule } from '@notopia-uit/api-document-nestjs-server';

import { DocumentController } from './document/document.controller';
import { DocumentEntity } from './document/document.entity';
import { DocumentRepository } from './document/document.repository';
import { DocumentService } from './document/document.service';
import { RevisionController } from './revision/revision.controller';
import { RevisionEntity } from './revision/revision.entity';
import { RevisionRepository } from './revision/revision.repository';
import { RevisionService } from './revision/revision.service';

@Module({
  imports: [
    TypeOrmModule.forFeature([DocumentEntity, RevisionEntity]),
    Object.assign(
      ApiModule.forRoot({
        apiImplementations: {
          documentApi: DocumentController,
          revisionApi: RevisionController,
        },
        providers: [
          DocumentRepository,
          DocumentService,
          RevisionRepository,
          RevisionService,
          {
            provide: ServerBlockNoteEditor,
            useFactory: () => ServerBlockNoteEditor.create(),
          },
        ],
      }),
      {
        imports: [TypeOrmModule.forFeature([DocumentEntity, RevisionEntity])],
      }
    ),
  ],
})
export class ApiHttpModule {}
