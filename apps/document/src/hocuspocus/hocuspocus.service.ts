import { Hocuspocus } from '@hocuspocus/server';
import { Injectable } from '@nestjs/common';
import { Traceable } from 'nestjs-otel';

import { AuthorizationService } from '#/authorization/authorization.service';
import { HocuspocusContext } from '#/hocuspocus/hocuspocus-context';
import { NoteService } from '#/note/note.service';

@Injectable()
@Traceable()
export class HocuspocusService {
  constructor(
    private readonly noteService: NoteService,
    private readonly hocuspocus: Hocuspocus,
    private readonly authorizationService: AuthorizationService
  ) {}

  // NOTE: We don't really need to check the permission canRead, because yeah
  // Because based on the casbin rules, it will be always canRead when this event fired
  async onRoleChanged({
    workspaceId,
    userId,
  }: {
    workspaceId: string;
    userId: string;
  }) {
    for (const [documentName, document] of this.hocuspocus.documents) {
      for (const [_, connection] of document.connections) {
        const context = connection.connection.context as HocuspocusContext;
        if (context.user.id !== userId) {
          continue;
        }
        const workspace = await this.noteService.getWorkspaceByNote({
          noteId: documentName,
          userId,
        });
        if (workspace.id !== workspaceId) {
          continue;
        }
        const permissions =
          await this.authorizationService.getWorkspaceItemPermissions({
            workspaceId,
            memberId: userId,
          });
        if (!permissions.canRead) {
          connection.connection.close({
            code: 4001,
            reason: 'Your access to this document has been revoked',
          });
        } else {
          connection.connection.readOnly = !permissions.canWrite;
        }
      }
    }
  }

  // NOTE: If we handle the published, then this should be adjusted
  async onMemberRemoved({
    workspaceId,
    userId,
  }: {
    workspaceId: string;
    userId: string;
  }) {
    for (const [documentName, document] of this.hocuspocus.documents) {
      for (const [_, connection] of document.connections) {
        const context = connection.connection.context as HocuspocusContext;
        if (context.user.id !== userId) {
          continue;
        }
        const workspace = await this.noteService.getWorkspaceByNote({
          noteId: documentName,
          userId,
        });
        if (workspace.id !== workspaceId) {
          continue;
        }
        connection.connection.close({
          code: 4001,
          reason: 'Your access to this document has been revoked',
        });
      }
    }
  }
}
