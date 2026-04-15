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
        } else if (!permissions.canWrite) {
          connection.connection.readOnly = true;
        }
      }
    }
  }
}
