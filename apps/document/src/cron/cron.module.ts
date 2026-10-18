import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { BillingModule } from '../billing/billing.module';
import { DocumentEntity } from '../document/document.entity';
import { DocumentModule } from '../document/document.module';
import { HocuspocusModule } from '../hocuspocus/hocuspocus.module';
import { NoteModule } from '../note/note.module';
import { RevisionEntity } from '../revision/revision.entity';
import { CronRegistrar } from './cron-registrar';
import { RevisionPurgeCron } from './revision-purge.cron';
import { SweeperCron } from './sweeper.cron';

@Module({
  imports: [
    TypeOrmModule.forFeature([DocumentEntity, RevisionEntity]),
    DocumentModule,
    HocuspocusModule,
    NoteModule,
    BillingModule,
  ],
  providers: [SweeperCron, RevisionPurgeCron, CronRegistrar],
})
export class CronModule {}
