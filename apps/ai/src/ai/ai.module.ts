import { Module } from '@nestjs/common';

import { AuthorizationModule } from '../authorization/authorization.module';
import { BillingModule } from '../billing/billing.module';
import { NoteModule } from '../note/note.module';
import { SearchModule } from '../search/search.module';
import { AssistantTaskService } from '../usage/assistant-task.service';
import { UsageService } from '../usage/usage.service';
import { AiController } from './ai.controller';
import { AiService } from './ai.service';

@Module({
  imports: [NoteModule, SearchModule, AuthorizationModule, BillingModule],
  controllers: [AiController],
  providers: [AiService, UsageService, AssistantTaskService],
  exports: [AiService, UsageService],
})
export class AiModule {}
