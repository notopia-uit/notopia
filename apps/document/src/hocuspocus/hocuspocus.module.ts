import { AuthorizationModule } from '../authorization/authorization.module';
import { DocumentModule } from '../document/document.module';
import { NoteModule } from '../note/note.module';
import { HocuspocusGateway } from './hocuspocus.gateway';
import { HocuspocusProvider } from './hocuspocus.provider';
import { Module } from '@nestjs/common';

@Module({
  imports: [DocumentModule, NoteModule, AuthorizationModule],
  providers: [HocuspocusGateway, HocuspocusProvider],
  exports: [HocuspocusProvider],
})
export class HocuspocusModule {}
