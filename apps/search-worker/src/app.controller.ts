import type { MyBlock } from '@blocknote/core';
import { Controller, Logger } from '@nestjs/common';
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
  private readonly logger = new Logger(AppController.name);

  constructor(private readonly appService: AppService) {}

  @EventPattern('events.integration.note.note.created')
  async handleNoteCreated(@Payload() data: ShareNoteCreatedEvent) {
    this.logger.log(`handleNoteCreated: received id=${data.id} workspaceId=${data.workspaceId}`);
    await this.appService.handleNoteCreated({
      id: data.id,
      name: data.name,
      workspaceId: data.workspaceId,
    });
    this.logger.log(`handleNoteCreated: done id=${data.id}`);
  }

  @EventPattern('events.integration.note.note.updated')
  async handleNoteUpdated(@Payload() data: ShareNoteUpdatedEvent) {
    this.logger.log(`handleNoteUpdated: received id=${data.id}`);
    await this.appService.handleNoteUpdated({
      id: data.id,
      name: data.name,
      folderId: data.folderId,
      folderName: data.folderName,
      trashed: data.trashed,
    });
    this.logger.log(`handleNoteUpdated: done id=${data.id}`);
  }

  @EventPattern('events.integration.document.document.committed')
  async handleDocumentCommitted(@Payload() data: ShareDocumentCommittedEvent) {
    this.logger.log(`handleDocumentCommitted: received id=${data.id}`);
    await this.appService.handleDocumentCommitted({
      id: data.id,
      tags: data.tags,
      // we should validate? or let blocknote throw error
      content: data.content as MyBlock[],
    });
    this.logger.log(`handleDocumentCommitted: done id=${data.id}`);
  }
}
