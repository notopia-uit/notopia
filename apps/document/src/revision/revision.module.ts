import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { DatabaseModule } from '#/database/database.module';

import { RevisionEntity } from './revision.entity';
import { RevisionService } from './revision.service';
import { RevisionApi } from './revision.api';

@Module({
  imports: [TypeOrmModule.forFeature([RevisionEntity]), DatabaseModule],
  providers: [RevisionService, RevisionApi],
  exports: [RevisionService],
})
export class RevisionModule {}