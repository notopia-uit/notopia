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
    return this.appService.indexNote({
      id: data.id,
      name: data.name,
    });
  }

  @EventPattern('events.integration.note.note.updated')
  handleNoteUpdated(@Payload() data: ShareNoteUpdatedEvent) {
    return this.appService.indexNote({
      id: data.id,
      name: data.name,
      tags: data.tags,
    });
  }

  @EventPattern('events.integration.document.document.committed')
  handleDocumentCommitted(@Payload() data: ShareDocumentCommittedEvent) {
    return this.appService.indexNote({
      id: data.id,
      tags: data.tags,
      // FIXME: addd!!! currently it send blocknote model, we need to transform it using editor
      content: 'please add',
    });
  }
}
