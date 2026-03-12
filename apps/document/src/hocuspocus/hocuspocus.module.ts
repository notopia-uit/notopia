import { forwardRef, Module } from '@nestjs/common';

import { DocumentModule } from '../document/document.module';
import { HocuspocusGateway } from './hocuspocus.controller';
import { HocuspocusProvider } from './hocuspocus.provider';

@Module({
  imports: [forwardRef(() => DocumentModule)],
  providers: [HocuspocusGateway, HocuspocusProvider],
  exports: [HocuspocusProvider],
})
export class HocuspocusModule {}
