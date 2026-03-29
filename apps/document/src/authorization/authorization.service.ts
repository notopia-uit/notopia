import { NoteService } from '../note/note.service';
import { Client, Code, ConnectError } from '@connectrpc/connect';
import {
  Injectable,
  InternalServerErrorException,
  UnprocessableEntityException,
} from '@nestjs/common';
import {
  AuthorizationService as AuthorizationServiceDefinition,
  WorkspaceItemPermission,
} from '@notopia-uit/pb/authorization';

export type AuthorizationClient = Client<typeof AuthorizationServiceDefinition>;

export type UserNotePermissions = 'read' | 'write' | 'delete';

@Injectable()
export class AuthorizationService {
  constructor(
    private readonly authorizationClient: AuthorizationClient,
    private readonly noteService: NoteService
  ) {}

  private toWorkspaceItemPermission(
    permission: UserNotePermissions
  ): WorkspaceItemPermission {
    switch (permission) {
      case 'read':
        return WorkspaceItemPermission.READ;
      case 'write':
        return WorkspaceItemPermission.WRITE;
      case 'delete':
        return WorkspaceItemPermission.DELETE;
      default:
        throw new UnprocessableEntityException(
          `Invalid permission: ${permission}`
        );
    }
  }

  async hasNotePermission({
    memberId,
    documentId,
    permission,
  }: {
    memberId: string;
    documentId: string;
    permission: UserNotePermissions;
  }): Promise<boolean> {
    try {
      const workspaceId =
        await this.noteService.getWorkspaceIdByNoteId(documentId);
      const response =
        await this.authorizationClient.hasWorkspaceItemPermission({
          permission: this.toWorkspaceItemPermission(permission),
          memberId,
          workspaceId,
        });
      return response.hasPermission;
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.NotFound) {
          return false;
        }
        throw new InternalServerErrorException(
          `Failed to check note permission: ${error.message}`
        );
      }
      throw error;
    }
  }

  async getUserNotePermissions(
    memberId: string,
    documentId: string
  ): Promise<{ canRead: boolean; canWrite: boolean; canDelete: boolean }> {
    try {
      const workspaceId =
        await this.noteService.getWorkspaceIdByNoteId(documentId);
      return await this.authorizationClient.getUserWorkspaceItemPermissions({
        memberId,
        workspaceId,
      });
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.NotFound) {
          return { canRead: false, canWrite: false, canDelete: false };
        }
        throw new InternalServerErrorException(
          `Failed to get user note permissions: ${error.message}`
        );
      }
      throw error;
    }
  }
}
