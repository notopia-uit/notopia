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

const typeOrmFeature = TypeOrmModule.forFeature([
  DocumentEntity,
  RevisionEntity,
]);

const apiDynamicModule = ApiModule.forRoot({
  apiImplementations: {
    documentApi: DocumentApiImpl,
    revisionApi: RevisionApiImpl,
  },
  providers: [
    DocumentRepository,
    RevisionRepository,
    DocumentService,
    RevisionService,
  ],
});

// ApiModule.forRoot() doesn't expose an `imports` option, so we extend the
// returned DynamicModule to inject TypeOrmModule.forFeature() into its scope.
// This makes TypeORM entity repositories available to DocumentRepository /
// RevisionRepository, which are both provided (and instantiated) inside ApiModule.
const apiModuleWithTypeOrm = {
  ...apiDynamicModule,
  imports: [typeOrmFeature],
};

@Module({
  imports: [apiModuleWithTypeOrm],
})
export class DocumentModule {}
