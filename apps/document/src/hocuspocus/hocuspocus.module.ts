import { Module } from '@nestjs/common';

import { AuthenticationModule } from '../authentication/authentication.module';
import { AuthorizationModule } from '../authorization/authorization.module';
import { DocumentModule } from '../document/document.module';
import { HocuspocusController } from './hocuspocus.controller';
import { HocuspocusService } from './hocuspocus.service';
import { NoteModule } from '../note/note.module';

import { Hocuspocus } from './hocuspocus';
import { HocuspocusGateway } from './hocuspocus.gateway';

@Module({
  controllers: [HocuspocusController],
  imports: [NoteModule, AuthorizationModule, AuthenticationModule, DocumentModule],
  providers: [HocuspocusGateway, HocuspocusService, Hocuspocus],
  exports: [HocuspocusService, Hocuspocus],
})
export class HocuspocusModule {}
