import { lockStoreContract } from '@nestjs/locks/testing';
import Redis from 'ioredis';

import { RedisLockStore } from './redis-lock.store';

const url = process.env.LOCKS_TEST_REDIS_URL;

describe.skipIf(!url)('RedisLockStore contract', () => {
  const pool: RedisLockStore[] = [];
  const clients: Redis[] = [];
  let next = 0;
  const store = () => {
    if (!pool.length) {
      for (let i = 0; i < 8; i++) {
        const client = new Redis(url!);
        clients.push(client);
        pool.push(new RedisLockStore(client));
      }
    }
    return pool[next++ % pool.length]!;
  };
  afterAll(async () => {
    await Promise.all(clients.map((c) => c.quit()));
  });
  for (const c of lockStoreContract(store, { concurrent: true })) {
    it(c.name, c.run, 60_000);
  }
});
