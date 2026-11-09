import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import type { ClientGrpc } from '@nestjs/microservices';
import {
  AUTHORIZATION_PACKAGE_NAME,
  AUTHORIZATION_SERVICE_NAME,
  type AuthorizationServiceClient,
  WorkspaceItemPermission,
} from '@notopia-uit/pb/authorization';
import { firstValueFrom, timeout } from 'rxjs';

export type WorkspaceSummary = {
  workspaceId: string;
  role: string;
};

@Injectable()
export class AuthorizationService implements OnModuleInit {
  private readonly logger = new Logger(AuthorizationService.name);
  private client!: AuthorizationServiceClient;

  constructor(@Inject(AUTHORIZATION_PACKAGE_NAME) private readonly grpc: ClientGrpc) {}

  onModuleInit(): void {
    this.client = this.grpc.getService<AuthorizationServiceClient>(AUTHORIZATION_SERVICE_NAME);
  }

  async permissions(userId: string, workspaceId: string) {
    return firstValueFrom(
      this.client
        .getUserWorkspaceItemPermissions({ memberId: userId, workspaceId })
        .pipe(timeout(5000))
    );
  }

  async canRead(userId: string, workspaceId: string) {
    const result = await firstValueFrom(
      this.client
        .hasWorkspaceItemPermission({
          memberId: userId,
          workspaceId,
          permission: WorkspaceItemPermission.WORKSPACE_ITEM_PERMISSION_READ,
        })
        .pipe(timeout(5000))
    );
    return result.hasPermission;
  }

  async getUserWorkspaces({ userId }: { userId: string }): Promise<WorkspaceSummary[]> {
    this.logger.debug('Fetching user workspaces for tool call');
    const response = await firstValueFrom(
      this.client.getUserWorkspaces({ userId }).pipe(timeout(5000))
    );
    return response.workspaces.map((workspace) => ({
      workspaceId: workspace.id,
      role: String(workspace.role),
    }));
  }
}
