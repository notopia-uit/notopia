# Billing 02 — Shared Foundations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Cross-cutting pieces for billing and enforcement: gateway groups + admin guard, identity config, service accounts, Go cron with Redis locks, NestJS locks, 402 contract, note batch lookup RPC, billing permissions.

**Architecture:** Small additions to `pkg/common`, a new `pkg/cron`, a locks module in `packages/lib-server`, contract files in `api/common` and `proto/`, and one read-model query in note. No billing logic.

**Tech Stack:** Go (gin, viper, gocron v2.22.0, gocron-redis-lock v2.2.1, go-redis v9), NestJS 11 (`@nestjs/schedule` 12.0.2, `@nestjs/locks` 0.0.1, `ioredis` 5.6.1), buf, oapi-codegen, sqlc, Authentik blueprints.

**Spec:** `docs/superpowers/specs/2026-10-11-billing-02-foundations-design.md`

## Global Constraints

- Traefik `agilezebra/jwt-middleware` v1.6.0 forwards array claims as JSON (`json.Marshal`), e.g. `X-Forwarded-Groups: ["notopia-admins"]`.
- Admin group default `notopia-admins`; env `NOTOPIA_{SERVICE}_IDENTITY_ADMIN_GROUP`, `NOTOPIA_{SERVICE}_IDENTITY_SERVICE_ACCOUNT_USER_ID`.
- Service accounts: `note-service` 1001 (exists), `document-service` 1002, `billing-service` 1003. User ids are Authentik pks as strings.
- Cron timezone `Asia/Ho_Chi_Minh`; lock key = job name, prefixed with the service (`note:...`).
- `@nestjs/locks` pinned to `0.0.1` (0.0.2 was published 2026-10-08; bump after 2026-10-22). `ioredis` `5.6.1` matches `@hocuspocus/extension-redis@4.0.0` (`~5.6.1`).
- Redis used for locks needs `maxmemory-policy noeviction` and persistence (`@nestjs/locks` doc).
- Codegen: `nx run api:gen`, `nx run proto:gen`, `nx run note:gen:sqlc`, `nx run note:gen:wire`. Never edit generated files.
- Skip obvious tests. `oxfmt` changed JS, `nx lint <project> --fix` before each commit. Conventional commits.

## Review Focus

- `X-Forwarded-Groups` missing, empty, or not JSON: user gets empty groups, admin routes return 403, nothing panics (Task 1).
- Two instances firing the same cron tick: the job runs once (Task 3 test). Jobs must still be idempotent: gocron's lock is released after the run, so a lagging instance can re-run a very short job.
- `BatchGetNoteWorkspaces` with ids of deleted notes or > 500 ids: unknown ids are omitted; > 500 rejected by protovalidate (Task 6).
- Document `UserGuard` with the real JSON header: `["a","b"]` must give `['a','b']`, not `['"a","b"']` (Task 1).
- Locks module without Redis in production mode: must start (in-memory store with `allowInMemoryStorage: true`) (Task 4).

---

### Task 1: Gateway groups/roles + `RequireGroup` (Go and NestJS)

**Files:**
- Modify: `pkg/common/http/user.go`
- Create: `pkg/common/http/group.go`
- Modify: `apps/document/src/common/user.guard.ts`

**Interfaces:**
- Produces: `commonhttp.User{ID, Email string; Groups, Roles []string}`, `commonhttp.RequireGroup(group string) gin.HandlerFunc`, `commonhttp.(*User).HasGroup(group string) bool`

- [ ] **Step 1: Go user parsing**

`pkg/common/http/user.go`:

```go
type User struct {
	ID     string   `json:"id"`
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
	Roles  []string `json:"roles"`
}

func (u *User) HasGroup(group string) bool {
	return slices.Contains(u.Groups, group)
}

func unmarshalHeaderList(v string) []string {
	var list []string
	if v == "" {
		return []string{}
	}
	if err := json.Unmarshal([]byte(v), &list); err != nil {
		return []string{}
	}
	return list
}

func GatewayUserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := &User{
			ID:     c.GetHeader("X-Forwarded-ID"),
			Email:  c.GetHeader("X-Forwarded-Email"),
			Groups: unmarshalHeaderList(c.GetHeader("X-Forwarded-Groups")),
			Roles:  unmarshalHeaderList(c.GetHeader("X-Forwarded-Roles")),
		}
		c.Set(UserCtxKey, user)
		c.Next()
	}
}
```

Imports: `encoding/json`, `slices`.

- [ ] **Step 2: `RequireGroup`**

```go
// pkg/common/http/group.go
package commonhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireGroup(group string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := UserFromContext(c)
		if !ok || !user.HasGroup(group) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "forbidden",
				"message": "requires group " + group,
			})
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 3: Fix document `UserGuard` parser**

Replace `parseHeaderList` in `apps/document/src/common/user.guard.ts`:

```ts
    const parseHeaderList = (header?: string): string[] => {
      if (!header) return [];
      try {
        const parsed: unknown = JSON.parse(header);
        return Array.isArray(parsed) ? parsed.filter((v): v is string => typeof v === 'string') : [];
      } catch {
        return [];
      }
    };
```

- [ ] **Step 4: Build, lint, commit**

```bash
go build ./pkg/... && nx lint note
oxfmt apps/document/src/common && nx lint document --fix
git add pkg/common/http apps/document/src/common
git commit -m "feat(common): parse gateway groups/roles as JSON and add RequireGroup"
```

---

### Task 2: Identity config, service accounts, admin group

**Files:**
- Create: `pkg/common/config/identity.go`
- Modify: `internal/note/config/config.go`, `internal/note/config/wire.go`, `cmd/note/.env`
- Modify: `apps/document/src/config/config.ts`, `apps/document/src/config/config.factory.ts`, `apps/document/src/app.module.ts`, `apps/document/.env` (or the env file the app loads; check `apps/document/project.json` `run` target)
- Modify: `deploy/compose/infrastructure/authentik-blueprints/users.yaml`

**Interfaces:**
- Produces: `commonconfig.Identity{AdminGroup, ServiceAccountUserID string}` provided by wire as `*commonconfig.Identity` in note; document `IDENTITY_CONFIG` token with `IdentityConfig { adminGroup: string; serviceAccountUserId: string }`.

- [ ] **Step 1: Go struct**

```go
// pkg/common/config/identity.go
package commonconfig

type Identity struct {
	AdminGroup           string `default:"notopia-admins" json:"adminGroup"           mapstructure:"admin_group"             validate:"required" yaml:"admin_group"`
	ServiceAccountUserID string `default:"-"              json:"serviceAccountUserId" mapstructure:"service_account_user_id" validate:"required" yaml:"service_account_user_id"`
}
```

- [ ] **Step 2: Note config**

`internal/note/config/config.go` `Config`: add
`Identity commonconfig.Identity \`json:"identity" mapstructure:"identity" validate:"required" yaml:"identity"\``.
`internal/note/config/wire.go`: add `"Identity"` to the `wire.FieldsOf(new(*Config), ...)` list.
`cmd/note/.env`:

```
NOTOPIA_NOTE_IDENTITY_SERVICE_ACCOUNT_USER_ID=1001
```

Run: `nx run note:gen:wire && nx build note`.

- [ ] **Step 3: Document config**

`config.ts`:

```ts
export const identityConfigSchema = z.object({
  adminGroup: z.string().min(1),
  serviceAccountUserId: z.string().min(1),
});
export type IdentityConfig = z.infer<typeof identityConfigSchema>;
```

`config.factory.ts`:

```ts
export const IDENTITY_CONFIG = Symbol('IDENTITY_CONFIG');

export const identityConfig = registerAs(IDENTITY_CONFIG, () =>
  identityConfigSchema.parse({
    adminGroup: process.env.NOTOPIA_DOCUMENT_IDENTITY_ADMIN_GROUP ?? 'notopia-admins',
    serviceAccountUserId: process.env.NOTOPIA_DOCUMENT_IDENTITY_SERVICE_ACCOUNT_USER_ID!,
  } satisfies IdentityConfig)
);
```

Add `identityConfig` to `ConfigModule.forRoot({ load: [...] })` in `app.module.ts`, and `NOTOPIA_DOCUMENT_IDENTITY_SERVICE_ACCOUNT_USER_ID=1002` to the document dev env file.

- [ ] **Step 4: Blueprint**

Append to `users.yaml` `entries` (same shape as `note-service`; no tokens):

```yaml
  - model: authentik_core.user
    id: document-service
    identifiers:
      pk: 1002
    attrs:
      username: document-service
      name: Document Service
      type: service_account
      path: goauthentik.io/service-accounts

  - model: authentik_core.user
    id: billing-service
    identifiers:
      pk: 1003
    attrs:
      username: billing-service
      name: Billing Service
      type: service_account
      path: goauthentik.io/service-accounts

  - model: authentik_core.group
    id: notopia-admins
    identifiers:
      name: notopia-admins
    attrs:
      users:
        - !Find [authentik_core.user, [username, admin]]
```

The `notopia-admins` group entry must come **after** the `admin` user entry.

- [ ] **Step 5: Verify**

`docker compose -f deploy/compose/compose.yaml up -d authentik` (or the project's usual command), wait for blueprint reconcile, log in as `admin`, and decode the access token at jwt.io: `groups` contains `notopia-admins`.

- [ ] **Step 6: Commit**

```bash
oxfmt apps/document/src && nx lint document --fix && nx lint note
git add pkg/common/config internal/note/config cmd/note/.env cmd/note/wire_gen.go apps/document deploy/compose/infrastructure/authentik-blueprints/users.yaml
git commit -m "feat: identity config, service accounts and notopia-admins group"
```

---

### Task 3: `pkg/cron` (gocron v2 + Redis lock)

**Files:**
- Create: `pkg/cron/cron.go`, `pkg/cron/locker.go`, `pkg/cron/cron_test.go`
- Modify: `go.mod`, `go.sum`

**Interfaces:**
- Produces:

```go
type Job struct {
	Name     string
	Schedule string // 5-field cron
	Timeout  time.Duration
	Run      func(ctx context.Context) error
}
func NewScheduler(locker gocron.Locker, logger *slog.Logger) (*Scheduler, error)
func (s *Scheduler) Register(jobs ...Job) error
func (s *Scheduler) Start()
func (s *Scheduler) Shutdown() error
func (s *Scheduler) RunNow(name string) error // tests and manual triggers
func NewRedisLocker(client redis.UniversalClient) (gocron.Locker, error)
```

- [ ] **Step 1: Dependencies**

```bash
go get github.com/go-co-op/gocron/v2@v2.22.0 github.com/go-co-op/gocron-redis-lock/v2@v2.2.1 github.com/alicebob/miniredis/v2@latest
go mod tidy
```

Check that `gocron-redis-lock/v2` requires `github.com/redis/go-redis/v9` (`go mod graph | grep gocron-redis-lock`).

- [ ] **Step 2: Write the failing test**

```go
// pkg/cron/cron_test.go
package cron_test

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/notopia-uit/notopia/pkg/cron"
	"github.com/redis/go-redis/v9"
)

func TestJobRunsOnceAcrossSchedulers(t *testing.T) {
	mr := miniredis.RunT(t)
	var runs atomic.Int32
	job := cron.Job{
		Name:     "test:job",
		Schedule: "0 0 1 1 *",
		Timeout:  time.Second,
		Run: func(ctx context.Context) error {
			runs.Add(1)
			time.Sleep(300 * time.Millisecond)
			return nil
		},
	}

	schedulers := make([]*cron.Scheduler, 2)
	for i := range schedulers {
		locker, err := cron.NewRedisLocker(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
		if err != nil {
			t.Fatal(err)
		}
		s, err := cron.NewScheduler(locker, slog.Default())
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Register(job); err != nil {
			t.Fatal(err)
		}
		s.Start()
		t.Cleanup(func() { _ = s.Shutdown() })
		schedulers[i] = s
	}

	var wg sync.WaitGroup
	for _, s := range schedulers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.RunNow("test:job")
		}()
	}
	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	if got := runs.Load(); got != 1 {
		t.Fatalf("runs = %d, want 1", got)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./pkg/cron/...`
Expected: FAIL, package `cron` has no `Job`/`NewScheduler`.

- [ ] **Step 4: Implement**

```go
// pkg/cron/locker.go
package cron

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
	redislock "github.com/go-co-op/gocron-redis-lock/v2"
	"github.com/redis/go-redis/v9"
)

func NewRedisLocker(client redis.UniversalClient) (gocron.Locker, error) {
	locker, err := redislock.NewRedisLocker(client, redislock.WithTries(1))
	if err != nil {
		return nil, fmt.Errorf("failed to create redis cron locker: %w", err)
	}
	return locker, nil
}
```

```go
// pkg/cron/cron.go
package cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	_ "time/tzdata"

	"github.com/go-co-op/gocron/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const DefaultLocation = "Asia/Ho_Chi_Minh"

type Job struct {
	Name     string
	Schedule string
	Timeout  time.Duration
	Run      func(ctx context.Context) error
}

type Scheduler struct {
	scheduler gocron.Scheduler
	logger    *slog.Logger
	jobs      map[string]gocron.Job
}

func NewScheduler(locker gocron.Locker, logger *slog.Logger) (*Scheduler, error) {
	loc, err := time.LoadLocation(DefaultLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to load cron location: %w", err)
	}
	s, err := gocron.NewScheduler(
		gocron.WithLocation(loc),
		gocron.WithDistributedLocker(locker),
		gocron.WithLogger(logger),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron scheduler: %w", err)
	}
	return &Scheduler{scheduler: s, logger: logger, jobs: map[string]gocron.Job{}}, nil
}

func (s *Scheduler) Register(jobs ...Job) error {
	for _, job := range jobs {
		j, err := s.scheduler.NewJob(
			gocron.CronJob(job.Schedule, false),
			gocron.NewTask(s.wrap(job)),
			gocron.WithName(job.Name),
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		)
		if err != nil {
			return fmt.Errorf("failed to register cron job %s: %w", job.Name, err)
		}
		s.jobs[job.Name] = j
	}
	return nil
}

func (s *Scheduler) wrap(job Job) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
		defer cancel()
		ctx, span := otel.Tracer("github.com/notopia-uit/notopia/pkg/cron").Start(ctx, "Cron."+job.Name)
		defer span.End()
		defer func() {
			if r := recover(); r != nil {
				span.SetStatus(codes.Error, fmt.Sprint(r))
				s.logger.ErrorContext(ctx, "cron job panicked", slog.String("job", job.Name), slog.Any("panic", r))
			}
		}()
		start := time.Now()
		if err := job.Run(ctx); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			s.logger.ErrorContext(ctx, "cron job failed", slog.String("job", job.Name), slog.Any("error", err))
			return
		}
		s.logger.InfoContext(ctx, "cron job done", slog.String("job", job.Name), slog.Duration("took", time.Since(start)))
	}
}

func (s *Scheduler) RunNow(name string) error {
	j, ok := s.jobs[name]
	if !ok {
		return fmt.Errorf("unknown cron job %s", name)
	}
	return j.RunNow()
}

func (s *Scheduler) Start() { s.scheduler.Start() }

func (s *Scheduler) Shutdown() error { return s.scheduler.Shutdown() }
```

If `gocron.WithLogger` does not accept `*slog.Logger` directly, wrap with `gocron.NewLogger` is not slog — write a 4-method adapter (`Debug/Info/Warn/Error(msg string, args ...any)`) delegating to `logger`.

- [ ] **Step 5: Run test**

Run: `go test ./pkg/cron/... -run TestJobRunsOnceAcrossSchedulers -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
go mod tidy
git add pkg/cron go.mod go.sum
git commit -m "feat(pkg): add cron scheduler with redis distributed lock"
```

---

### Task 4: NestJS locks setup (`packages/lib-server`)

**Files:**
- Create: `packages/lib-server/src/locks/redis-lock.store.ts`, `packages/lib-server/src/locks/locks-setup.module.ts`, `packages/lib-server/src/locks/index.ts`
- Create: `packages/lib-server/src/locks/redis-lock.store.spec.ts`, `packages/lib-server/vitest.config.ts`
- Modify: `packages/lib-server/package.json`, `pnpm-workspace.yaml` (catalog)

**Interfaces:**
- Produces: `@notopia-uit/lib-server/locks` exporting `LocksSetupModule.forRoot({ redisUrl?: string, ttl?: string }): DynamicModule`, `RedisLockStore`.

- [ ] **Step 1: Dependencies**

`pnpm-workspace.yaml` catalog:

```yaml
  "@nestjs/locks": "0.0.1"
  "@nestjs/schedule": "12.0.2"
  "ioredis": "5.6.1"
```

`packages/lib-server/package.json`:

```json
  "exports": {
    "./package.json": "./package.json",
    "./block-note": { "...": "unchanged" },
    "./locks": {
      "@nx/source": "./src/locks/index.ts",
      "types": "./dist/locks/index.d.ts",
      "default": "./dist/locks/index.js"
    }
  },
  "peerDependencies": {
    "@nestjs/common": "catalog:",
    "@nestjs/locks": "catalog:",
    "ioredis": "catalog:"
  }
```

(keep existing peers). Run `pnpm install`.

- [ ] **Step 2: Redis store (from the `@nestjs/locks` docs, "The store on Redis")**

```ts
// packages/lib-server/src/locks/redis-lock.store.ts
import { Injectable } from '@nestjs/common';
import { LocksStorage, type LockAcquireResult, type LockStore } from '@nestjs/locks';
import type Redis from 'ioredis';

const ACQUIRE = `
if redis.call('SET', KEYS[1], ARGV[1], 'NX', 'PX', ARGV[2]) then
  return redis.call('INCR', KEYS[2])
end
return 0
`;

const RENEW = `
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return 0
`;

const RELEASE = `
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`;

const FENCING_COUNTER = 'locks:fencing-token';

@Injectable()
export class RedisLockStore implements LockStore {
  constructor(
    private readonly redis: Redis,
    storage?: LocksStorage
  ) {
    storage?.registerSource(this);
  }

  async acquire(key: string, owner: string, ttl: number): Promise<LockAcquireResult> {
    const token = Number(await this.redis.eval(ACQUIRE, 2, this.key(key), FENCING_COUNTER, owner, ttl));
    return token > 0 ? { acquired: true, fencingToken: token } : { acquired: false };
  }

  async renew(key: string, owner: string, ttl: number): Promise<boolean> {
    return (await this.redis.eval(RENEW, 1, this.key(key), owner, ttl)) === 1;
  }

  async release(key: string, owner: string): Promise<boolean> {
    return (await this.redis.eval(RELEASE, 1, this.key(key), owner)) === 1;
  }

  private key(key: string) {
    return `locks:key:${key}`;
  }
}
```

- [ ] **Step 3: Setup module**

```ts
// packages/lib-server/src/locks/locks-setup.module.ts
import { type DynamicModule, Inject, Module, type OnApplicationShutdown } from '@nestjs/common';
import { LocksModule, LocksStorage } from '@nestjs/locks';
import Redis from 'ioredis';

import { RedisLockStore } from './redis-lock.store';

const LOCKS_REDIS = Symbol('LOCKS_REDIS');

export interface LocksSetupOptions {
  redisUrl?: string;
  ttl?: string;
}

@Module({})
export class LocksSetupModule implements OnApplicationShutdown {
  constructor(@Inject(LOCKS_REDIS) private readonly redis: Redis | null) {}

  static forRoot({ redisUrl, ttl = '30s' }: LocksSetupOptions): DynamicModule {
    if (!redisUrl) {
      return {
        module: LocksSetupModule,
        imports: [LocksModule.forRoot({ ttl, allowInMemoryStorage: true })],
        providers: [{ provide: LOCKS_REDIS, useValue: null }],
      };
    }
    return {
      module: LocksSetupModule,
      imports: [LocksModule.forRoot({ ttl })],
      providers: [
        { provide: LOCKS_REDIS, useFactory: () => new Redis(redisUrl) },
        {
          provide: RedisLockStore,
          inject: [LOCKS_REDIS, LocksStorage],
          useFactory: (redis: Redis, storage: LocksStorage) => new RedisLockStore(redis, storage),
        },
      ],
    };
  }

  async onApplicationShutdown() {
    await this.redis?.quit();
  }
}
```

```ts
// packages/lib-server/src/locks/index.ts
export * from './locks-setup.module';
export * from './redis-lock.store';
```

- [ ] **Step 4: Contract test (runs only with Redis)**

```ts
// packages/lib-server/src/locks/redis-lock.store.spec.ts
import { lockStoreContract } from '@nestjs/locks/testing';
import Redis from 'ioredis';

import { RedisLockStore } from './redis-lock.store';

const url = process.env.LOCKS_TEST_REDIS_URL;

describe.skipIf(!url)('RedisLockStore contract', () => {
  const pool = Array.from({ length: 8 }, () => new RedisLockStore(new Redis(url!)));
  let next = 0;
  for (const c of lockStoreContract(() => pool[next++ % pool.length]!, { concurrent: true })) {
    it(c.name, c.run, 60_000);
  }
});
```

`packages/lib-server/vitest.config.ts`: copy `apps/document/vitest.config.ts`, set `test.name: 'lib-server'`, `include: ['src/**/*.spec.ts']`, coverage dir `../../coverage/lib-server`.

Run: `LOCKS_TEST_REDIS_URL=redis://localhost:6373 nx test lib-server` (note_redis port from `cmd/note/.env`).
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
oxfmt packages/lib-server && nx lint lib-server --fix
git add packages/lib-server pnpm-workspace.yaml pnpm-lock.yaml
git commit -m "feat(lib-server): add nestjs locks setup with redis store"
```

---

### Task 5: 402 contract (`api/common`)

**Files:**
- Create: `api/common/components/schemas/PaymentRequiredError.yaml`, `api/common/components/responses/PaymentRequired.yaml`

**Interfaces:**
- Produces: `$ref: ../../common/components/responses/PaymentRequired.yaml` usable from path files (plan 04); codes `featureNotInPlan`, `planLimitExceeded`.

- [ ] **Step 1: Schema and response**

```yaml
# api/common/components/schemas/PaymentRequiredError.yaml
# vim: ft=yaml.openapi
allOf:
  - $ref: ./Error.yaml
  - type: object
    properties:
      details:
        type: object
        properties:
          kind:
            type: string
            enum: [feature, limit]
          key:
            type: string
            example: members
          current:
            type: integer
            format: int64
          max:
            type: integer
            format: int64
          requiredPlan:
            type: string
            enum: [plus, pro]
        required: [kind, key, requiredPlan]
```

Copy the shape of an existing file in `api/common/components/responses/` (e.g. the 403 one) for `PaymentRequired.yaml`, with description `Feature or limit not included in the workspace plan` and the schema `$ref: ../schemas/PaymentRequiredError.yaml`.

- [ ] **Step 2: Lint, commit**

```bash
nx lint api
git add api/common
git commit -m "feat(api): add PaymentRequired error contract"
```

---

### Task 6: Note `BatchGetNoteWorkspaces` RPC + document client

**Files:**
- Modify: `proto/note/note.proto`
- Modify: `internal/note/infra/persistence/pgsqlc/note_read.sql`
- Create: `internal/note/infra/persistence/pgreadmodel/notesworkspaces.go`
- Modify: `internal/note/app/readmodel.go`, `internal/note/infra/persistence/wire.go`
- Create: `internal/note/app/q_batchgetnoteworkspaces.go`
- Modify: `internal/note/app/app.go`, `internal/note/app/wire.go`, `internal/note/controller/grpc/route.go`
- Modify: `apps/document/src/note/note.service.ts`

**Interfaces:**
- Produces: gRPC `note.NoteService/BatchGetNoteWorkspaces`; `app.BatchGetNoteWorkspacesQuery` (`commonhandler.Query[BatchGetNoteWorkspaces, map[uuid.UUID]uuid.UUID]`); TS `NoteService.batchGetNoteWorkspaces(noteIds: string[]): Promise<Map<string, string>>`.

- [ ] **Step 1: Proto**

Add to `service NoteService`:

```proto
  rpc BatchGetNoteWorkspaces(BatchGetNoteWorkspacesRequest) returns (BatchGetNoteWorkspacesResponse);
```

and messages:

```proto
message BatchGetNoteWorkspacesRequest {
  repeated string note_ids = 1 [(buf.validate.field).repeated = {
    min_items: 1
    max_items: 500
    items: {string: {uuid: true}}
  }];
}

message BatchGetNoteWorkspacesResponse {
  map<string, string> workspace_ids = 1;
}
```

Run: `nx run proto:gen`.

- [ ] **Step 2: sqlc query**

Append to `note_read.sql`:

```sql
-- name: ReadGetNotesWorkspaceIDs :many
SELECT
  n.id AS note_id,
  f.workspace_id
FROM
  notes n
JOIN
  folders f ON f.id = n.folder_id
WHERE
  n.id = ANY(@ids::uuid[]);
```

Run: `nx run note:gen:sqlc`.

- [ ] **Step 3: Read model + query**

`app/readmodel.go`:

```go
type GetNotesWorkspaceIDsReadModel interface {
	Handle(ctx context.Context, noteIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error)
}
```

```go
// internal/note/infra/persistence/pgreadmodel/notesworkspaces.go
package pgreadmodel

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type NotesWorkspaceIDs struct {
	queries *pgsqlc.Queries
}

var _ app.GetNotesWorkspaceIDsReadModel = (*NotesWorkspaceIDs)(nil)

func NewNotesWorkspaceIDs(queries *pgsqlc.Queries) *NotesWorkspaceIDs {
	return &NotesWorkspaceIDs{queries: queries}
}

var ProvideNotesWorkspaceIDs = NewNotesWorkspaceIDs

func (h *NotesWorkspaceIDs) Handle(ctx context.Context, noteIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	rows, err := h.queries.ReadGetNotesWorkspaceIDs(ctx, noteIDs)
	if err != nil {
		return nil, toErr(err)
	}
	result := make(map[uuid.UUID]uuid.UUID, len(rows))
	for _, r := range rows {
		result[r.NoteID] = r.WorkspaceID
	}
	return result, nil
}
```

`persistence/wire.go` `PGReadModelProviderSet`: add `pgreadmodel.ProvideNotesWorkspaceIDs` and
`wire.Bind(new(app.GetNotesWorkspaceIDsReadModel), new(*pgreadmodel.NotesWorkspaceIDs))`.

```go
// internal/note/app/q_batchgetnoteworkspaces.go
package app

import (
	"context"

	"github.com/google/uuid"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type BatchGetNoteWorkspaces struct {
	NoteIDs []uuid.UUID
}

type BatchGetNoteWorkspacesHandler struct {
	readModel GetNotesWorkspaceIDsReadModel
}

func NewBatchGetNoteWorkspacesHandler(readModel GetNotesWorkspaceIDsReadModel) *BatchGetNoteWorkspacesHandler {
	return &BatchGetNoteWorkspacesHandler{readModel: readModel}
}

var ProvideBatchGetNoteWorkspacesHandler = NewBatchGetNoteWorkspacesHandler

type BatchGetNoteWorkspacesQuery commonhandler.Query[BatchGetNoteWorkspaces, map[uuid.UUID]uuid.UUID]

var _ BatchGetNoteWorkspacesQuery = (*BatchGetNoteWorkspacesHandler)(nil)

// Internal-only (gRPC from other services): no permission check.
func (h *BatchGetNoteWorkspacesHandler) Handle(ctx context.Context, q *BatchGetNoteWorkspaces) (map[uuid.UUID]uuid.UUID, error) {
	return h.readModel.Handle(ctx, q.NoteIDs)
}
```

Register in `app/app.go` (`Queries` field `BatchGetNoteWorkspaces BatchGetNoteWorkspacesQuery`, constructor param, `commonhandler.DecorateQuery(hp, BatchGetNoteWorkspacesHandler)`) and `app/wire.go`.

- [ ] **Step 4: gRPC route**

```go
func (s *ServiceServer) BatchGetNoteWorkspaces(
	ctx context.Context,
	req *pb.BatchGetNoteWorkspacesRequest,
) (*pb.BatchGetNoteWorkspacesResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.NoteIds))
	for _, raw := range req.NoteIds {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, errs.NewInvalid(fmt.Sprintf("invalid note id: %v", err))
		}
		ids = append(ids, id)
	}
	result, err := s.app.Queries.BatchGetNoteWorkspaces.Handle(ctx, &app.BatchGetNoteWorkspaces{NoteIDs: ids})
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(result))
	for noteID, workspaceID := range result {
		out[noteID.String()] = workspaceID.String()
	}
	return &pb.BatchGetNoteWorkspacesResponse{WorkspaceIds: out}, nil
}
```

Run: `nx run note:gen:wire && nx build note && nx lint note`.

- [ ] **Step 5: Document client**

`apps/document/src/note/note.service.ts`:

```ts
  async batchGetNoteWorkspaces(noteIds: string[]): Promise<Map<string, string>> {
    if (noteIds.length === 0) return new Map();
    const response = await firstValueFrom(
      this.noteServiceClient.batchGetNoteWorkspaces({ noteIds })
    );
    return new Map(Object.entries(response.workspaceIds));
  }
```

Callers chunk to ≤ 500 ids.

- [ ] **Step 6: Verify + commit**

With note running: `grpcurl -plaintext -d '{"note_ids":["<existing note id>","00000000-0000-4000-8000-000000000999"]}' localhost:18081 note.NoteService/BatchGetNoteWorkspaces` → map with only the existing id (use the note gRPC port from config; reflection must be on, otherwise skip and rely on plan 04 usage).

```bash
oxfmt apps/document/src/note && nx lint document --fix
git add proto/note internal/note cmd/note/wire_gen.go pkg/pb packages/pb apps/document/src/note
git commit -m "feat(note): add BatchGetNoteWorkspaces gRPC"
```

---

### Task 7: Billing permissions in authorization

**Files:**
- Modify: `internal/authorization/app/policy.csv`, `internal/authorization/app/common.go`
- Modify: `proto/authorization/authorization.proto`, `internal/authorization/controller/grpc/service.go`
- Modify: `internal/authorization/app/hasworkspacepermission_test.go`
- Modify: `internal/note/app/svc_authorization.go`, `internal/note/infra/service/authorization.go` (enum mapping, so the switch stays exhaustive)

**Interfaces:**
- Produces: `WORKSPACE_PERMISSION_BILLING_READ = 4`, `WORKSPACE_PERMISSION_BILLING_MANAGE = 5`; casbin actions `billing_read`, `billing_manage` on object `workspace`.

- [ ] **Step 1: Failing test rows**

Append to the table in `hasworkspacepermission_test.go`:

```go
		{"W111-Owner: Manage Billing", "111", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionBillingManage, true},
		{"W111-Editor: Read Billing", "112", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionBillingRead, true},
		{"W111-Editor: CANNOT Manage Billing", "112", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionBillingManage, false},
		{"W111-Viewer: Read Billing", "110", "00000000-0000-4000-8000-000000000111", app.WorkspacePermissionBillingRead, true},
```

Run: `go test ./internal/authorization/app/...` → FAIL (undefined constants).

- [ ] **Step 2: Implement**

`common.go`:

```go
	WorkspacePermissionBillingRead   WorkspacePermission = "billing_read"
	WorkspacePermissionBillingManage WorkspacePermission = "billing_manage"
```

`policy.csv`:

```
p, owner, workspace, billing_read
p, owner, workspace, billing_manage
p, editor, workspace, billing_read
p, viewer, workspace, billing_read
```

`authorization.proto` enum `WorkspacePermission`: add `WORKSPACE_PERMISSION_BILLING_READ = 4;` and `WORKSPACE_PERMISSION_BILLING_MANAGE = 5;`. Run `nx run proto:gen`.

`service.go` `pbWorkspacePermissionToApp`: add the two cases. Note's `svc_authorization.go` / `infra/service/authorization.go`: add `WorkspacePermissionBillingRead`/`BillingManage` to note's `WorkspacePermission` enum and the app→pb mapping (billing in plan 03 uses its own copy).

- [ ] **Step 3: Run tests**

Run: `go test ./internal/authorization/...` → PASS. `nx lint authorization note`.

- [ ] **Step 4: Commit**

```bash
git add internal/authorization proto/authorization pkg/pb packages/pb internal/note
git commit -m "feat(authorization): add billing_read and billing_manage permissions"
```

---

### Task 8: Docs

**Files:**
- Modify: `pkg/AGENTS.md`, `internal/authorization/AGENTS.md`, `apps/document/AGENTS.md`

- [ ] **Step 1: Update**

- `pkg/AGENTS.md`: table rows for `cron` (gocron v2 + Redis lock, `Job`, register from `controller/cron`) and `common/config.Identity`; `GatewayUserAuth` now reads `X-Forwarded-Groups`/`X-Forwarded-Roles` (JSON arrays) and `RequireGroup`.
- `internal/authorization/AGENTS.md`: policy now includes `billing_read` / `billing_manage`.
- `apps/document/AGENTS.md`: `@notopia-uit/lib-server/locks` for crons.

- [ ] **Step 2: Commit**

```bash
git add pkg/AGENTS.md internal/authorization/AGENTS.md apps/document/AGENTS.md
git commit -m "docs: document cron, identity config and billing permissions"
```
