import { Database } from '@hocuspocus/extension-database';
import { Redis, type RedisInstance } from '@hocuspocus/extension-redis';
import { type Extension, Hocuspocus } from '@hocuspocus/server';

import { BillingService } from '../billing';
import { DocumentService } from '../document/document.service';
import { NoteService } from '../note/note.service';
import { AUTO_COMMIT_PRIORITY, AutoCommitExtension } from './auto-commit.extension';
import {
  CONTRIBUTOR_FLUSH_PRIORITY,
  ContributorFlushExtension,
} from './contributor-flush.extension';
import { ContributorTracker } from './contributor-tracker';

const HOCUSPOCUS_DEFAULT_PRIORITY = 100;

describe('Hocuspocus extension priority', () => {
  it('runs contributor flush, then the Redis lock, then the database store, then auto-commit', () => {
    const redis = new Redis({
      createClient: () => ({ on: vi.fn<() => void>() }) as unknown as RedisInstance,
    });
    const flush = new ContributorFlushExtension(
      new ContributorTracker(),
      {} as unknown as DocumentService
    );
    const database = new Database({});
    const autoCommit = new AutoCommitExtension(
      {} as unknown as DocumentService,
      {} as unknown as NoteService,
      {} as unknown as BillingService,
      60
    );

    const databasePriority = (database as Extension).priority ?? HOCUSPOCUS_DEFAULT_PRIORITY;

    expect(CONTRIBUTOR_FLUSH_PRIORITY).toBeGreaterThan(redis.priority);
    expect(redis.priority).toBeGreaterThan(databasePriority);
    expect(databasePriority).toBeGreaterThan(AUTO_COMMIT_PRIORITY);

    const server = new Hocuspocus({ extensions: [autoCommit, database, redis, flush] });
    expect(server.configuration.extensions.slice(0, 4)).toEqual([
      flush,
      redis,
      database,
      autoCommit,
    ]);
  });
});
