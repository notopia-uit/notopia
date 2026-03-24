import { Client, Code, ConnectError } from '@connectrpc/connect';
import { Injectable, InternalServerErrorException } from '@nestjs/common';
import {
  AuthorizationService as AuthorizationServiceDefinition,
  WorkspaceItemPermission,
} from '@notopia-uit/pb/authorization';

export type AuthorizationClient = Client<typeof AuthorizationServiceDefinition>;

@Injectable()
export class AuthorizationService {
  constructor(private readonly authorizationClient: AuthorizationClient) {}

  async hasNotePermission(
    permission: WorkspaceItemPermission,
    memberId: string
  ): Promise<{ hasPermission: boolean }> {
    try {
      const response =
        await this.authorizationClient.hasWorkspaceItemPermission({
          permission,
          memberId,
        });
      return { hasPermission: response.hasPermission };
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.NotFound) {
          return { hasPermission: false };
        }
        throw new InternalServerErrorException(
          `Failed to check note permission: ${error.message}`
        );
      }
      throw error;
    }
  }

  async hasWriteNotePermission(memberId: string): Promise<boolean> {
    const { hasPermission } = await this.hasNotePermission(
      WorkspaceItemPermission.WRITE,
      memberId
    );
    return hasPermission;
  }

  async hasReadNotePermission(memberId: string): Promise<boolean> {
    const { hasPermission } = await this.hasNotePermission(
      WorkspaceItemPermission.READ,
      memberId
    );
    return hasPermission;
  }

  async getUserNotePermissions(
    memberId: string
  ): Promise<{ canRead: boolean; canWrite: boolean; canDelete: boolean }> {
    try {
      return await this.authorizationClient.getUserWorkspaceItemPermissions({
        memberId,
        workspaceId: '', // TODO: it need workspace
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
