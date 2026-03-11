import { Server } from '@hocuspocus/server';
import { Module } from '@nestjs/common';

import { DatabaseModule } from '../database/database.module';
import { HocuspocusProvider } from './server';

@Module({
  imports: [DatabaseModule],
  providers: [HocuspocusProvider],
  exports: [Server],
})
export class HocuspocusModule {}
