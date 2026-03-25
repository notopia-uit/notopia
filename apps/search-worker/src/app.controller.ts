import { AppService } from './app.service';
import { Controller } from '@nestjs/common';
import { EventPattern, Payload } from '@nestjs/microservices';
import type {
  ShareDocumentCommittedEvent,
  ShareNoteCreatedEvent,
  ShareNoteUpdatedEvent,
} from '@notopia-uit/api-gen';

@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}
  @EventPattern('event.NoteCreated')
  handleNoteCreated(@Payload() data: ShareNoteCreatedEvent) {
    return this.appService.indexNote({
      id: data.id,
      name: data.name,
    });
  }

  @EventPattern('event.NoteUpdated')
  handleNoteUpdated(@Payload() data: ShareNoteUpdatedEvent) {
    return this.appService.indexNote({
      id: data.id,
      name: data.name,
      tags: data.tags,
    });
  }

  @EventPattern('event.DocumentCommitted')
  handleDocumentCommitted(@Payload() data: ShareDocumentCommittedEvent) {
    return this.appService.indexNote({
      id: data.id,
      tags: data.tags,
      content: 'please add',
    });
  }
}
