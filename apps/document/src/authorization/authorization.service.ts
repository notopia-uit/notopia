import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { InternalServerErrorException, UnprocessableEntityException } from '@nestjs/common';
import { type ClientGrpc } from '@nestjs/microservices';
import {
  AUTHORIZATION_PACKAGE_NAME,
  AUTHORIZATION_SERVICE_NAME,
  AuthorizationServiceClient,
  WorkspaceItemPermission as WorkspaceItemPermissionPb,
} from '@notopia-uit/pb/authorization';
import { firstValueFrom } from 'rxjs';

import { NoteService } from '#/note';

import { UserNotePermissions as UserNotePermission, WorkspaceItemPermission } from './models';

@Injectable()
export class AuthorizationService implements OnModuleInit {
  private authorizationServiceClient!: AuthorizationServiceClient;

  constructor(
    @Inject(AUTHORIZATION_PACKAGE_NAME) private readonly client: ClientGrpc,
    private readonly noteService: NoteService
  ) {}

  onModuleInit(): void {
    this.authorizationServiceClient = this.client.getService<AuthorizationServiceClient>(
      AUTHORIZATION_SERVICE_NAME
    );
  }

  private toWorkspaceItemPermissionPb(permission: UserNotePermission): WorkspaceItemPermissionPb {
    switch (permission) {
      case 'read':
        return WorkspaceItemPermissionPb.WORKSPACE_ITEM_PERMISSION_READ;
      case 'write':
        return WorkspaceItemPermissionPb.WORKSPACE_ITEM_PERMISSION_WRITE;
      case 'delete':
        return WorkspaceItemPermissionPb.WORKSPACE_ITEM_PERMISSION_DELETE;
      default: {
        const exhaustiveCheck: never = permission;
        throw new UnprocessableEntityException(`Invalid permission: ${String(exhaustiveCheck)}`);
      }
    }
  }

  async hasNotePermission({
    memberId,
    documentId,
    permission,
  }: {
    memberId: string;
    documentId: string;
    permission: UserNotePermission;
  }): Promise<boolean> {
    try {
      const workspace = await this.noteService.getWorkspaceByNote({
        noteId: documentId,
        userId: memberId,
      });
      const response = await firstValueFrom(
        this.authorizationServiceClient.hasWorkspaceItemPermission({
          permission: this.toWorkspaceItemPermissionPb(permission),
          memberId,
          workspaceId: workspace.id,
        })
      );
      return response.hasPermission;
    } catch (error) {
      throw new InternalServerErrorException(
        `Failed to check note permission: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  async getWorkspaceItemPermissions({
    memberId,
    workspaceId,
  }: {
    memberId: string;
    workspaceId: string;
  }): Promise<WorkspaceItemPermission> {
    try {
      const response = await firstValueFrom(
        this.authorizationServiceClient.getUserWorkspaceItemPermissions({
          memberId,
          workspaceId,
        })
      );
      return {
        canRead: response.canRead,
        canWrite: response.canWrite,
        canDelete: response.canDelete,
      };
    } catch (error) {
      throw new InternalServerErrorException(
        `Failed to get workspace item permissions: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  async getUserDocumentPermissions(
    memberId: string,
    documentId: string
  ): Promise<WorkspaceItemPermission> {
    try {
      const workspace = await this.noteService.getWorkspaceByNote({
        noteId: documentId,
        userId: memberId,
      });
      const response = await firstValueFrom(
        this.authorizationServiceClient.getUserWorkspaceItemPermissions({
          memberId,
          workspaceId: workspace.id,
        })
      );
      return {
        canRead: response.canRead,
        canWrite: response.canWrite,
        canDelete: response.canDelete,
      };
    } catch (error) {
      throw new InternalServerErrorException(
        `Failed to get user note permissions: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }
}
