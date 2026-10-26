# Billing 02 — Shared foundations

> Sub-project 2 of 4. Can run in parallel with 01. Sub-projects 03 and 04 depend on it.
> Order: 01 + 02 (parallel) → 03 → 04. Progress is tracked by the task checkboxes in each plan under `docs/superpowers/plans/`.

## Goal

Add the cross-cutting pieces billing and enforcement need, without any billing logic:
cron with distributed locks (Go + NestJS), admin group + service-account identity, the 402
error contract, shared event schemas, note batch lookup RPC, billing permissions.

## Components

### F1 — Gateway groups and admin guard (`pkg/common/http`)

- `User.Groups []string` and `User.Roles []string`, from `X-Forwarded-Groups` /
  `X-Forwarded-Roles` (already mapped in `traefik/dynamic/jwt-middleware.yaml`). The plugin
  forwards array claims as a **JSON array string** (e.g. `["notopia-admins","authentik Admins"]`):
  `json.Unmarshal` into `[]string`; empty or invalid → empty slice. Same approach already used
  in another project of the team:

  ```go
  func unmarshalHeader(v string) []string {
  	var list []string
  	if v == "" {
  		return []string{}
  	}
  	if err := json.Unmarshal([]byte(v), &list); err != nil {
  		return []string{}
  	}
  	return list
  }
  ```

  NestJS `UserGuard` / `@User()` in `apps/document/src/common/` parses the same way.
- `RequireGroup(group string) gin.HandlerFunc` → `403` with `Error{code: "forbidden"}` when
  missing.

### F2 — Service config additions (`pkg/common/config`)

New embeddable struct, used by note, document (zod equivalent) and billing:

```go
type Identity struct {
	AdminGroup           string `default:"notopia-admins" json:"adminGroup" mapstructure:"admin_group" validate:"required" yaml:"adminGroup"`
	ServiceAccountUserID string `default:"-" json:"serviceAccountUserId" mapstructure:"service_account_user_id" validate:"required" yaml:"serviceAccountUserId"`
}
```

Env: `NOTOPIA_{SERVICE}_IDENTITY_ADMIN_GROUP`, `NOTOPIA_{SERVICE}_IDENTITY_SERVICE_ACCOUNT_USER_ID`
(viper prefix + `.`→`_`). Wire into note config now; document gets
`NOTOPIA_DOCUMENT_IDENTITY_*` in its zod config.

The service account id replaces any `"system"` actor: crons and system commands pass it as
`userID`. User ids are Authentik pks as strings (`sub_mode: user_id`), so `"1001"` fits
`BaseEvent.UserID string`.

### F3 — Authentik blueprint (`deploy/compose/infrastructure/authentik-blueprints/users.yaml`)

- Service accounts, same shape as `note-service` (pk 1001):
  `document-service` (pk 1002), `billing-service` (pk 1003). Tokens only if a service calls
  the Authentik API (billing does not; skip tokens unless needed).
- Group `notopia-admins`; add user `admin` (pk 101) to it for local testing.
- `.env` files: note `…_SERVICE_ACCOUNT_USER_ID=1001`, document `1002`.

### F4 — Go cron (`pkg/cron`, new)

```go
type Job struct {
	Name     string        // also the lock key
	Schedule string        // cron expression
	Timeout  time.Duration
	Run      func(ctx context.Context) error
}

type Scheduler interface {
	Register(jobs ...Job) error
	Start()
	Shutdown() error
}

func NewScheduler(locker gocron.Locker, logger *slog.Logger, location *time.Location) (Scheduler, error)
func NewRedisLocker(client redis.UniversalClient, opts ...redislock.LockerOption) (gocron.Locker, error) // gocron-redis-lock
```

- `github.com/go-co-op/gocron/v2`, `github.com/go-co-op/gocron-redis-lock/v2` (pin exact
  versions; check the module path supports go-redis v9, which the repo uses).
- Location passed in by the caller (never hard-coded); services pass `time.Local`, which Go resolves from `TZ` or the system zone (`time/tzdata` embedded). `WithSingletonMode(LimitModeReschedule)` per job (no
  self-overlap), otel span `Cron.{Name}` + slog around each run, panic recovery.
- Distributed lock on every job via `gocron.WithDistributedJobLocker`.
- Services register jobs from `internal/{svc}/controller/cron/` and start/stop the scheduler
  in their `Server.Run` lifecycle. Each controller file carries
  `// NOTE: move to cmd/{svc}-worker when crons are split out of the API process.`
- Test: two schedulers on one Redis run a job once per tick.

### F5 — NestJS locks (`packages/lib-server`)

- `@nestjs/locks` + `@nestjs/schedule` (exact versions, added to the pnpm catalog).
- `RedisLockStore` — the Lua `ACQUIRE`/`RENEW`/`RELEASE` store from the `@nestjs/locks` docs
  ("The store on Redis"), using `ioredis` `eval`, keys `locks:key:{key}`, counter
  `locks:fencing-token`.
- `LocksSetupModule.forRoot({ redisUrl?: string })`: no URL → `LocksModule.forRoot({ allowInMemoryStorage: true })`
  (single instance); URL → `LocksModule.forRoot()` + `RedisLockStore` registered via
  `LocksStorage.registerSource`. Closes Redis in `onApplicationShutdown`.
- Jobs use `@Cron(expr, { name })` (no `timeZone`: process zone from `TZ`) +
  `@OnOneInstance({ key })` + `@WithoutOverlapping()`; services call
  `app.enableShutdownHooks()`.
- Test the store with `lockStoreContract()` from `@nestjs/locks/testing` against Redis.
- Redis for locks needs `maxmemory-policy noeviction` and persistence (doc requirement).

### F6 — 402 contract (`api/common`)

- `components/schemas/PaymentRequiredError.yaml`:
  `allOf: [Error]` + `details {kind: feature|limit, key, current?, max?, requiredPlan: plus|pro}`
  (`required: [kind, key, requiredPlan]`).
- `components/responses/PaymentRequired.yaml` (402) referencing it.
- Codes: `featureNotInPlan`, `planLimitExceeded`.
- `nx lint api` clean; nothing references it yet (sub-project 4 does).

### F7 — Shared event schemas (`api/share`)

- `SubscriptionChangedEvent {workspaceId: uuid, plan: free|plus|pro, status: string}` — best-effort client notification only (no consumer
  caches; see spec 03 section 10). (`WorkspaceDeletedIntegrationEvent` is added in sub-project 01.)
- `nx gen api` → `@notopia-uit/api-share-gen` and `pkg/api/share`.

### F8 — Note batch lookup RPC (`proto/note/note.proto`)

```proto
rpc BatchGetNoteWorkspaces(BatchGetNoteWorkspacesRequest) returns (BatchGetNoteWorkspacesResponse);

message BatchGetNoteWorkspacesRequest {
  repeated string note_ids = 1 [(buf.validate.field).repeated = {min_items: 1, max_items: 500, items: {string: {uuid: true}}}];
}
message BatchGetNoteWorkspacesResponse { map<string, string> workspace_ids = 1; }
```

- Read-model query `q_batchgetnoteworkspaces.go` (one `SELECT id, workspace_id FROM notes WHERE id = ANY($1)`,
  includes trashed notes, excludes hard-deleted). No permission check: internal service call.
- Implement in `internal/note/controller/grpc`, add client method in
  `apps/document/src/note/`.

### F9 — Billing permissions (`internal/authorization`)

- `policy.csv`: `billing_read` for owner, editor, viewer; `billing_manage` for owner, on
  object `workspace`.
- Extend the permission enum in `authorization.proto`, `nx gen proto`, map in
  `controller/grpc/service.go`; extend `model_test.go` / `policy_test.csv`.

## Out of scope

Any billing domain code, entitlement checks, cron jobs themselves (only the scheduler and lock
plumbing).

## Testing

Only the lock behaviour (pkg/cron two-scheduler test, `lockStoreContract()`) is tested.
Everything else is verified by `nx lint`, `nx gen api` / `nx gen proto` producing no diff, and
use in sub-projects 03/04.

## Acceptance

- `pkg/cron` and `packages/lib-server` locks usable with a one-line registration.
- Admin request with group `notopia-admins` passes `RequireGroup`; others get 403.
- `BatchGetNoteWorkspaces` returns a map for existing notes and omits unknown ids.
- `pkg/AGENTS.md` lists `cron` and the new `common/config.Identity`.
