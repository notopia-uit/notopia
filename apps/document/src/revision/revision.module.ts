import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { RevisionController } from './revision.controller';
import { RevisionEntity } from './revision.entity';
import { RevisionRepository } from './revision.repository';
import { RevisionService } from './revision.service';

@Module({
  imports: [TypeOrmModule.forFeature([RevisionEntity])],
  providers: [RevisionRepository, RevisionService, RevisionController],
  exports: [RevisionService, RevisionRepository, RevisionController],
})
export class RevisionModule {}
