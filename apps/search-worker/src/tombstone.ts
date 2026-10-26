import { Logger, type OnApplicationShutdown } from '@nestjs/common';
import Redis from 'ioredis';
import { LRUCache } from 'lru-cache';

export abstract class TombstoneStore {
  abstract mark(id: string): Promise<void>;
  abstract has(id: string): Promise<boolean>;
}

export class MemoryTombstoneStore extends TombstoneStore {
  private readonly cache: LRUCache<string, true>;

  constructor(ttlSeconds: number, max: number) {
    super();
    new Logger(MemoryTombstoneStore.name).warn(
      'tombstone: no redis url, using in-memory store (single instance only)'
    );
    this.cache = new LRUCache({ max, ttl: ttlSeconds * 1000 });
  }

  async mark(id: string): Promise<void> {
    this.cache.set(id, true);
  }

  async has(id: string): Promise<boolean> {
    return this.cache.has(id);
  }
}

export class RedisTombstoneStore extends TombstoneStore implements OnApplicationShutdown {
  private static readonly prefix = 'search-worker:tombstone:note:';
  private readonly logger = new Logger(RedisTombstoneStore.name);
  private readonly redis: Redis;

  constructor(
    url: string,
    private readonly ttlSeconds: number
  ) {
    super();
    this.redis = new Redis(url);
    this.redis.on('error', (err) => this.logger.error(`tombstone: redis error: ${err.message}`));
  }

  async mark(id: string): Promise<void> {
    await this.redis.set(RedisTombstoneStore.prefix + id, '1', 'EX', this.ttlSeconds);
  }

  async has(id: string): Promise<boolean> {
    return (await this.redis.exists(RedisTombstoneStore.prefix + id)) === 1;
  }

  async onApplicationShutdown() {
    await this.redis.quit();
  }
}
