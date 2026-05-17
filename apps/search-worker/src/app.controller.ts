import type { MyBlock } from '@blocknote/core';
import { Controller, Logger } from '@nestjs/common';
import { EventPattern, Payload } from '@nestjs/microservices';
import type {
  DocumentCommittedEvent,
  NoteCreatedEvent,
  NoteUpdatedEvent,
} from '@notopia-uit/api-share-gen';

import { AppService } from './app.service';

// TODO: blocknote editor from schema, from the raw payload json. Then to html
// TODO: Suggest add a class implement the event (dto?) for validator
@Controller()
export class AppController {
  private readonly logger = new Logger(AppController.name);

  constructor(private readonly appService: AppService) {}

  @EventPattern('events.integration.note.note.created')
  async handleNoteCreated(@Payload() data: NoteCreatedEvent) {
    this.logger.log({ id: data.id, workspaceId: data.workspaceId }, 'handleNoteCreated: received');
    try {
      await this.appService.handleNoteCreated({
        id: data.id,
        name: data.name,
        workspaceId: data.workspaceId,
      });
    } catch (error) {
      this.logger.error({ err: error, id: data.id }, 'handleNoteCreated: error');
      throw error;
    }
  }

  @EventPattern('events.integration.note.note.updated')
  async handleNoteUpdated(@Payload() data: NoteUpdatedEvent) {
    this.logger.log({ id: data.id }, 'handleNoteUpdated: received');
    try {
      await this.appService.handleNoteUpdated({
        id: data.id,
        name: data.name,
        folderId: data.folderId,
        folderName: data.folderName,
        trashed: data.trashed,
      });
    } catch (error) {
      this.logger.error({ err: error, id: data.id }, 'handleNoteUpdated: error');
      throw error;
    }
    this.logger.log({ id: data.id }, 'handleNoteUpdated: done');
  }

  @EventPattern('events.integration.document.document.committed')
  async handleDocumentCommitted(@Payload() data: DocumentCommittedEvent) {
    this.logger.log({ id: data.id }, 'handleDocumentCommitted: received');
    try {
      await this.appService.handleDocumentCommitted({
        id: data.id,
        tags: data.tags,
        // we should validate? or let blocknote throw error
        content: data.content as MyBlock[],
      });
    } catch (error) {
      this.logger.error({ err: error, id: data.id }, 'handleDocumentCommitted: error');
      throw error;
    }
  }
}
