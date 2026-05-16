import { Controller, Logger } from '@nestjs/common';
import { MessagePattern, Payload } from '@nestjs/microservices';
import type {
  ShareUserWorkspaceRoleUpdatedEvent,
  ShareWorkspaceMemberRemovedEvent,
} from '@notopia-uit/api-gen';

import { HocuspocusService } from '#/hocuspocus/hocuspocus.service';

@Controller()
export class HocuspocusController {
  private readonly logger = new Logger(HocuspocusController.name);

  constructor(private readonly hocuspocusService: HocuspocusService) {}

  @MessagePattern('events.integration.authorization.user_workspace_role_updated')
  async handleUserWorkspaceRoleUpdated(@Payload() data: ShareUserWorkspaceRoleUpdatedEvent) {
    this.logger.log(
      { workspaceId: data.workspaceId, userId: data.userId },
      'handleUserWorkspaceRoleUpdated: received'
    );
    await this.hocuspocusService.onRoleChanged({
      workspaceId: data.workspaceId,
      userId: data.userId,
    });
    this.logger.log(
      { workspaceId: data.workspaceId, userId: data.userId },
      'handleUserWorkspaceRoleUpdated: done'
    );
  }

  @MessagePattern('events.integration.authorization.workspace_member_removed')
  async handleWorkspaceMemberRemoved(@Payload() data: ShareWorkspaceMemberRemovedEvent) {
    this.logger.log(
      { workspaceId: data.workspaceId, userId: data.userId },
      'handleWorkspaceMemberRemoved: received'
    );
    await this.hocuspocusService.onMemberRemoved({
      workspaceId: data.workspaceId,
      userId: data.userId,
    });
    this.logger.log(
      { workspaceId: data.workspaceId, userId: data.userId },
      'handleWorkspaceMemberRemoved: done'
    );
  }
}
