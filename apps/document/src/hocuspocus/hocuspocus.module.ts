import { Module } from '@nestjs/common';

import { AuthenticationModule } from '../authentication/authentication.module';
import { AuthorizationModule } from '../authorization/authorization.module';
import { DocumentModule } from '../document/document.module';
import { NoteModule } from '../note/note.module';
import { Hocuspocus } from './hocuspocus';
import { HocuspocusController } from './hocuspocus.controller';
import { HocuspocusGateway } from './hocuspocus.gateway';
import { HocuspocusService } from './hocuspocus.service';

@Module({
  controllers: [HocuspocusController],
  imports: [NoteModule, AuthorizationModule, AuthenticationModule, DocumentModule],
  providers: [HocuspocusGateway, HocuspocusService, Hocuspocus],
  exports: [HocuspocusService, Hocuspocus],
})
export class HocuspocusModule {}
