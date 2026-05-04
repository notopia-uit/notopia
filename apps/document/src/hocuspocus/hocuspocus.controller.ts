import { Controller } from '@nestjs/common';
import { MessagePattern, Payload } from '@nestjs/microservices';
import type {
  ShareUserWorkspaceRoleUpdatedEvent,
  ShareWorkspaceMemberRemovedEvent,
} from '@notopia-uit/api-gen';

import { HocuspocusService } from '#/hocuspocus/hocuspocus.service';

@Controller()
export class HocuspocusController {
  constructor(private readonly hocuspocusService: HocuspocusService) {}

  @MessagePattern('events.integration.authorization.user_workspace_role_updated')
  async handleUserWorkspaceRoleUpdated(@Payload() data: ShareUserWorkspaceRoleUpdatedEvent) {
    await this.hocuspocusService.onRoleChanged({
      workspaceId: data.workspaceId,
      userId: data.userId,
    });
  }

  @MessagePattern('events.integration.authorization.workspace_member_removed')
  async handleWorkspaceMemberRemoved(@Payload() data: ShareWorkspaceMemberRemovedEvent) {
    await this.hocuspocusService.onMemberRemoved({
      workspaceId: data.workspaceId,
      userId: data.userId,
    });
  }
}
