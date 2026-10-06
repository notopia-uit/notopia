import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import type { ClientGrpc } from '@nestjs/microservices';
import {
  AUTHORIZATION_PACKAGE_NAME,
  AUTHORIZATION_SERVICE_NAME,
  type AuthorizationServiceClient,
} from '@notopia-uit/pb/authorization';
import { firstValueFrom } from 'rxjs';

export type WorkspaceSummary = {
  workspaceId: string;
  role: string;
};

@Injectable()
export class AuthorizationClientService implements OnModuleInit {
  private readonly logger = new Logger(AuthorizationClientService.name);
  private client!: AuthorizationServiceClient;

  constructor(@Inject(AUTHORIZATION_PACKAGE_NAME) private readonly grpc: ClientGrpc) {}

  onModuleInit(): void {
    this.client = this.grpc.getService<AuthorizationServiceClient>(AUTHORIZATION_SERVICE_NAME);
  }

  async getUserWorkspaces({ userId }: { userId: string }): Promise<WorkspaceSummary[]> {
    this.logger.debug('Fetching user workspaces for tool call');
    const response = await firstValueFrom(this.client.getUserWorkspaces({ userId }));
    return response.workspaces.map((workspace) => ({
      workspaceId: workspace.id,
      role: String(workspace.role),
    }));
  }
}
