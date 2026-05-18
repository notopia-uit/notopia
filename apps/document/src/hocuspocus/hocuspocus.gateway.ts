import { Server as HttpServer, IncomingMessage } from 'node:http';
import { Duplex } from 'node:stream';

import { Injectable, Logger, OnModuleDestroy, OnModuleInit } from '@nestjs/common';
import { HttpAdapterHost } from '@nestjs/core';
import { Traceable } from 'nestjs-otel';
import { WebSocketServer } from 'ws';

import { Hocuspocus } from './hocuspocus';

@Traceable()
@Injectable()
export class HocuspocusGateway implements OnModuleInit, OnModuleDestroy {
  private server: WebSocketServer | undefined;

  private httpServer: HttpServer | undefined;
  private readonly logger = new Logger(HocuspocusGateway.name);

  constructor(
    private readonly hocuspocus: Hocuspocus,
    private readonly adapterHost: HttpAdapterHost
  ) {}

  onModuleInit() {
    this.httpServer = this.adapterHost.httpAdapter.getHttpServer() as HttpServer;
    this.server = new WebSocketServer({ noServer: true });

    this.httpServer.on('upgrade', (request, socket, head) => this.onUpgrade(request, socket, head));

    this.server?.on('connection', (ws, request) => {
      this.logger.debug(
        { remoteAddress: request.socket.remoteAddress },
        'WebSocket connection established'
      );
      const protocol = (request.headers['x-forwarded-proto'] || 'http') as string;
      const webRequest = new Request(`${protocol}://${request.headers.host}${request.url}`, {
        headers: new Headers(request.headers as any),
        method: request.method,
      });

      const clientConnection = this.hocuspocus.hocuspocus.handleConnection(ws, webRequest);

      ws.on('message', (data: Buffer) => {
        clientConnection.handleMessage(new Uint8Array(data));
      });

      ws.on('close', (code: number, reason: Buffer) => {
        clientConnection.handleClose({ code, reason: reason.toString() });
      });
    });
  }

  private onUpgrade(request: IncomingMessage, socket: Duplex, head: Buffer) {
    const url = new URL(request.url || '', `http://${request.headers.host}`);
    if (url.pathname === '/document/ws/document') {
      this.logger.debug(`Upgrading connection to WebSocket for ${url.pathname}`);
      this.server?.handleUpgrade(request, socket, head, (ws) => {
        this.server?.emit('connection', ws, request);
      });
    }
  }

  onModuleDestroy() {
    this.hocuspocus.hocuspocus.closeConnections();
    this.server?.close();
    this.httpServer?.off('upgrade', this.onUpgrade.bind(this));
  }
}
