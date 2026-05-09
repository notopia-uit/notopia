import { Hocuspocus } from '@hocuspocus/server';
import { UseGuards } from '@nestjs/common';
import { ConnectedSocket, OnGatewayConnection, WebSocketGateway } from '@nestjs/websockets';
import { Traceable } from 'nestjs-otel';
import { WebSocket } from 'ws';

import { type User } from '#/common/user';
import { ReqUser } from '#/common/user.decorator';
import { WsUserGuard } from '#/common/user.guard';

import { HocuspocusContext } from './hocuspocus-context';

@WebSocketGateway({ path: '/document/ws/document' })
@Traceable()
export class HocuspocusGateway implements OnGatewayConnection {
  constructor(private readonly hocuspocus: Hocuspocus) {}

  @UseGuards(WsUserGuard)
  handleConnection(@ConnectedSocket() socket: WebSocket, @ReqUser() user: User, req: Request) {
    const context: HocuspocusContext = { user };
    this.hocuspocus.handleConnection(socket, req, context);
  }
}
