# Document Service (nx: `document`)

Owns document *content* (note metadata lives in `internal/note`). NestJS: HTTP REST +
Yjs WebSocket + Kafka producer/consumer. Spec: `api/document/`.

## Modules (`src/`)

| Module | Note |
| --- | --- |
| `document/` | `DocumentEntity` (Yjs state as `bytea` + `modified` flag), commit workflow, attachment URLs |
| `revision/` | Revision snapshots, paginated read, rename, hard delete |
| `hocuspocus/` | Yjs server, connection auth, live permission revocation |
| `blocknote/` | Schema provider + YDoc↔blocks conversion, tag/link extraction |
| `note/`, `authorization/` | gRPC clients (generated from `proto/`) |
| `authentication/` | Multi-issuer JWKS validation — for the **WebSocket** token |
| `common/` | `UserGuard` + `@User()` — for **HTTP**, reads gateway-forwarded headers |
| `storage/`, `kafka/`, `database/`, `config/` | S3 presign, Kafka client, TypeORM, Zod config |

Two auth paths, easy to trip on: HTTP trusts Traefik's forwarded headers; the WebSocket
gets a raw JWT in the connect payload and validates it itself in
`Hocuspocus.onAuthenticate`.

## The commit pipeline

`DocumentService.commitIfDue()` is the hinge of the whole system. In one transaction
(pessimistic write lock on the document) it: saves a `RevisionEntity` of the current
blocks, clears `modified`, and emits
`events.integration.document.document.committed`. Three callers, one revision when they race
(row lock + min-interval/hash check):

| Trigger | Caller | Gate |
| --- | --- | --- |
| `manual` | `POST /document/documents/{id}/commit` | write permission (viewers 403); always commits changed content |
| `auto` | `AutoCommitExtension.afterStoreDocument` | plan feature `auto_commit` + `auto_commit_min_interval_seconds` |
| `sweeper` | `SweeperCron` | stale `modified` docs idle ≥ `idleSeconds`; walks all of them in batches of 500 per tick |

Manual responses: `201 {revisionId}` revision created, `200 {revisionId}` content equals the
latest revision (returns it), `204` nothing changed and no revision exists. When the content
hash equals the latest revision only `modified` is cleared; `pendingContributorIds` is kept
for the next real commit.

Callers clear the in-memory `modified` flag only when the loaded Y.Doc's content hash equals
the committed one (state vectors miss delete-only edits).

Billing unreachable: auto-commit is skipped (warn), the sweeper and purge skip the batch,
editing continues.

Downstream: `internal/note` (`e_documentcommitted.go`) updates note size/tags/outgoing
links; `search-worker` indexes the content into Meilisearch. **Nothing else re-indexes** —
if a document is never committed, its content is never searchable and its links never
appear in the graph.

## Yjs / Hocuspocus

Extension hooks run by descending priority, not array order:

| Extension | Priority | Role |
| --- | --- | --- |
| `ContributorFlushExtension` | 1100 | Flushes tracked editors to `pendingContributorIds` in `onStoreDocument`, and retries in `afterStoreDocument` (a brand-new document has no row yet during its first store) |
| `@hocuspocus/extension-redis` | 1000 | Store lock; an instance missing it throws `SkipFurtherHooksError`, skipping later hooks, hence the flush runs before it |
| Database (`HocuspocusDatabase`) | 100 | Upserts the row |
| `AutoCommitExtension` | 50 | Must run after the row is stored. No forced commit on disconnect (the sweeper covers idle docs). Resolves the workspace with `NoteService.batchGetNoteWorkspaces` (no auth, marked `HACK:`; should become a service-account role) |

- `HocuspocusDatabase.store` upserts the row and writes `modified` from the Yjs metadata
  (`DocumentService.upsertData`). Stores are debounced server-wide by
  `NOTOPIA_DOCUMENT_HOCUSPOCUS_DEBOUNCE_MS` (default 2000) and
  `NOTOPIA_DOCUMENT_HOCUSPOCUS_MAX_DEBOUNCE_MS` (default 10000). The manual commit calls
  `HocuspocusService.flushStore` first, so edits still in the debounce are stored before
  `commitIfDue` reads the row.
- `AutoCommitExtension` skips the note/billing calls when `lastCommittedAt` is less than
  `NOTOPIA_DOCUMENT_AUTO_COMMIT_CHECK_FLOOR_SECONDS` (default 60, the shortest plan interval;
  `0` disables) ago. Only a cheap local precondition, not a cache: the real interval comes
  from the plan's `auto_commit_min_interval_seconds`.
- `afterLoadDocument` applies the DB `modified` column into the `metadata` Y.Map
  (`YDocMetadataMap`) with origin `{ source: 'local', skipStoreHooks: true }`, so loading
  does not trigger a store. Clients read `{ modified }` from it.
  After a commit, callers use `HocuspocusService.clearModifiedIfCommitted` (or
  `DocumentService.clearModifiedIfCommitted` with the loaded `Y.Doc`), which clears the flag
  only when the loaded doc's content hash equals the committed one, so edits not stored yet
  keep `modified`.
- `note.deleted` → `HocuspocusService.onNoteDeleted` marks the id in `deletedDocumentIds`
  (the store guard), closes its connections, waits for in-flight stores
  (`waitForPendingStores`) and deletes the row; revisions cascade by FK. Without the guard,
  the final store on unload would upsert the row back. The set only holds loaded documents
  and is cleared in `afterUnloadDocument`, so it does not grow.
  Scaling out: the instance consuming `note.deleted` may not hold the live document, so a
  shared set (Redis) does not help. The event must reach every instance (one consumer group
  per instance, or Hocuspocus Redis extension).
- `onRoleChanged` / `onMemberRemoved` consume Kafka events from `authorization` and walk
  live connections to flip `readOnly` or close with codes 4001/4002.

## Crons and locks

`@notopia-uit/lib-server/locks` provides `LocksSetupModule.forRoot({ redisUrl?, ttl? })`.
With no `redisUrl` it uses an in-memory store (single instance only, a warning is logged, intended even in production, see `packages/lib-server/AGENTS.md`); with one it uses `RedisLockStore`. The lock
Redis must run with `noeviction` and persistence, and the app must call
`app.enableShutdownHooks()`. `redis.url` (`NOTOPIA_DOCUMENT_REDIS_URL`) picks the lock store;
`scaling.enabled` (`NOTOPIA_DOCUMENT_SCALING_ENABLED`, requires the URL) only adds
`@hocuspocus/extension-redis`. Its clients come from the app's ioredis via `createClient` (the
extension's own ioredis 5.6.1 parses URLs with the deprecated `url.parse()`). Local compose
Redis for this service is `document_redis`.

Crons (`src/cron/`) are not `@Cron` decorators (those cannot read config): `CronRegistrar`
adds them to `SchedulerRegistry` in `onModuleInit` from `NOTOPIA_DOCUMENT_CRON_*` (name = lock
key, schedule, enabled, timeout seconds), with no `timeZone`, so they run in the process TZ.
Each tick runs in `Locks.withLock(name, …)`; `LockNotAcquiredError` means another instance has
it. Jobs get an `AbortSignal` (lock lost or timeout) and stop between batches when it aborts.
Defaults: `document:sweeper` `*/10 * * * *` (timeout 300 s, idle 600 s),
`document:revision-purge` `30 3 * * *` (timeout 1800 s); each has `_ENABLED=false` to disable.
Revision purge deletes revisions older than the plan's `revision_history_days` and always keeps
the latest one.

## Revisions are hard-deleted

Revisions have no `deletedAt` (no soft delete); the API delete and the revision purge cron
remove rows. There are no migration files (TypeORM `synchronize` outside production), so an
environment with existing data must run this **before** deploying, otherwise dropping the
column brings soft-deleted rows back:

```sql
DELETE FROM revisions WHERE "deletedAt" IS NOT NULL;
```

## Conventions

- HTTP handlers extend the generated abstract class from
  `@notopia-uit/api-document-nestjs-server` — change `api/document/` and run `nx gen api`
  first, then implement the new method.
- Kafka payload types come from `@notopia-uit/api-share-gen` (defined in `api/share/`),
  never hand-written interfaces.
- `@Traceable()` on services; structured logger calls are `this.logger.log({...}, 'msg')`.

## Database schema and migrations

The app never syncs or migrates the schema on startup (`synchronize: false`). All schema
changes go through the TypeORM CLI against `database/datasource.ts`:

| Target (`nx run document:…`) | Use |
| --- | --- |
| `db:sync` | Dev only: `schema:sync`, forces the DB to match the entities. Destructive, never on prod |
| `db:generate -- database/migrations/Name` | Diff entities against the DB and write a migration. Review it before committing |
| `db:create -- database/migrations/Name` | Empty migration for hand-written changes |
| `db:migrate` / `db:revert` / `db:status` | Run, undo the last, or list migrations |

`just fixtures-document` runs `db:sync` first, since fixtures need the tables to exist.
