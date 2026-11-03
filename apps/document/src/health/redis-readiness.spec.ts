import { once } from 'node:events';
import { createServer, type Socket } from 'node:net';

import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { Test } from '@nestjs/testing';
import { DataSource } from 'typeorm';
import { expect, it } from 'vitest';

import { REDIS_CONFIG, SCALING_CONFIG } from '../config';
import { DatabaseModule } from '../database';
import { HealthModule } from './health.module';

@Module({
  providers: [{ provide: DataSource, useValue: { query: () => Promise.resolve([]) } }],
  exports: [DataSource],
})
class FakeDatabaseModule {}

it('returns unavailable and closes the socket when connected Redis stops answering', async () => {
  const sockets = new Set<Socket>();
  let pings = 0;
  let socketClosed: Promise<unknown> = Promise.resolve();
  const server = createServer((socket) => {
    sockets.add(socket);
    socketClosed = once(socket, 'close');
    socket.on('close', () => sockets.delete(socket));
    let buffer = '';
    socket.on('data', (data) => {
      buffer += data.toString();
      while (buffer.startsWith('*')) {
        const parts = buffer.split('\r\n');
        const count = Number(parts[0].slice(1));
        if (parts.length < count * 2 + 2) break;
        const command = parts[2].toLowerCase();
        buffer = parts.slice(count * 2 + 1).join('\r\n');
        if (command === 'ping') {
          pings++;
          continue;
        }
        if (command === 'info') {
          const info = 'redis_version:8.0.0\r\n';
          socket.write(`$${Buffer.byteLength(info)}\r\n${info}\r\n`);
        } else socket.write('+OK\r\n');
      }
    });
  });
  server.listen(0, '127.0.0.1');
  await once(server, 'listening');
  const address = server.address();
  if (!address || typeof address === 'string') throw new Error('No TCP address');
  const module = await Test.createTestingModule({
    imports: [ConfigModule.forRoot({ isGlobal: true }), HealthModule],
  })
    .overrideModule(DatabaseModule)
    .useModule(FakeDatabaseModule)
    .overrideProvider(ConfigService)
    .useValue({
      getOrThrow: (key: symbol) =>
        key === REDIS_CONFIG
          ? { url: `redis://127.0.0.1:${address.port}` }
          : key === SCALING_CONFIG
            ? { enabled: true }
            : undefined,
    })
    .compile();
  const app = module.createNestApplication();
  await app.listen(0, '127.0.0.1');
  try {
    const result = await Promise.race([
      fetch(`${await app.getUrl()}/document/health/ready`).then((response) => response.status),
      new Promise<string>((resolve) => setTimeout(() => resolve('timed out'), 3500)),
    ]);
    expect(result).toBe(503);
    expect(pings).toBe(1);
    await Promise.race([
      socketClosed,
      new Promise<never>((_, reject) =>
        setTimeout(() => reject(new Error('Redis socket was not closed')), 1000)
      ),
    ]);
    expect(sockets.size).toBe(0);
  } finally {
    for (const socket of sockets) socket.destroy();
    await app.close();
    server.close();
    await once(server, 'close');
  }
}, 10000);
