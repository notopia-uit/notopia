import { Hocuspocus } from '@hocuspocus/server';
import { UseGuards } from '@nestjs/common';
import { ConnectedSocket, OnGatewayConnection, WebSocketGateway } from '@nestjs/websockets';
import { Traceable } from 'nestjs-otel';
import { WebSocket } from 'ws';

import { WsUserGuard } from '#/common/user.guard';

@WebSocketGateway({ path: '/document/ws/document' })
@Traceable()
export class HocuspocusGateway implements OnGatewayConnection {
  constructor(private readonly hocuspocus: Hocuspocus) {}

  @UseGuards(WsUserGuard)
  handleConnection(@ConnectedSocket() socket: WebSocket, req: Request) {
    this.hocuspocus.handleConnection(socket, req);
  }
}
