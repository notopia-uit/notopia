import { randomUUID } from 'node:crypto';

import { Redis } from 'ioredis';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import { usageConfigSchema } from '../src/config';
import { RELEASE, RESERVE, START, THROTTLE } from '../src/usage/usage.scripts';
import { usagePeriod } from '../src/usage/usage.service';

it('uses UTC calendar months and keeps accounting through the reset grace period', () => {
  expect(usagePeriod(new Date('2026-12-31T23:59:59Z'))).toEqual({
    month: '2026-12',
    resetAt: '2027-01-01T00:00:00.000Z',
    expiresAt: Date.parse('2027-01-08T00:00:00Z') / 1000,
  });
});
it('requires shared accounting to be configured', () => {
  expect(() => usageConfigSchema.parse({})).toThrow();
  expect(usageConfigSchema.parse({ redisUrl: 'redis://127.0.0.1:6376' }).redisUrl).toBe(
    'redis://127.0.0.1:6376'
  );
});

describe.skipIf(!process.env.AI_TEST_REDIS_URL)('Redis atomic AI accounting', () => {
  let redis: Redis;
  const prefix = `ai-test:${randomUUID()}`;
  const keys = (name: string, request = 'request') => [
    `${prefix}:${name}:usage`,
    `${prefix}:${name}:${request}`,
    `${prefix}:${name}:streams`,
    `${prefix}:${name}:rate`,
  ];
  const reserve = (
    name: string,
    request = 'request',
    quota = 2,
    concurrent = 100,
    rate = 100,
    now = Date.now()
  ) =>
    redis.eval(
      RESERVE,
      4,
      ...keys(name, request),
      now,
      quota,
      Math.floor(Date.now() / 1000) + 60,
      concurrent,
      rate,
      1000,
      request
    );
  beforeAll(async () => {
    redis = new Redis(process.env.AI_TEST_REDIS_URL!, {
      lazyConnect: true,
      maxRetriesPerRequest: 0,
    });
    await redis.connect();
  });
  afterAll(async () => {
    if (!redis) return;
    const owned = await redis.keys(`${prefix}:*`);
    if (owned.length) await redis.del(...owned);
    await redis.quit();
  });
  it('never overspends the quota under concurrent requests', async () => {
    const results = await Promise.all(
      Array.from({ length: 30 }, (_, i) => reserve('race', `req${i}`, 5))
    );
    expect(results.filter((result) => result === 'ok')).toHaveLength(5);
    expect(results.filter((result) => result === 'quota')).toHaveLength(25);
    expect(await redis.get(keys('race')[0])).toBe('5');
  });
  it('deduplicates requests without spending another credit', async () => {
    expect(await reserve('duplicate')).toBe('ok');
    expect(await reserve('duplicate')).toBe('duplicate');
    expect(await redis.get(keys('duplicate')[0])).toBe('1');
  });
  it('refunds only reservations which have not started, once', async () => {
    await reserve('refund');
    await redis.eval(RELEASE, 3, ...keys('refund').slice(0, 3), 'request', '1');
    await redis.eval(RELEASE, 3, ...keys('refund').slice(0, 3), 'request', '1');
    expect(await redis.get(keys('refund')[0])).toBe('0');
    expect(await reserve('refund')).toBe('ok');
    expect(await redis.eval(START, 1, keys('refund')[1], 'request')).toBe(1);
    expect(await redis.eval(START, 1, keys('refund')[1], 'request')).toBe(0);
    await redis.eval(RELEASE, 3, ...keys('refund').slice(0, 3), 'request', '1');
    expect(await redis.get(keys('refund')[0])).toBe('1');
    expect(await reserve('refund')).toBe('duplicate');
  });
  it('limits active streams and recovers abandoned leases', async () => {
    const now = Date.now();
    expect(await reserve('streams', 'one', 20, 1, 100, now)).toBe('ok');
    expect(await reserve('streams', 'two', 20, 1, 100, now)).toBe('concurrency');
    expect(await reserve('streams', 'two', 20, 1, 100, now + 1001)).toBe('ok');
  });
  it('limits request bursts without debiting rejected requests', async () => {
    expect(await reserve('rate', 'one', 20, 100, 1)).toBe('ok');
    expect(await reserve('rate', 'two', 20, 100, 1)).toBe('rate');
    expect(await redis.get(keys('rate')[0])).toBe('1');
  });
  it('throttles repeated attempts independently of billing and reservations', async () => {
    const key = `${prefix}:attempts`;
    expect(await redis.eval(THROTTLE, 1, key, 2)).toBe(1);
    expect(await redis.eval(THROTTLE, 1, key, 2)).toBe(1);
    expect(await redis.eval(THROTTLE, 1, key, 2)).toBe(0);
    expect(await redis.ttl(key)).toBeGreaterThan(0);
  });
  it('tracks unlimited plans without imposing a credit ceiling', async () => {
    const results = await Promise.all(
      Array.from({ length: 10 }, (_, i) => reserve('unlimited', `req${i}`, -1))
    );
    expect(results.every((result) => result === 'ok')).toBe(true);
  });
});
