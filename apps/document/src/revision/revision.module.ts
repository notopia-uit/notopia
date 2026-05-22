import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { DatabaseModule } from '#/database/database.module';

import { RevisionApi } from './revision.api';
import { RevisionEntity } from './revision.entity';
import { RevisionService } from './revision.service';

@Module({
  imports: [TypeOrmModule.forFeature([RevisionEntity]), DatabaseModule],
  providers: [RevisionService, RevisionApi],
  exports: [RevisionService],
})
export class RevisionModule {}
