import { Server } from '@hocuspocus/server';
import { Module } from '@nestjs/common';

import { DocumentModule } from '../document/document.module';
import { HocuspocusGateway } from './hocuspocus.controller';
import { HocuspocusProvider } from './hocuspocus.provider';

@Module({
  imports: [DocumentModule],
  providers: [HocuspocusGateway, HocuspocusProvider],
  exports: [Server],
})
export class HocuspocusModule {}
