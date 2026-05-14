import { Module } from '@nestjs/common';

import { AuthenticationModule } from '#/authentication/authentication.module';
import { AuthorizationModule } from '#/authorization/authorization.module';
import { DocumentModule } from '#/document/document.module';
import { HocuspocusController } from '#/hocuspocus/hocuspocus.controller';
import { HocuspocusService } from '#/hocuspocus/hocuspocus.service';
import { NoteModule } from '#/note/note.module';

import { HocuspocusGateway } from './hocuspocus.gateway';
import { HocuspocusProvider } from './hocuspocus.provider';

@Module({
  controllers: [HocuspocusController],
  imports: [NoteModule, AuthorizationModule, AuthenticationModule, DocumentModule],
  providers: [HocuspocusGateway, HocuspocusProvider, HocuspocusService],
  exports: [HocuspocusProvider, HocuspocusService],
})
export class HocuspocusModule {}
