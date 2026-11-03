import { EventEmitter } from 'node:events';

import { Module } from '@nestjs/common';
import { TerminusModule } from '@nestjs/terminus';
import { Test } from '@nestjs/testing';
import { Meilisearch } from 'meilisearch';
import { afterEach, expect, it, vi } from 'vitest';

import { HealthController, KafkaConsumerState } from './health';
import { TrackedKafkaServer } from './kafka.server';

@Module({
  imports: [TerminusModule],
  controllers: [HealthController],
  providers: [
    KafkaConsumerState,
    { provide: Meilisearch, useValue: { health: () => Promise.resolve() } },
  ],
})
class HealthTestModule {}

afterEach(() => vi.restoreAllMocks());

it('serves liveness while Kafka retries, then tracks consumer failure and recovery over HTTP', async () => {
  const module = await Test.createTestingModule({ imports: [HealthTestModule] }).compile();
  const app = module.createNestApplication();
  const state = app.get(KafkaConsumerState);
  const events = new EventEmitter();
  const names = Object.fromEntries(
    [
      'CONNECT',
      'DISCONNECT',
      'STOP',
      'CRASH',
      'GROUP_JOIN',
      'HEARTBEAT',
      'REBALANCING',
      'REQUEST_TIMEOUT',
    ].map((name) => [name, name])
  );
  const consumer = {
    events: names,
    on: (event: string, callback: () => void) => events.on(event, callback),
    connect: vi
      .fn()
      .mockRejectedValueOnce(new Error('broker unavailable'))
      .mockResolvedValue(undefined),
    disconnect: vi.fn().mockResolvedValue(undefined),
    subscribe: vi.fn().mockResolvedValue(undefined),
    run: vi.fn().mockImplementation(() => {
      events.emit('GROUP_JOIN');
      return Promise.resolve();
    }),
  };
  const producer = {
    events: names,
    on: () => {},
    connect: vi.fn().mockResolvedValue(undefined),
    disconnect: vi.fn().mockResolvedValue(undefined),
  };
  const server = new TrackedKafkaServer(
    { client: { brokers: ['localhost:9092'] }, consumer: { groupId: 'test' } },
    state
  );
  vi.spyOn(server, 'createClient').mockReturnValue({
    consumer: () => consumer,
    producer: () => producer,
  });
  app.connectMicroservice({ strategy: server });
  try {
    await app.listen(0, '127.0.0.1');
    const url = await app.getUrl();
    const starting = app.startAllMicroservices();
    const status = async (path: string) =>
      (await fetch(`${url}/search-worker/health/${path}`)).status;
    expect(await status('live')).toBe(200);
    expect(await status('ready')).toBe(503);
    await starting;
    await vi.waitFor(async () => expect(await status('ready')).toBe(200));
    const clock = vi.spyOn(Date, 'now').mockReturnValue(Date.now() + 16000);
    expect(await status('ready')).toBe(503);
    events.emit('HEARTBEAT');
    expect(await status('ready')).toBe(200);
    clock.mockRestore();
    events.emit('REQUEST_TIMEOUT');
    await vi.waitFor(async () => expect(await status('ready')).toBe(503));
    events.emit('HEARTBEAT');
    await vi.waitFor(async () => expect(await status('ready')).toBe(200));
    events.emit('CRASH');
    await vi.waitFor(async () => expect(await status('ready')).toBe(503));
    expect(await status('live')).toBe(200);
  } finally {
    await app.close();
  }
}, 15000);
