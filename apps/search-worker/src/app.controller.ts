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
    try {
      await this.appService.handleNoteCreated({
        id: data.id,
        name: data.name,
        workspaceId: data.workspaceId,
      });
    } catch (error) {
      this.logger.error(`handleNoteCreated: error occurred id=${data.id}`, error);
      throw error;
    }
  }

  @EventPattern('events.integration.note.note.updated')
  async handleNoteUpdated(@Payload() data: ShareNoteUpdatedEvent) {
    this.logger.log(`handleNoteUpdated: received id=${data.id}`);
    try {
      await this.appService.handleNoteUpdated({
        id: data.id,
        name: data.name,
        folderId: data.folderId,
        folderName: data.folderName,
        trashed: data.trashed,
      });
    } catch (error) {
      this.logger.error(`handleNoteUpdated: error occurred id=${data.id}`, error);
      throw error;
    }
    this.logger.log(`handleNoteUpdated: done id=${data.id}`);
  }

  @EventPattern('events.integration.document.document.committed')
  async handleDocumentCommitted(@Payload() data: ShareDocumentCommittedEvent) {
    this.logger.log(`handleDocumentCommitted: received id=${data.id}`);
    try {
      await this.appService.handleDocumentCommitted({
        id: data.id,
        tags: data.tags,
        // we should validate? or let blocknote throw error
        content: data.content as MyBlock[],
      });
    } catch (error) {
      this.logger.error(`handleDocumentCommitted: error occurred id=${data.id}`, error);
      throw error;
    }
  }
}
