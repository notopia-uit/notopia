import { Module } from '@nestjs/common';

import { AuthorizationModule } from '../authorization/authorization.module';
import { NoteModule } from '../note/note.module';
import { SearchModule } from '../search/search.module';
import { AiController } from './ai.controller';
import { AiService } from './ai.service';

@Module({
  imports: [NoteModule, SearchModule, AuthorizationModule],
  controllers: [AiController],
  providers: [AiService],
  exports: [AiService],
})
export class AiModule {}
