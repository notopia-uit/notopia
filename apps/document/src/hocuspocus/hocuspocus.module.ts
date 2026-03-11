import { Server } from '@hocuspocus/server';
import { Module } from '@nestjs/common';

import { DocumentModule } from '../document/document.module';
import { HocuspocusProvider } from './hocuspocus.provider';

@Module({
  imports: [DocumentModule],
  providers: [HocuspocusProvider],
  exports: [Server],
})
export class HocuspocusModule {}
