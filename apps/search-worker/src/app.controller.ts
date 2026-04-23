import type { MyBlock } from '@blocknote/core';
import { Controller } from '@nestjs/common';
import { EventPattern, Payload } from '@nestjs/microservices';
import type {
  ShareDocumentCommittedEvent,
  ShareNoteCreatedEvent,
  ShareNoteUpdatedEvent,
} from '@notopia-uit/api-gen';

import { AppService } from './app.service';

// TODO: blocknote editor from schema, from the raw payload json. Then to html
// TODO: Suggest add a class implement the event (dto?) for validator
@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  @EventPattern('events.integration.note.note.created')
  handleNoteCreated(@Payload() data: ShareNoteCreatedEvent) {
    return this.appService.handleNoteCreated({
      id: data.id,
      name: data.name,
      workspaceId: data.workspaceId,
    });
  }

  @EventPattern('events.integration.note.note.updated')
  handleNoteUpdated(@Payload() data: ShareNoteUpdatedEvent) {
    return this.appService.handleNoteUpdated({
      id: data.id,
      name: data.name,
      folderId: data.folderId,
      folderName: data.folderName,
      trashed: data.trashed,
    });
  }

  @EventPattern('events.integration.document.document.committed')
  handleDocumentCommitted(@Payload() data: ShareDocumentCommittedEvent) {
    return this.appService.handleDocumentCommitted({
      id: data.id,
      tags: data.tags,
      // we should validate? or let blocknote throw error
      content: data.content as MyBlock[],
    });
  }
}
