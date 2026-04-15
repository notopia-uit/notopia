import { Controller } from '@nestjs/common';
import { MessagePattern, Payload } from '@nestjs/microservices';
import type { ShareUserWorkspaceRoleUpdatedEvent } from '@notopia-uit/api-gen';

import { HocuspocusService } from '#/hocuspocus/hococuspocus.service';

@Controller()
export class HocuspocusController {
  constructor(private readonly hocuspocusService: HocuspocusService) {}

  @MessagePattern('events.integration.note.user_workspace_role_updated')
  async handleOrderCreated(
    @Payload() data: ShareUserWorkspaceRoleUpdatedEvent
  ) {
    await this.hocuspocusService.onRoleChanged({
      workspaceId: data.workspaceId,
      userId: data.userId,
    });
  }
}
