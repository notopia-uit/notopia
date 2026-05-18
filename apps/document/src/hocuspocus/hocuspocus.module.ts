import { Module } from '@nestjs/common';

import { AuthenticationModule } from '#/authentication';
import { AuthorizationModule } from '#/authorization';
import { DocumentModule } from '#/document';
import { HocuspocusController } from '#/hocuspocus';
import { HocuspocusService } from '#/hocuspocus';
import { NoteModule } from '#/note';

import { Hocuspocus } from './hocuspocus';
import { HocuspocusGateway } from './hocuspocus.gateway';

@Module({
  controllers: [HocuspocusController],
  imports: [NoteModule, AuthorizationModule, AuthenticationModule, DocumentModule],
  providers: [HocuspocusGateway, HocuspocusService, Hocuspocus],
  exports: [HocuspocusService, Hocuspocus],
})
export class HocuspocusModule {}
