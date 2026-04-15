import { Module } from '@nestjs/common';

import { AuthorizationModule } from '#/authorization/authorization.module';
import { HocuspocusService } from '#/hocuspocus/hococuspocus.service';
import { HocuspocusController } from '#/hocuspocus/hocuspocus.controller';
import { NoteModule } from '#/note/note.module';

import { HocuspocusGateway } from './hocuspocus.gateway';
import { HocuspocusProvider } from './hocuspocus.provider';

@Module({
  controllers: [HocuspocusController],
  imports: [NoteModule, AuthorizationModule],
  providers: [HocuspocusGateway, HocuspocusProvider, HocuspocusService],
  exports: [HocuspocusProvider],
})
export class HocuspocusModule {}
