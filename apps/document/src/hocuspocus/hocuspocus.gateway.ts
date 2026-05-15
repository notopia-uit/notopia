import { Logger, UseGuards } from '@nestjs/common';
import { ConnectedSocket, OnGatewayConnection, WebSocketGateway } from '@nestjs/websockets';
import { Traceable } from 'nestjs-otel';
import { WebSocket } from 'ws';

import { WsUserGuard } from '#/common/user.guard';

import { Hocuspocus } from './hocuspocus';

@WebSocketGateway({ path: '/document/ws/document' })
@Traceable()
export class HocuspocusGateway implements OnGatewayConnection {
  private readonly logger = new Logger(HocuspocusGateway.name);
  constructor(private readonly hocuspocus: Hocuspocus) {}

  @UseGuards(WsUserGuard)
  handleConnection(@ConnectedSocket() socket: WebSocket, req: Request) {
    this.logger.log(`New WebSocket connection: ${socket.url}`);
    this.hocuspocus.hocuspocus.handleConnection(socket, req);
  }
}
