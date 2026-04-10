import { Module } from '@nestjs/common';

import { AuthorizationModule } from '#/authorization/authorization.module';
import { NoteModule } from '#/note/note.module';

import { HocuspocusGateway } from './hocuspocus.gateway';
import { HocuspocusProvider } from './hocuspocus.provider';

@Module({
  imports: [NoteModule, AuthorizationModule],
  providers: [HocuspocusGateway, HocuspocusProvider],
  exports: [HocuspocusProvider],
})
export class HocuspocusModule {}
