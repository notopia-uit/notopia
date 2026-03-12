import { Hocuspocus } from '@hocuspocus/server';
import {
  ConnectedSocket,
  OnGatewayConnection,
  WebSocketGateway,
} from '@nestjs/websockets';
import { IncomingMessage } from 'http';
import { Traceable } from 'nestjs-otel';
import { WebSocket } from 'ws';

import type { User } from '../common/user';
import { ReqUser } from '../common/user.decorator';

@WebSocketGateway({ path: '/document/ws/document' })
@Traceable()
export class HocuspocusGateway implements OnGatewayConnection {
  constructor(private readonly hocuspocus: Hocuspocus) {}

  handleConnection(
    @ConnectedSocket() socket: WebSocket,
    @ReqUser() user: User,
    req: IncomingMessage
  ) {
    this.hocuspocus.handleConnection(socket, req, user);
  }
}
