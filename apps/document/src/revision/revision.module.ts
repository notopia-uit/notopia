import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { RevisionApi } from './revision.api';
import { RevisionEntity } from './revision.entity';
import { RevisionRepository } from './revision.repository';
import { RevisionService } from './revision.service';

@Module({
  imports: [TypeOrmModule.forFeature([RevisionEntity])],
  providers: [RevisionRepository, RevisionService, RevisionApi],
  exports: [RevisionService, RevisionRepository, RevisionApi],
})
export class RevisionModule {}
