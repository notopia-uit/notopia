import { Connection } from '@hocuspocus/server';
import { Injectable } from '@nestjs/common';
import { Traceable } from 'nestjs-otel';

import { AuthorizationService } from '#/authorization';
import { HocuspocusContext } from '#/hocuspocus';
import { NoteService, WorkspaceModel, WorkspaceNoteNotFoundException } from '#/note';

import { Hocuspocus } from './hocuspocus';

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
  async onRoleChanged({ workspaceId, userId }: { workspaceId: string; userId: string }) {
    for (const [documentName, document] of this.hocuspocus.hocuspocus.documents) {
      for (const connection of document.getConnections() as Connection<HocuspocusContext>[]) {
        const context = connection.context as HocuspocusContext;
        if (context.user.id !== userId) {
          continue;
        }
        let workspace: WorkspaceModel | undefined;
        try {
          workspace = await this.noteService.getWorkspaceByNote({
            noteId: documentName,
            userId,
          });
        } catch (e) {
          if (e instanceof WorkspaceNoteNotFoundException) {
            connection.close({
              code: 4001,
              reason: `Your access to document ${documentName} has been revoked due to workspace be found for this document.`,
            });
            continue;
          }
        }
        if (workspace?.id !== workspaceId) {
          continue;
        }
        const permissions = await this.authorizationService.getWorkspaceItemPermissions({
          workspaceId,
          memberId: userId,
        });
        if (!permissions.canRead) {
          connection.close({
            code: 4002,
            reason:
              'Your access to this document has been revoked because you no longer have read permission.',
          });
        } else {
          connection.readOnly = !permissions.canWrite;
        }
      }
    }
  }

  // NOTE: If we handle the published, then this should be adjusted
  async onMemberRemoved({ workspaceId, userId }: { workspaceId: string; userId: string }) {
    for (const [documentName, document] of this.hocuspocus.hocuspocus.documents) {
      for (const connection of document.getConnections() as Connection<HocuspocusContext>[]) {
        const context = connection.context as HocuspocusContext;
        if (context.user.id !== userId) {
          continue;
        }
        let workspace: WorkspaceModel | undefined;
        try {
          workspace = await this.noteService.getWorkspaceByNote({
            noteId: documentName,
            userId,
          });
        } catch (e) {
          if (e instanceof WorkspaceNoteNotFoundException) {
            connection.close({
              code: 4001,
              reason: `Your access to document ${documentName} has been revoked due to workspace be found for this document.`,
            });
            continue;
          }
        }
        if (workspace?.id !== workspaceId) {
          continue;
        }
        connection.close({
          code: 4002,
          reason:
            'Your access to this document has been revoked because you have been removed from the workspace.',
        });
      }
    }
  }
}
