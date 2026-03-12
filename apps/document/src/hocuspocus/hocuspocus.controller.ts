import {
  ConnectedSocket,
  MessageBody,
  SubscribeMessage,
  WebSocketGateway,
} from '@nestjs/websockets';
import { Socket } from 'socket.io';

@WebSocketGateway({ namespace: 'documents' })
export class DocumentGateway {
  @SubscribeMessage('watchDocument')
  handleEvent(@MessageBody() data: any, @ConnectedSocket() client: Socket) {
    // Your WebSocket logic here
    return { event: 'watching', data: data.documentId };
  }
}
