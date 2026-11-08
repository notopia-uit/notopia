# Document Service (nx: `document`)

Owns document *content* (metadata in `internal/note`). NestJS: HTTP REST +
Yjs WebSocket + Kafka producer/consumer. Spec: `api/document/`.

## Modules (`src/`)

| Module | Note |
| --- | --- |
| `document/` | `DocumentEntity` (Yjs state as `bytea` + `modified` flag), commit workflow, attachment URLs |
| `revision/` | Revision snapshots, paginated read, rename, hard delete |
| `hocuspocus/` | Yjs server, connection auth, live permission revocation |
| `blocknote/` | Schema provider + YDoc↔blocks conversion, tag/link extraction |
| `note/`, `authorization/` | gRPC clients (generated from `proto/`) |
| `authentication/` | Multi-issuer JWKS validation — for **WebSocket** token |
| `common/` | `UserGuard` + `@User()` — for **HTTP**, reads gateway-forwarded headers |
| `storage/`, `kafka/`, `database/`, `config/` | S3 presign, Kafka client, TypeORM, Zod config |

Two auth paths: HTTP trusts Traefik forwarded headers; WebSocket validates raw JWT
from connect payload itself in `Hocuspocus.onAuthenticate`.

## The commit pipeline

`DocumentService.commitIfDue()` hinges system. One transaction (pessimistic write lock
on document): saves `RevisionEntity` of current blocks, clears `modified`, emits
`events.integration.document.document.committed`. Three callers, one revision on race
(row lock + min-interval/hash check):

| Trigger | Caller | Gate |
| --- | --- | --- |
| `manual` | `POST /document/documents/{id}/commit` | write permission (viewers 403); always commits changed content |
| `auto` | `AutoCommitExtension.afterStoreDocument` | plan feature `auto_commit` + `auto_commit_min_interval_seconds` |
| `sweeper` | `SweeperCron` | stale `modified` docs idle ≥ `idleSeconds`; walks all in batches of 500 per tick |

Manual responses: `201 {revisionId}` revision created, `200 {revisionId}` content equals
latest revision (returns it), `204` nothing changed, no revision exists. If content hash
equals latest revision, only `modified` cleared; `pendingContributorIds` kept for next real commit.

Callers clear in-memory `modified` only when loaded Y.Doc content hash equals committed
hash (state vectors miss delete-only edits).

Billing unreachable: auto-commit skipped (warn), sweeper and purge skip batch,
editing continues.

Downstream: `internal/note` (`e_documentcommitted.go`) updates note size/tags/outgoing
links; `search-worker` indexes content into Meilisearch. **Nothing else re-indexes** —
uncommitted document content never searchable, links never appear in graph.

## Yjs / Hocuspocus

Extension hooks run descending priority, not array order:

| Extension | Priority | Role |
| --- | --- | --- |
| `ContributorFlushExtension` | 1100 | Flushes tracked editors to `pendingContributorIds` in `onStoreDocument`, retries in `afterStoreDocument` (new document has no row during first store) |
| `@hocuspocus/extension-redis` | 1000 | Store lock; instance missing it throws `SkipFurtherHooksError`, skips later hooks, so flush runs first |
| Database (`HocuspocusDatabase`) | 100 | Upserts row |
| `AutoCommitExtension` | 50 | Runs after row stored. No forced commit on disconnect (sweeper covers idle docs). Resolves workspace with `NoteService.batchGetNoteWorkspaces` (no auth, marked `HACK:`; should become service-account role) |

- `HocuspocusDatabase.store` upserts row, writes `modified` from Yjs metadata
  (`DocumentService.upsertData`). Stores debounced server-wide by
  `NOTOPIA_DOCUMENT_HOCUSPOCUS_DEBOUNCE_MS` (default 2000) and
  `NOTOPIA_DOCUMENT_HOCUSPOCUS_MAX_DEBOUNCE_MS` (default 10000). Manual commit calls
  `HocuspocusService.flushStore` first, storing debounced edits before
  `commitIfDue` reads row.
- `AutoCommitExtension` skips note/billing calls when `lastCommittedAt` is less than
  `NOTOPIA_DOCUMENT_AUTO_COMMIT_CHECK_FLOOR_SECONDS` (default 60, shortest plan interval;
  `0` disables) ago. Cheap local precondition, not cache: real interval from plan's
  `auto_commit_min_interval_seconds`.
- `afterLoadDocument` applies DB `modified` column to `metadata` Y.Map
  (`YDocMetadataMap`) with origin `{ source: 'local', skipStoreHooks: true }`, preventing
  store on load. Clients read `{ modified }` from it.
  After commit, callers use `HocuspocusService.clearModifiedIfCommitted` (or
  `DocumentService.clearModifiedIfCommitted` with loaded `Y.Doc`), clearing flag only
  when loaded doc content hash equals committed hash, so unstored edits keep `modified`.
- `note.deleted` → `HocuspocusService.onNoteDeleted` marks id in `deletedDocumentIds`
  (store guard), closes connections, waits for in-flight stores
  (`waitForPendingStores`), deletes row; revisions cascade by FK. Without guard,
  final store on unload would upsert row back. Set holds only loaded documents,
  cleared in `afterUnloadDocument`, so does not grow.
  Scaling out: instance consuming `note.deleted` may not hold live document, so shared
  set (Redis) does not help. Event must reach every instance (one consumer group
  per instance, or Hocuspocus Redis extension).
- `onRoleChanged` / `onMemberRemoved` consume Kafka events from `authorization`, walk
  live connections to flip `readOnly` or close with codes 4001/4002.

## Crons and locks

`@notopia-uit/lib-server/locks` provides `LocksSetupModule.forRoot({ redisUrl?, ttl? })`.
Without `redisUrl`, uses in-memory store (single instance only, logs warning, intended even in production, see `packages/lib-server/AGENTS.md`); with URL uses `RedisLockStore`. Lock
Redis must use `noeviction` and persistence; app must call
`app.enableShutdownHooks()`. `redis.url` (`NOTOPIA_DOCUMENT_REDIS_URL`) picks lock store;
`scaling.enabled` (`NOTOPIA_DOCUMENT_SCALING_ENABLED`, requires URL) only adds
`@hocuspocus/extension-redis`. Its clients come from app's ioredis via `createClient` (extension's
own ioredis 5.6.1 parses URLs with deprecated `url.parse()`). Local compose Redis:
`document_redis`.

Crons (`src/cron/`) not `@Cron` decorators (cannot read config): `CronRegistrar`
adds to `SchedulerRegistry` in `onModuleInit` from `NOTOPIA_DOCUMENT_CRON_*` (name = lock
key, schedule, enabled, timeout seconds), no `timeZone`, so run in process TZ.
Each tick runs in `Locks.withLock(name, …)`; `LockNotAcquiredError` means another instance
holds lock. Jobs get `AbortSignal` (lock lost or timeout), stop between batches on abort.
Defaults: `document:sweeper` `*/10 * * * *` (timeout 300 s, idle 600 s),
`document:revision-purge` `30 3 * * *` (timeout 1800 s); each has `_ENABLED=false` to disable.
Revision purge deletes revisions older than plan's `revision_history_days`, always keeps
latest.

## Revisions are hard-deleted

Revisions have no `deletedAt` (no soft delete); API delete and revision purge cron
remove rows. For existing schema with `deletedAt`, remove soft-deleted rows
before dropping column, else rows become visible again:

```sql
DELETE FROM revisions WHERE "deletedAt" IS NOT NULL;
```

## Conventions

- HTTP handlers extend generated abstract class from
  `@notopia-uit/api-document-nestjs-server` — change `api/document/`, run `nx gen api`
  first, then implement new method.
- Kafka payload types from `@notopia-uit/api-share-gen` (defined in `api/share/`),
  never hand-written interfaces.
- `@Traceable()` on services; structured logger calls are `this.logger.log({...}, 'msg')`.

## Database schema and migrations

No sync or migrate on startup (`synchronize: false`). Schema changes via TypeORM
CLI: `nx run document:db:{sync,generate,create,migrate,revert,status}` (see `package.json`).
`db:sync` dev only, destructive; production runs `dist/migrate.cjs` before starting
service. Initial migration targets empty database; existing schemas need carefully
verified baseline before recording it as applied. `just fixtures-document` runs `db:sync` first.
Health endpoints: `/document/health/live`, `/document/health/ready`, and
`/document/health/startup` on main port.