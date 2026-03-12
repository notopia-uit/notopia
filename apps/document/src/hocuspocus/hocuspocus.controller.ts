import { Hocuspocus } from '@hocuspocus/server';
import {
  ConnectedSocket,
  OnGatewayConnection,
  WebSocketGateway,
} from '@nestjs/websockets';
import { IncomingMessage } from 'http';
import { Traceable } from 'nestjs-otel';
import { WebSocket } from 'ws';

@WebSocketGateway({ path: '/document/ws/document' })
@Traceable()
export class HocuspocusGateway implements OnGatewayConnection {
  constructor(private readonly hocuspocus: Hocuspocus) {}

  handleConnection(
    @ConnectedSocket() socket: WebSocket,
    request: IncomingMessage
  ) {
    this.hocuspocus.handleConnection(socket, request);
  }
}
