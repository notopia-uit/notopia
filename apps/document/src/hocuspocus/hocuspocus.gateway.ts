import { Server as HttpServer } from 'node:http'; // or 'node:https' if you use SSL

import { Injectable, Logger, OnModuleDestroy, OnModuleInit } from '@nestjs/common';
import { HttpAdapterHost } from '@nestjs/core';
import { Traceable } from 'nestjs-otel';
import { WebSocketServer } from 'ws';

import { Hocuspocus } from './hocuspocus';

// You know, it is not really a nestjs gateway at all
@Traceable()
@Injectable()
export class HocuspocusGateway implements OnModuleInit, OnModuleDestroy {
  private server: WebSocketServer | undefined;
  private readonly logger = new Logger(HocuspocusGateway.name);

  constructor(
    private readonly hocuspocus: Hocuspocus,
    private readonly adapterHost: HttpAdapterHost
  ) {}

  onModuleInit() {
    const server = this.adapterHost.httpAdapter.getHttpServer() as HttpServer;
    this.server = new WebSocketServer({ noServer: true });

    server.on('upgrade', (request, socket, head) => {
      const url = new URL(request.url || '', `http://${request.headers.host}`);
      if (url.pathname === '/document/ws/document') {
        this.logger.debug(`Upgrading connection to WebSocket for ${url.pathname}`);
        this.server?.handleUpgrade(request, socket, head, (ws) => {
          this.server?.emit('connection', ws, request);
        });
      }
    });

    this.server?.on('connection', (ws, request) => {
      this.logger.debug(`WebSocket connection established from ${request.socket.remoteAddress}`);
      const protocol = request.headers['x-forwarded-proto'] || 'http';
      const webRequest = new Request(`${protocol}://${request.headers.host}${request.url}`, {
        headers: new Headers(request.headers as any),
        method: request.method,
      });

      this.hocuspocus.hocuspocus.handleConnection(ws, webRequest);
    });
  }

  onModuleDestroy() {
    this.hocuspocus.hocuspocus.closeConnections();
    this.server?.close();
  }
}
