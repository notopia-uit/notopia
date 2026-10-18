# Billing 04 — Enforcement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make plans matter: note and document read entitlements from billing over gRPC (no cache), enforce gates with 402, run plan-aware trash/revision retention, auto-commit documents on paid plans, and push billing changes to clients over note SSE.

**Architecture:** first, cron jobs become env-configurable and cron controllers stop importing Redis (`pkg/cron.Locker` injected by wire; billing refactored too). note gets a `BillingSvc` port (gRPC adapter to billing) with typed `PlanFeature`/`PlanLimit`/`PlanQuota`, a shared permanent-delete path (domain `TrashService` + app loader), a trash-purge cron, and a Kafka consumer that forwards billing notifications to the existing `WorkspaceEventHub`. document gets a billing gRPC client, revision/document columns, a transactional `commitIfDue`, a contributor flush extension, a Hocuspocus `AutoCommitExtension`, config-driven sweeper + revision-purge crons (`SchedulerRegistry` + `Locks.withLock`), and separate `redis` / `scaling` config.

**Tech Stack:** Go (gin strict server, grpc, Watermill, `pkg/cron`), NestJS 11 (TypeORM synchronize in dev, `@nestjs/schedule` 12.0.2, `@nestjs/locks` 0.0.2 via `@notopia-uit/lib-server/locks`, Hocuspocus v4.0.0, `@hocuspocus/extension-redis` 4.0.0), vitest + @suites.

**Spec:** `docs/superpowers/specs/2026-10-11-billing-04-enforcement-design.md`

**Depends on:** plans 01, 02, 03 (billing gRPC + `events.integration.billing.*`).

## Decisions made while planning (confirm in review)

1. **`published_notes` is not enforced yet.** `cmd_publishnote.go` / `cmd_publishworkspace.go` are stubs (`return nil`) and there is no publish column. The limit stays in billing; the gate lands with the publish feature (TODO in `internal/note/AGENTS.md`).
2. **Only subscription changes are pushed over SSE.** Admin plan edits are not broadcast; clients see new limits on their next entitlements fetch.
3. **Members limit** is checked against the _requested_ member list of `UpdateWorkspaceMembers` (it replaces the whole list): blocked only when the new count exceeds the limit **and** grows past the current count, so shrinking or role changes on an over-limit workspace still work.
4. **Document schema** changes rely on TypeORM `synchronize` (dev) like the existing entities; no migration files exist in `apps/document` today.

## Global Constraints

- Billing calls are uncached; billing unreachable → gates return 503 (`billingServiceUnavailable`), auto-commit skips, crons skip the batch.
- 402 body = `PaymentRequiredError` (`api/common`): `{code: featureNotInPlan|planLimitExceeded, message, details: {kind, key, current?, max?, requiredPlan}}`; `requiredPlan` = cheapest tier (`plus` < `pro`) that unlocks it.
- `-1` = unlimited; missing feature = false; missing limit = 0.
- Cron actors: note `NOTOPIA_NOTE_IDENTITY_SERVICE_ACCOUNT_USER_ID=1001`, document `NOTOPIA_DOCUMENT_IDENTITY_SERVICE_ACCOUNT_USER_ID=1002`.
- Cron time zone is never hard-coded: Go passes `time.Local`, NestJS `CronJob` omits `timeZone`; both follow the process `TZ` env (else system zone).
- Every cron job reads `name` (also the lock key), `schedule`, `timeout`, `enabled` from env with defaults; `enabled` defaults to `true` for every job. Defaults: note `note:purge-expired-trash` `0 3 * * *`; document `document:revision-purge` `30 3 * * *`, `document:sweeper` `*/10 * * * *` (idle 600 s); billing `billing:reconcile-subscriptions` `0 * * * *`, `billing:prune-provider-events` `0 4 * * *`.
- Cron controllers never import Redis or gocron; they take `pkgcron.Locker` from wire.
- `PLAN_TIER_UNSPECIFIED` (or unknown) from billing is an error in note and document; no fallback plan.
- Hocuspocus facts (v4.0.0): store payload `{clientsCount, document, documentName, instance, lastContext, lastTransactionOrigin}`; `afterStoreDocument` runs in the same `saveMutex` right after `onStoreDocument`; debounce/maxDebounce are server-wide only (defaults 2000/10000). `@hocuspocus/extension-redis` 4.0.0 takes a Redlock in `onStoreDocument`; an instance that misses it throws `SkipFurtherHooksError`, skipping every later store hook on that instance.
- Go tests use testify (`require` for preconditions, `assert` for checks); no `t.Fatal`/`t.Errorf` assertions. Skip obvious tests. `oxfmt` changed JS; `nx lint <project> --fix`; conventional commits.

## Review Focus

- Over-limit workspace after a downgrade (5 members on free, limit 3): removing a member or changing a role must still succeed (Task 2 test).
- Payer demoted/removed through `UpdateWorkspaceMembers` or leaving: blocked with 409 while billing says `requires_payer`; allowed when the subscription is lifetime or cancelling (Task 2 test).
- Two auto-commit triggers racing (afterStoreDocument + sweeper + manual): exactly one revision (row lock + interval/hash check in `commitIfDue`, Task 8 test).
- Scaling on, editors on two instances, only one wins the Hocuspocus store lock: both editors end up in `pendingContributorIds` (flush runs before the Redis extension, Task 8).
- Edits that end with identical content (type then undo): no new revision, `modified` cleared (Task 8 test).
- Revision purge on a document whose only revision is older than retention: the latest revision is kept (Task 10 test).

---

## Part 0 — shared cron plumbing

### Task 0: Configurable cron jobs, injected locker (billing refactor)

**Files:**

- Modify: `pkg/cron/locker.go` (export `Locker`), `pkg/cron/cron.go` (`Job` built from config, skip disabled)
- Create: `pkg/common/config/cron.go`
- Modify: `internal/billing/config/config.go` (`Cron` section + defaults), `internal/billing/infra/redis/redis.go` (+ `ProvideCronLocker`), `internal/billing/infra/wire.go`, `internal/billing/controller/cron/cron.go`, `cmd/billing/wire_gen.go`

**Interfaces:**

- Produces:

```go
// pkg/cron
type Locker = gocron.Locker
func NewRedisLocker(client redis.UniversalClient) (Locker, error) // unchanged body
// Register skips (and logs) a job whose Enabled is false.
type Job struct {
	Name     string
	Schedule string
	Timeout  time.Duration
	Enabled  bool
	Run      func(ctx context.Context) error
}

// pkg/common/config
type CronJob struct {
	Name     string        `mapstructure:"name"     validate:"required"`
	Schedule string        `mapstructure:"schedule" validate:"required"`
	Timeout  time.Duration `mapstructure:"timeout"  validate:"required"`
	Enabled  bool          `mapstructure:"enabled"`
}

// pkg/cron
func JobFromConfig(cfg commonconfig.CronJob, run func(ctx context.Context) error) Job
```

- [ ] **Step 1: pkg/cron + commonconfig.CronJob**

`Locker` alias, `Enabled` on `Job`, `pkgcron.JobFromConfig(cfg, run)` helper (config stays free of runtime imports). Extend `pkg/cron/cron_test.go` with one case: a disabled job is not registered (`RunNow` returns unknown job).

- [ ] **Step 2: billing config**

```go
type Cron struct {
	ReconcileSubscriptions commonconfig.CronJob `mapstructure:"reconcile_subscriptions" validate:"required"`
	PruneProviderEvents    commonconfig.CronJob `mapstructure:"prune_provider_events"   validate:"required"`
}
```

Defaults in `New` (same style as the port defaults): `billing:reconcile-subscriptions` / `0 * * * *` / 15m / enabled; `billing:prune-provider-events` / `0 4 * * *` / 5m / enabled. Env e.g. `NOTOPIA_BILLING_CRON_RECONCILE_SUBSCRIPTIONS_SCHEDULE`. Check viper binds nested env keys without a config file (`AutomaticEnv` only resolves keys viper knows; register them with `viper.SetDefault` if `Unmarshal` misses them). Expose `*config.Cron` through the config provider set like the other sections.

- [ ] **Step 3: locker in infra, controller without Redis**

`infra/redis/redis.go`: `func NewCronLocker(client *goredis.Client) (pkgcron.Locker, error) { return pkgcron.NewRedisLocker(client) }`, `var ProvideCronLocker = NewCronLocker`, add to the infra provider set.

```go
// internal/billing/controller/cron/cron.go
func New(a *app.Server, locker pkgcron.Locker, cfg *config.Cron, logger *slog.Logger) (*Cron, error) {
	scheduler, err := pkgcron.NewScheduler(locker, logger, time.Local)
	…
	err = scheduler.Register(
		pkgcron.JobFromConfig(cfg.ReconcileSubscriptions, func(ctx context.Context) error {
			return a.Cmds.ReconcileSubscriptions.Handle(ctx, &app.ReconcileSubscriptions{})
		}),
		pkgcron.JobFromConfig(cfg.PruneProviderEvents, func(ctx context.Context) error {
			return a.Cmds.PruneProviderEvents.Handle(ctx, &app.PruneProviderEvents{})
		}),
	)
	…
}
```

No `goredis` import left in `controller/`.

- [ ] **Step 4: Verify, commit**

```bash
go test ./pkg/cron/...
nx run billing:gen:wire && nx build billing && nx lint billing
git add pkg/cron pkg/common/config internal/billing cmd/billing/wire_gen.go
git commit -m "refactor(cron): env-configurable jobs and injected locker"
```

---

## Part A — note

### Task 1: Billing port, typed plan keys, 402/409/503 errors

**Files:**

- Create: `internal/note/app/svc_billing.go`, `internal/note/app/plan.go`, `internal/note/infra/service/billing.go`, `internal/note/errs/billing.go`
- Modify: `internal/note/config/config.go` (`Services.Billing`), `internal/note/infra/service/wire.go` (or the file holding `service.ProviderSet`), `internal/note/controller/http/err.go`, `cmd/note/.env`

**Interfaces:**

- Produces:

```go
type PlanTier uint8    // PlanTierUnspecified (zero), PlanTierFree, PlanTierPlus, PlanTierPro
type PlanFeature uint8 // PlanFeatureUnspecified, PlanFeatureWorkspaceGraph
type PlanLimit uint8   // PlanLimitUnspecified, PlanLimitMembers, PlanLimitNoteGraphMaxDepth, PlanLimitTrashRetentionDays
type PlanQuota int64   // PlanQuotaUnlimited = -1
// all enums: IsValid(), String() (snake-case wire key); ParsePlanFeature / ParsePlanLimit (string) (X, bool)
func (q PlanQuota) IsUnlimited() bool
func (q PlanQuota) Allows(n int64) bool // unlimited or n <= q

type Entitlements struct {
	Plan     PlanTier
	Features map[PlanFeature]bool
	Limits   map[PlanLimit]PlanQuota
}
func (e Entitlements) Feature(f PlanFeature) bool
func (e Entitlements) Limit(l PlanLimit) PlanQuota

type PlanEntitlements struct {
	Tier     PlanTier
	Features map[PlanFeature]bool
	Limits   map[PlanLimit]PlanQuota
}

type BillingSvc interface {
	GetEntitlements(ctx context.Context, workspaceID uuid.UUID) (Entitlements, error)
	BatchGetEntitlements(ctx context.Context, workspaceIDs []uuid.UUID) (map[uuid.UUID]Entitlements, error)
	GetPayer(ctx context.Context, workspaceID uuid.UUID) (payerID *string, requiresPayer bool, err error)
	ListPlans(ctx context.Context) ([]PlanEntitlements, error)
}
func RequiredPlanForFeature(plans []PlanEntitlements, f PlanFeature) PlanTier
func RequiredPlanForLimit(plans []PlanEntitlements, l PlanLimit, needed int64) PlanTier
```

- Mapper (`infra/service/billing.go`): `toAppPlanTier(pb.PlanTier) (app.PlanTier, error)`; `PLAN_TIER_UNSPECIFIED` (or any unknown value) is an internal error. Feature/limit maps: `app.ParsePlanFeature` / `app.ParsePlanLimit`; keys note does not use (`auto_commit`, …) are skipped silently.
- errs: `NewFeatureNotInPlan(key, requiredPlan string)`, `NewPlanLimitExceeded(key string, current, max int64, requiredPlan string)`, `NewBillingPayerMustTransfer(workspaceID uuid.UUID)`, `NewBillingServiceUnavailable(err error)`.

- [ ] **Step 1: Plan types + port**

```go
// internal/note/app/plan.go
package app

type PlanTier uint8

const (
	PlanTierUnspecified PlanTier = iota
	PlanTierFree
	PlanTierPlus
	PlanTierPro
)

// IsValid / String as in internal/note/domain/trashed.go ("free", "plus", "pro").

type PlanFeature uint8

const (
	PlanFeatureUnspecified PlanFeature = iota
	PlanFeatureWorkspaceGraph
)

// String: "workspace_graph". ParsePlanFeature(s string) (PlanFeature, bool).

type PlanLimit uint8

const (
	PlanLimitUnspecified PlanLimit = iota
	PlanLimitMembers
	PlanLimitNoteGraphMaxDepth
	PlanLimitTrashRetentionDays
)

// String: "members", "note_graph_max_depth", "trash_retention_days". ParsePlanLimit(s string) (PlanLimit, bool).

type PlanQuota int64

const PlanQuotaUnlimited PlanQuota = -1

func (q PlanQuota) IsUnlimited() bool { return q == PlanQuotaUnlimited }

func (q PlanQuota) Allows(n int64) bool { return q.IsUnlimited() || n <= int64(q) }

type Entitlements struct {
	Plan     PlanTier
	Features map[PlanFeature]bool
	Limits   map[PlanLimit]PlanQuota
}

func (e Entitlements) Feature(f PlanFeature) bool { return e.Features[f] }
func (e Entitlements) Limit(l PlanLimit) PlanQuota { return e.Limits[l] }

type PlanEntitlements struct {
	Tier     PlanTier
	Features map[PlanFeature]bool
	Limits   map[PlanLimit]PlanQuota
}

var paidTiers = []PlanTier{PlanTierPlus, PlanTierPro}

func planByTier(plans []PlanEntitlements, tier PlanTier) (PlanEntitlements, bool) { … }

func RequiredPlanForFeature(plans []PlanEntitlements, f PlanFeature) PlanTier {
	for _, tier := range paidTiers {
		if p, ok := planByTier(plans, tier); ok && p.Features[f] {
			return tier
		}
	}
	return PlanTierPro
}

func RequiredPlanForLimit(plans []PlanEntitlements, l PlanLimit, needed int64) PlanTier {
	for _, tier := range paidTiers {
		if p, ok := planByTier(plans, tier); ok && p.Limits[l].Allows(needed) {
			return tier
		}
	}
	return PlanTierPro
}
```

```go
// internal/note/app/svc_billing.go
type BillingSvc interface {
	GetEntitlements(ctx context.Context, workspaceID uuid.UUID) (Entitlements, error)
	BatchGetEntitlements(ctx context.Context, workspaceIDs []uuid.UUID) (map[uuid.UUID]Entitlements, error)
	GetPayer(ctx context.Context, workspaceID uuid.UUID) (payerID *string, requiresPayer bool, err error)
	ListPlans(ctx context.Context) ([]PlanEntitlements, error)
}
```

- [ ] **Step 2: Errors**

```go
// internal/note/errs/billing.go
package errs

const (
	CodeFeatureNotInPlan          Code = "featureNotInPlan"
	CodePlanLimitExceeded         Code = "planLimitExceeded"
	CodeBillingPayerMustTransfer  Code = "billingPayerMustTransfer"
	CodeBillingServiceUnavailable Code = "billingServiceUnavailable"
)

type PaymentRequiredKind string

const (
	PaymentRequiredKindFeature PaymentRequiredKind = "feature"
	PaymentRequiredKindLimit   PaymentRequiredKind = "limit"
)

type PaymentRequired struct {
	Err
	Kind         PaymentRequiredKind
	Key          string
	Current      *int64
	Max          *int64
	RequiredPlan string
}

func NewFeatureNotInPlan(key, requiredPlan string) *PaymentRequired { … }
func NewPlanLimitExceeded(key string, current, max int64, requiredPlan string) *PaymentRequired { … }

func NewBillingPayerMustTransfer(workspaceID uuid.UUID) *Err {
	return &Err{code: CodeBillingPayerMustTransfer, message: fmt.Sprintf("the billing payer of workspace %s cannot leave or lose owner role; another owner must claim billing first", workspaceID)}
}

func NewBillingServiceUnavailable(err error) *Err {
	return &Err{code: CodeBillingServiceUnavailable, message: "billing service unavailable", err: err}
}
```

`errs` stays free of app types: keys and `requiredPlan` are wire strings, so app callers pass `PlanLimitMembers.String()` / `RequiredPlanFor…(…).String()`.

`internal/note/controller/http/err.go`:

- `strictServerToHTTPErr`: `CodeBillingPayerMustTransfer` → 409, `CodeBillingServiceUnavailable` → 503, `CodeFeatureNotInPlan`/`CodePlanLimitExceeded` → 402.
- In `strictHandlerError`, before the generic path:

```go
	var pr *errs.PaymentRequired
	if errors.As(err, &pr) {
		details := gin.H{"kind": pr.Kind, "key": pr.Key, "requiredPlan": pr.RequiredPlan}
		if pr.Current != nil {
			details["current"] = *pr.Current
		}
		if pr.Max != nil {
			details["max"] = *pr.Max
		}
		c.JSON(http.StatusPaymentRequired, gin.H{"code": pr.Code().String(), "message": pr.Message(), "details": details})
		return
	}
```

- [ ] **Step 3: gRPC adapter**

`internal/note/config/config.go` `Services`: add `Billing commonconfig.Service \`json:"billing" mapstructure:"billing" validate:"required" yaml:"billing"\``.
`cmd/note/.env`:

```
NOTOPIA_NOTE_SERVICES_BILLING_URL=localhost:18084
NOTOPIA_NOTE_SERVICES_BILLING_LIVE_URL=http://localhost:28084/billing/health/live
```

`internal/note/infra/service/billing.go`: copy the connection setup of `infra/service/authorization.go` (otelgrpc stats handler, logging interceptor, insecure creds) with `servicesCfg.Billing.URL` and an error interceptor returning `errs.NewBillingServiceUnavailable(err)` for `codes.Unavailable`/`DeadlineExceeded` and wrapping others as internal. Methods:

```go
func toAppPlanTier(tier pb.PlanTier) (app.PlanTier, error) {
	switch tier {
	case pb.PlanTier_PLAN_TIER_FREE:
		return app.PlanTierFree, nil
	case pb.PlanTier_PLAN_TIER_PLUS:
		return app.PlanTierPlus, nil
	case pb.PlanTier_PLAN_TIER_PRO:
		return app.PlanTierPro, nil
	case pb.PlanTier_PLAN_TIER_UNSPECIFIED:
		return app.PlanTierUnspecified, errs.NewInternal("unspecified plan tier from billing")
	default:
		return app.PlanTierUnspecified, errs.NewInternal(fmt.Sprintf("invalid plan tier from billing: %v", tier))
	}
}

func toAppFeatures(in map[string]bool) map[app.PlanFeature]bool {
	out := make(map[app.PlanFeature]bool, len(in))
	for k, v := range in {
		if f, ok := app.ParsePlanFeature(k); ok {
			out[f] = v
		}
	}
	return out
}

func toAppLimits(in map[string]int64) map[app.PlanLimit]app.PlanQuota { … same shape with ParsePlanLimit … }

func toEntitlements(e *pb.Entitlements) (app.Entitlements, error) {
	plan, err := toAppPlanTier(e.GetPlan())
	if err != nil {
		return app.Entitlements{}, err
	}
	return app.Entitlements{Plan: plan, Features: toAppFeatures(e.GetFeatures()), Limits: toAppLimits(e.GetLimits())}, nil
}

func (b *Billing) GetEntitlements(ctx context.Context, workspaceID uuid.UUID) (app.Entitlements, error) {
	res, err := b.client.GetWorkspaceEntitlements(ctx, &pb.GetWorkspaceEntitlementsRequest{WorkspaceId: workspaceID.String()})
	if err != nil {
		return app.Entitlements{}, err
	}
	return toEntitlements(res.GetEntitlements())
}

// BatchGetEntitlements: chunks of 500 ids → BatchGetWorkspaceEntitlements, uuid.Parse each key, toEntitlements each value.
// GetPayer: return res.PayerId, res.GetRequiresPayer(), nil
// ListPlans: toAppPlanTier + toAppFeatures + toAppLimits per plan.
```

`var ProvideBilling = NewBilling`, `var _ app.BillingSvc = (*Billing)(nil)`; add to the service `ProviderSet` with `wire.Bind(new(app.BillingSvc), new(*Billing))`. Health (`controller/health/health.go`): add a `billingService` readiness check on `servicesCfg.Billing.LiveURL` like `authorizationService`.

- [ ] **Step 4: Build, commit**

```bash
nx run note:gen:wire && nx build note && nx lint note
git add internal/note cmd/note/.env cmd/note/wire_gen.go
git commit -m "feat(note): billing client, typed plan keys and payment-required errors"
```

---

### Task 2: Gates — members limit, payer guard, graph

**Files:**

- Modify: `internal/note/app/cmd_updateworkspacemembers.go`, `internal/note/app/cmd_leaveworkspace.go`, `internal/note/app/q_getnotegraph.go`, `internal/note/app/q_getworkspacegraph.go`
- Create: `internal/note/app/billinggate.go`, `internal/note/app/billinggate_test.go`

**Interfaces:**

- Consumes: `BillingSvc`, `AuthorizationSvc.GetWorkspaceMembers(ctx, userID, workspaceID) ([]WorkspaceMember, error)` (existing; check field names in `internal/note/app/model.go`).
- Produces: `checkMembersLimit(ctx, billingSvc BillingSvc, workspaceID uuid.UUID, current, requested int) error`, `checkPayerKept(ctx, billingSvc BillingSvc, workspaceID uuid.UUID, members []WorkspaceMemberUpdate) error`, `checkPayerMayLeave(ctx, billingSvc BillingSvc, workspaceID uuid.UUID, userID string) error`.

- [ ] **Step 1: Write the failing test**

```go
// internal/note/app/billinggate_test.go
package app

type fakeBilling struct {
	BillingSvc
	ent           Entitlements
	payer         *string
	requiresPayer bool
}

func (f fakeBilling) GetEntitlements(context.Context, uuid.UUID) (Entitlements, error) { return f.ent, nil }
func (f fakeBilling) GetPayer(context.Context, uuid.UUID) (*string, bool, error) {
	return f.payer, f.requiresPayer, nil
}
func (f fakeBilling) ListPlans(context.Context) ([]PlanEntitlements, error) {
	return []PlanEntitlements{
		{Tier: PlanTierPlus, Limits: map[PlanLimit]PlanQuota{PlanLimitMembers: 10}},
		{Tier: PlanTierPro, Limits: map[PlanLimit]PlanQuota{PlanLimitMembers: PlanQuotaUnlimited}},
	}, nil
}

func TestMembersLimit(t *testing.T) {
	free := fakeBilling{ent: Entitlements{Plan: PlanTierFree, Limits: map[PlanLimit]PlanQuota{PlanLimitMembers: 3}}}
	ws := uuid.New()

	var pr *errs.PaymentRequired
	err := checkMembersLimit(context.Background(), free, ws, 3, 4)
	require.ErrorAs(t, err, &pr, "growing past limit")
	assert.Equal(t, PlanTierPlus.String(), pr.RequiredPlan)
	assert.EqualValues(t, 3, *pr.Max)

	assert.NoError(t, checkMembersLimit(context.Background(), free, ws, 5, 4), "shrinking an over-limit workspace")
	assert.NoError(t, checkMembersLimit(context.Background(), free, ws, 5, 5), "same count (role change)")
}

func TestPayerGuard(t *testing.T) {
	payer := "112"
	ws := uuid.New()
	g := fakeBilling{payer: &payer, requiresPayer: true}

	demoted := []WorkspaceMemberUpdate{{ID: "112", Role: WorkspaceRoleEditor}, {ID: "111", Role: WorkspaceRoleOwner}}
	assert.True(t, isCode(checkPayerKept(context.Background(), g, ws, demoted), errs.CodeBillingPayerMustTransfer), "demoting payer")
	assert.True(t, isCode(checkPayerMayLeave(context.Background(), g, ws, "112"), errs.CodeBillingPayerMustTransfer), "payer leaving")

	noCharge := fakeBilling{payer: &payer, requiresPayer: false}
	assert.NoError(t, checkPayerMayLeave(context.Background(), noCharge, ws, "112"), "lifetime or cancelling")
}

func isCode(err error, code errs.Code) bool {
	var e errs.Error
	return errors.As(err, &e) && e.Code() == code
}
```

Run: `go test ./internal/note/app/ -run 'TestMembersLimit|TestPayerGuard'` → FAIL. (Adjust `WorkspaceMemberUpdate{ID, Role}` field names to the existing struct.)

- [ ] **Step 2: Implement gates**

```go
// internal/note/app/billinggate.go
package app

func checkMembersLimit(ctx context.Context, billingSvc BillingSvc, workspaceID uuid.UUID, current, requested int) error {
	if requested <= current {
		return nil
	}
	ent, err := billingSvc.GetEntitlements(ctx, workspaceID)
	if err != nil {
		return err
	}
	quota := ent.Limit(PlanLimitMembers)
	if quota.Allows(int64(requested)) {
		return nil
	}
	plans, err := billingSvc.ListPlans(ctx)
	if err != nil {
		return err
	}
	return errs.NewPlanLimitExceeded(
		PlanLimitMembers.String(),
		int64(requested),
		int64(quota),
		RequiredPlanForLimit(plans, PlanLimitMembers, int64(requested)).String(),
	)
}

func checkPayerKept(ctx context.Context, billingSvc BillingSvc, workspaceID uuid.UUID, members []WorkspaceMemberUpdate) error {
	payer, requiresPayer, err := billingSvc.GetPayer(ctx, workspaceID)
	if err != nil {
		return err
	}
	if payer == nil || !requiresPayer {
		return nil
	}
	for _, m := range members {
		if m.ID == *payer && m.Role == WorkspaceRoleOwner {
			return nil
		}
	}
	return errs.NewBillingPayerMustTransfer(workspaceID)
}

func checkPayerMayLeave(ctx context.Context, billingSvc BillingSvc, workspaceID uuid.UUID, userID string) error {
	payer, requiresPayer, err := billingSvc.GetPayer(ctx, workspaceID)
	if err != nil {
		return err
	}
	if requiresPayer && payer != nil && *payer == userID {
		return errs.NewBillingPayerMustTransfer(workspaceID)
	}
	return nil
}
```

Wire the gates (inject `billingSvc BillingSvc` into each handler's constructor):

- `UpdateWorkspaceMembersHandler.Handle`, after the existing owner validation:

```go
	current, err := h.authorizationSvc.GetWorkspaceMembers(ctx, cmd.UserID, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	if err := checkMembersLimit(ctx, h.billingSvc, cmd.WorkspaceID, len(current), len(cmd.Members)); err != nil {
		return err
	}
	if err := checkPayerKept(ctx, h.billingSvc, cmd.WorkspaceID, cmd.Members); err != nil {
		return err
	}
```

- `LeaveWorkspaceHandler.Handle`, before `LeaveWorkspace`: `checkPayerMayLeave(ctx, h.billingSvc, cmd.WorkspaceID, cmd.UserID)`.
- `GetNoteGraphHandler.Handle`, after the permission check and **before** `if query.Depth <= 0`:

```go
	ent, err := h.billingSvc.GetEntitlements(ctx, workspaceID)
	if err != nil {
		return Graph{}, err
	}
	maxDepth := ent.Limit(PlanLimitNoteGraphMaxDepth)
	if !maxDepth.IsUnlimited() && (query.Depth <= 0 || !maxDepth.Allows(int64(query.Depth))) {
		query.Depth = int(max(maxDepth, 1))
	}
```

A missing or `0` limit clamps to depth `1`, never `0`: `Depth <= 0` means unlimited below.

- `GetWorkspaceGraphHandler.Handle`, after the permission check:

```go
	ent, err := h.billingSvc.GetEntitlements(ctx, query.ID)
	if err != nil {
		return Graph{}, err
	}
	if !ent.Feature(PlanFeatureWorkspaceGraph) {
		plans, err := h.billingSvc.ListPlans(ctx)
		if err != nil {
			return Graph{}, err
		}
		return Graph{}, errs.NewFeatureNotInPlan(
			PlanFeatureWorkspaceGraph.String(),
			RequiredPlanForFeature(plans, PlanFeatureWorkspaceGraph).String(),
		)
	}
```

- [ ] **Step 3: Tests, build, commit**

```bash
go test ./internal/note/app/...
nx run note:gen:wire && nx build note && nx lint note
git add internal/note cmd/note/wire_gen.go
git commit -m "feat(note): enforce member limit, payer guard and graph entitlements"
```

---

### Task 3: Shared permanent delete (refactor)

**Files:**

- Modify: `internal/note/domain/trashservice.go`
- Create: `internal/note/app/permanentdelete.go`
- Modify: `internal/note/app/cmd_permanentlydeletefolder.go`, `cmd_permanentlydeleteworkspaceitems.go`, `cmd_emptytrash.go` (inject `trashService *domain.TrashService` where missing)

**Interfaces:**

- Produces:

```go
// domain, pure like TrashFoldersWithChildren / RestoreFoldersWithChildren
func (s *TrashService) PermanentlyDeleteNotes(notes []*Note, userID string)
func (s *TrashService) PermanentlyDeleteFoldersWithChildren(folders []*Folder, childNotes []*Note, userID string)

// app
type permanentDeleteTargets struct {
	notes   []*domain.Note   // requested notes + every note under the folders, deduped
	folders []*domain.Folder
}
func loadPermanentDeleteTargets(ctx context.Context, r domain.RepoRegistry, noteIDs, folderIDs []uuid.UUID) (permanentDeleteTargets, error)
func savePermanentDeleteTargets(ctx context.Context, r domain.RepoRegistry, t permanentDeleteTargets) error // notes before folders
```

Notes are deleted (and saved) before folders so each note emits `NotePermanentlyDeletedEvent` instead of disappearing through the SQL cascade.

- [ ] **Step 1: Domain + loader**

`loadPermanentDeleteTargets` is the body of today's `PermanentlyDeleteWorkspaceItemsHandler` up to the delete loop (`GetMany` with `ForUpdate`, `GetRecursiveChildrenFromFolder`, map dedupe). Folder children are passed to `PermanentlyDeleteFoldersWithChildren`; requested notes to `PermanentlyDeleteNotes`.

- [ ] **Step 2: Use it in the three handlers**

- `PermanentlyDeleteWorkspaceItems`: permission check, then load → domain → save.
- `PermanentlyDeleteFolder`: same with `folderIDs = {cmd.ID}`.
- `EmptyTrash`: collect trashed note ids and folder ids (existing `GetMany` with `TrashOnly`), then load → domain → save.

Behaviour is unchanged; no new tests (covered by the existing flows, check manually: empty trash, delete folder with nested notes → `note.deleted` per note).

- [ ] **Step 3: Verify, commit**

```bash
nx run note:gen:wire && nx build note && nx lint note
git add internal/note cmd/note/wire_gen.go
git commit -m "refactor(note): share permanent delete through TrashService"
```

---

### Task 4: Trash purge cron

**Files:**

- Create: `internal/note/app/cmd_purgeexpiredtrash.go`
- Modify: `internal/note/infra/persistence/pgsqlc/workspace_read.sql`, `internal/note/app/readmodel.go`, create `internal/note/infra/persistence/pgreadmodel/trashworkspaces.go`, modify `internal/note/infra/persistence/wire.go`
- Create: `internal/note/infra/common/redis.go` (moved from `infra/workspaceevent/redis.go`), modify `internal/note/infra/common/wire.go`, `internal/note/infra/workspaceevent/{wire.go,workspaceevent.go}`
- Modify: `internal/note/config/config.go` (`Cron`), `cmd/note/.env`
- Create: `internal/note/controller/cron/{cron.go,wire.go}`; modify `internal/note/controller/wire.go`, `internal/note/server.go`, `internal/note/app/app.go`, `wire.go`
- Modify: `deploy/compose/services/note.yaml` (`note_redis` now holds cron locks: `command: ["redis-server", "--appendonly", "yes", "--maxmemory-policy", "noeviction"]`, same as `billing_redis`)

**Interfaces:**

- Consumes: Task 0 (`pkgcron.Locker`, `pkgcron.JobFromConfig`, `commonconfig.CronJob`), Task 3 (`loadPermanentDeleteTargets`, `TrashService`), `BillingSvc.BatchGetEntitlements`, `commonconfig.Identity.ServiceAccountUserID`.
- Produces: `common.RedisClient` (shared), `common.ProvideCronLocker`; `Cmds.PurgeExpiredTrash`; `TrashWorkspacesReadModel.Handle(ctx, before time.Time) ([]uuid.UUID, error)`; job default `note:purge-expired-trash`.

- [ ] **Step 1: Shared Redis client + locker**

Move `RedisClient` / `NewRedisClient` to `infra/common/redis.go` and drop the hard-coded `DB: 1` (the SSE hub uses pub/sub, which ignores the database number). Add:

```go
func NewCronLocker(client *RedisClient) (pkgcron.Locker, error) {
	return pkgcron.NewRedisLocker(client.Client)
}

var ProvideCronLocker = NewCronLocker
```

`common.ProviderSet` gains `ProvideRedisClient`, `ProvideCronLocker`; `workspaceevent.ProviderSet` drops `ProvideRedisClient` and the hub takes `*common.RedisClient`.

- [ ] **Step 2: Read model for workspaces with old trash**

`workspace_read.sql`:

```sql
-- name: ReadWorkspaceIDsWithTrashBefore :many
SELECT DISTINCT f.workspace_id FROM folders f
WHERE f.trashed_by = 'purpose' AND f.trashed_at < @before::timestamptz
UNION
SELECT DISTINCT f.workspace_id FROM notes n JOIN folders f ON f.id = n.folder_id
WHERE n.trashed_by = 'purpose' AND n.trashed_at < @before::timestamptz;
```

`readmodel.go`: `type TrashWorkspacesReadModel interface { Handle(ctx context.Context, before time.Time) ([]uuid.UUID, error) }`; `pgreadmodel/trashworkspaces.go` implements it with `ReadWorkspaceIDsWithTrashBefore` (same shape as `pgreadmodel/notesworkspaces.go` from plan 02); bind in `persistence/wire.go`. Run `nx run note:gen:sqlc`.

- [ ] **Step 3: Purge command**

```go
// internal/note/app/cmd_purgeexpiredtrash.go
type PurgeExpiredTrash struct{}

type PurgeExpiredTrashHandler struct {
	uow          domain.UnitOfWork
	readModel    TrashWorkspacesReadModel
	billingSvc   BillingSvc
	trashService *domain.TrashService
	identity     *commonconfig.Identity
}

// constructor, ProvidePurgeExpiredTrashHandler, PurgeExpiredTrashCmd type + assertion as in other cmd files

// No permission check: cron-only, runs as the note service account.
func (h *PurgeExpiredTrashHandler) Handle(ctx context.Context, _ *PurgeExpiredTrash) error {
	now := time.Now()
	workspaceIDs, err := h.readModel.Handle(ctx, now)
	if err != nil {
		return err
	}
	if len(workspaceIDs) == 0 {
		return nil
	}
	ents, err := h.billingSvc.BatchGetEntitlements(ctx, workspaceIDs)
	if err != nil {
		return err
	}
	for _, workspaceID := range workspaceIDs {
		days := ents[workspaceID].Limit(PlanLimitTrashRetentionDays)
		if days.IsUnlimited() {
			continue
		}
		cutoff := now.Add(-time.Duration(days) * 24 * time.Hour)
		if err := h.purgeWorkspace(ctx, workspaceID, cutoff); err != nil {
			slog.ErrorContext(ctx, "trash purge failed", slog.String("workspace_id", workspaceID.String()), slog.Any("error", err))
		}
	}
	return nil
}

func (h *PurgeExpiredTrashHandler) purgeWorkspace(ctx context.Context, workspaceID uuid.UUID, cutoff time.Time) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		// GetMany folders + notes with WorkspaceID, TrashOnly, TrashedBy: domain.TrashedByPurpose;
		// keep ids whose TrashedAt() is before cutoff; none → return nil
		targets, err := loadPermanentDeleteTargets(ctx, r, noteIDs, folderIDs)
		if err != nil {
			return err
		}
		h.trashService.PermanentlyDeleteNotes(targets.notes, h.identity.ServiceAccountUserID)
		h.trashService.PermanentlyDeleteFoldersWithChildren(targets.folders, nil, h.identity.ServiceAccountUserID)
		return savePermanentDeleteTargets(ctx, r, targets)
	})
}
```

(Match the exact split between "requested notes" and "folder children" chosen in Task 3.) Register in `app.go` (`Cmds.PurgeExpiredTrash`) and `wire.go`.

- [ ] **Step 4: Config + cron controller**

`config.go`:

```go
type Cron struct {
	PurgeExpiredTrash commonconfig.CronJob `mapstructure:"purge_expired_trash" validate:"required"`
}
```

Default `note:purge-expired-trash` / `0 3 * * *` / 30m / enabled (same defaults mechanism as Task 0). Env `NOTOPIA_NOTE_CRON_PURGE_EXPIRED_TRASH_{NAME,SCHEDULE,TIMEOUT,ENABLED}`.

```go
// internal/note/controller/cron/cron.go
// NOTE: move this controller to cmd/note-worker when crons are split out of the API process.
type Cron struct {
	scheduler *pkgcron.Scheduler
}

func New(a *app.Server, locker pkgcron.Locker, cfg *config.Cron, logger *slog.Logger) (*Cron, error) {
	scheduler, err := pkgcron.NewScheduler(locker, logger, time.Local)
	if err != nil {
		return nil, err
	}
	if err := scheduler.Register(
		pkgcron.JobFromConfig(cfg.PurgeExpiredTrash, func(ctx context.Context) error {
			return a.Cmds.PurgeExpiredTrash.Handle(ctx, &app.PurgeExpiredTrash{})
		}),
	); err != nil {
		return nil, err
	}
	return &Cron{scheduler: scheduler}, nil
}

var Provide = New

func (c *Cron) Run(ctx context.Context) error {
	c.scheduler.Start()
	<-ctx.Done()
	return c.scheduler.Shutdown()
}
```

`controller/cron/wire.go`: `ProviderSet = wire.NewSet(Provide)`; add to `controller.ProviderSet`; `internal/note/server.go`: add `cron *cron.Cron` to `Server`/`NewServer` and `g.Go(func() error { return s.cron.Run(ctx) })`.

- [ ] **Step 5: Verify, commit**

```bash
nx run note:gen:sqlc && nx run note:gen:wire && nx build note && nx lint note
```

Manual: as admin set free `trash_retention_days` to `0` (`PATCH /billing/admin/plans/free`), trash a note in a free workspace, backdate it (`UPDATE notes SET trashed_at = now() - interval '2 days' WHERE id = …`), run with `NOTOPIA_NOTE_CRON_PURGE_EXPIRED_TRASH_SCHEDULE='* * * * *'` → the note is gone, `note.deleted` published, document row and Meilisearch entry removed (plan 01 consumers). Restore the plan limit. SSE still works after the Redis move.

```bash
git add internal/note cmd/note/.env cmd/note/wire_gen.go deploy/compose/services/note.yaml
git commit -m "feat(note): plan-aware trash purge cron"
```

---

### Task 5: Billing change notifications over SSE

**Files:**

- Create: `api/note/components/schemas/WorkspaceBillingChangedEvent.yaml`
- Modify: `api/note/paths/workspaces_{workspaceId}_events.yaml`
- Modify: `internal/note/app/workspaceevent.go`
- Create: `internal/note/app/e_billingchanged.go`, `internal/note/controller/event/billingchanged.go`
- Modify: `internal/note/controller/event/event.go`, `app.go`, `wire.go`

**Interfaces:**

- Consumes: `share.SubscriptionChangedEvent` (plan 02 Task 5b) on `events.integration.billing.subscription.changed`.
- Produces: SSE event `WorkspaceBillingChangedEvent {id, event, data: {workspaceId, plan}}` published on the workspace's existing channel.

- [ ] **Step 1: Schema**

```yaml
# api/note/components/schemas/WorkspaceBillingChangedEvent.yaml
# vim: ft=yaml.openapi
type: object

properties:
  id:
    type: string
    format: uuid
    example: dd57d134-d23b-457f-88fc-a4d7d6b284e5

  event:
    type: string
    enum:
      - WorkspaceBillingChangedEvent
    example: WorkspaceBillingChangedEvent

  data:
    type: object
    x-go-type-name: WorkspaceBillingChangedEventData
    description: Refetch /billing/workspaces/{id}/entitlements.
    properties:
      workspaceId:
        $ref: ./Workspace.yaml#/properties/id
      plan:
        type: string
        example: plus
    required:
      - workspaceId
      - plan

required:
  - id
  - event
  - data
```

Add it to the `oneOf` and the discriminator `mapping` in `workspaces_{workspaceId}_events.yaml`. Run `nx lint api && nx run api:gen`.

- [ ] **Step 2: Event type**

`app/workspaceevent.go`:

```go
type WorkspaceEventBillingChanged struct {
	workspaceEvent[note.WorkspaceBillingChangedEventEvent]
}
```

and a case in `NewEmptyWorkspaceEventFromType` for `note.WorkspaceBillingChangedEventEventWorkspaceBillingChangedEvent`. The hub (`infra/workspaceevent`) is unchanged.

- [ ] **Step 3: Handler + consumer**

```go
// internal/note/app/e_billingchanged.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/pkg/api/note"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type BillingChanged struct {
	WorkspaceID uuid.UUID
	Plan        PlanTier
}

type BillingChangedHandler struct {
	publisher WorkspaceEventPublisher
	identity  *commonconfig.Identity
}

func NewBillingChangedHandler(publisher WorkspaceEventPublisher, identity *commonconfig.Identity) *BillingChangedHandler {
	return &BillingChangedHandler{publisher: publisher, identity: identity}
}

var ProvideBillingChangedHandler = NewBillingChangedHandler

func (h *BillingChangedHandler) Handle(ctx context.Context, ev *BillingChanged) error {
	return h.publisher.Publish(ctx, ev.WorkspaceID, h.identity.ServiceAccountUserID, &WorkspaceEventBillingChanged{
		workspaceEvent[note.WorkspaceBillingChangedEventEvent]{
			Id:    uuid.Must(uuid.NewV7()),
			Event: note.WorkspaceBillingChangedEventEventWorkspaceBillingChangedEvent,
			Data:  note.WorkspaceBillingChangedEventData{WorkspaceId: ev.WorkspaceID, Plan: ev.Plan.String()},
		},
	})
}
```

(add `BillingChanged *BillingChangedHandler` to `Events` in `app.go` and the provider to `wire.go`.)

```go
// internal/note/controller/event/billingchanged.go
package event

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

func toAppPlanTierFromShare(p share.SubscriptionChangedEventPlan) (app.PlanTier, error) {
	switch p {
	case share.SubscriptionChangedEventPlanFree:
		return app.PlanTierFree, nil
	case share.SubscriptionChangedEventPlanPlus:
		return app.PlanTierPlus, nil
	case share.SubscriptionChangedEventPlanPro:
		return app.PlanTierPro, nil
	default:
		return app.PlanTierUnspecified, fmt.Errorf("invalid plan in SubscriptionChangedEvent: %q", p)
	}
}

func (e *Event) subscriptionChangedHandler(msg *message.Message) error {
	var ev share.SubscriptionChangedEvent
	if err := json.Unmarshal(msg.Payload, &ev); err != nil {
		return fmt.Errorf("failed to unmarshal SubscriptionChangedEvent: %w", err)
	}
	plan, err := toAppPlanTierFromShare(ev.Plan)
	if err != nil {
		slog.WarnContext(msg.Context(), "dropping billing notification", slog.Any("error", err))
		return nil
	}
	return e.app.Events.BillingChanged.Handle(msg.Context(), &app.BillingChanged{WorkspaceID: ev.WorkspaceId, Plan: plan})
}
```

`event.go` `setup()`:

```go
	e.router.AddConsumerHandler("BillingSubscriptionChangedHandler", "events.integration.billing.subscription.changed", e.subcriber, e.subscriptionChangedHandler)
```

- [ ] **Step 4: Verify, commit**

```bash
nx run note:gen:wire && nx build note && nx lint note
```

Manual: open the workspace SSE stream (`curl -N -H 'Authorization: Bearer …' http://api.notopia.localhost/note/workspaces/W/events`), complete a checkout for W → a `WorkspaceBillingChangedEvent` arrives with `workspaceId: W`.

```bash
git add api/note pkg/api/note packages/api-gen internal/note cmd/note/wire_gen.go
git commit -m "feat(note): forward billing changes to workspace SSE"
```

---

### Task 6: Plan notes in the note OpenAPI

**Files:**

- Modify: `api/note/paths/workspaces_{workspaceId}_members.yaml`, `workspaces_{workspaceId}_leave.yaml`, `notes_{noteId}_graph.yaml`, `workspaces_{workspaceId}_graph.yaml`, `workspaces_{workspaceId}_show-trash.yaml`

- [ ] **Step 1: Descriptions + responses**

Append to each operation's `description` (add `description` if missing):

| File                | Line                                                                                                                                                                                                                  |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| members (PUT/PATCH) | "Plan limit `members` (free 3, plus 10, pro unlimited). Adding beyond the limit returns 402. The billing payer cannot lose the owner role while a recurring subscription is active (409 `billingPayerMustTransfer`)." |
| leave               | "The billing payer cannot leave while a recurring subscription is active (409 `billingPayerMustTransfer`)."                                                                                                           |
| note graph          | "`depth` is capped by plan limit `note_graph_max_depth` (free 5)."                                                                                                                                                    |
| workspace graph     | "Requires feature `workspace_graph` (plus, pro); free returns 402."                                                                                                                                                   |
| show-trash          | "Trashed items older than the plan's `trash_retention_days` are purged daily."                                                                                                                                        |

Add responses where a gate applies:

```yaml
402:
  $ref: ../../common/components/responses/PaymentRequired.yaml
```

and `409` (inline `Error`) on members/leave, `503` (`InternalServerError` shape, description "Billing service unavailable") on gated operations.

- [ ] **Step 2: Generate, lint, commit**

```bash
nx lint api && nx run api:gen && nx build note
git add api/note pkg/api/note packages/api-gen apps/api-web
git commit -m "docs(api): document plan limits on note endpoints"
```

---

## Part B — document

### Task 7: Billing gRPC client in document

**Files:**

- Create: `apps/document/src/billing/{billing.module.ts,billing.service.ts,plan.ts,index.ts}`
- Modify: `apps/document/src/config/config.ts`, `config.factory.ts` (`billingUrl`), `apps/document/.env`, `apps/document/src/app.module.ts`
- Modify: `packages/pb/package.json` (export `./billing` if `nx run proto:gen` did not add it — mirror `./note`)

Named `billing` like note's `BillingSvc` and the existing `src/authorization/`: it is the client of the billing service.

**Interfaces:**

- Produces:

```ts
// plan.ts
export const PlanTier = { Free: 'free', Plus: 'plus', Pro: 'pro' } as const;
export type PlanTier = (typeof PlanTier)[keyof typeof PlanTier];
export const PlanFeature = { AutoCommit: 'auto_commit' } as const;
export type PlanFeature = (typeof PlanFeature)[keyof typeof PlanFeature];
export const PlanLimit = {
  AutoCommitMinIntervalSeconds: 'auto_commit_min_interval_seconds',
  RevisionHistoryDays: 'revision_history_days',
} as const;
export type PlanLimit = (typeof PlanLimit)[keyof typeof PlanLimit];
export const PLAN_QUOTA_UNLIMITED = -1;

export interface Entitlements {
  plan: PlanTier;
  features: Partial<Record<PlanFeature, boolean>>;
  limits: Partial<Record<PlanLimit, number>>;
}
export interface PlanEntitlements {
  tier: PlanTier;
  features: Partial<Record<PlanFeature, boolean>>;
  limits: Partial<Record<PlanLimit, number>>;
}
export const feature = (e: Pick<Entitlements, 'features'>, f: PlanFeature) =>
  e.features[f] ?? false;
export const limit = (e: Pick<Entitlements, 'limits'>, l: PlanLimit) => e.limits[l] ?? 0;
export const isUnlimited = (quota: number) => quota === PLAN_QUOTA_UNLIMITED;

// billing.service.ts
class BillingService {
  getEntitlements(workspaceId: string): Promise<Entitlements>;
  batchGetEntitlements(workspaceIds: string[]): Promise<Map<string, Entitlements>>; // chunks of 500
  listPlans(): Promise<PlanEntitlements[]>;
}
```

Mapping: `toPlanTier(PlanTierPb)` converts the generated `PlanTier` enum from `@notopia-uit/pb/billing` into the local `PlanTier`; `PLAN_TIER_UNSPECIFIED` / `UNRECOGNIZED` throws (no fallback plan, same as note). A missing `entitlements` message in the response also throws. Feature/limit maps keep only the keys listed in `PlanFeature` / `PlanLimit`.

- [ ] **Step 1: Config**

`servicesConfigSchema`: add `billingUrl: z.string()`; factory: `billingUrl: process.env.NOTOPIA_DOCUMENT_SERVICES_BILLING_GRPC_URL!`; `.env`: `NOTOPIA_DOCUMENT_SERVICES_BILLING_GRPC_URL=127.0.0.1:18084`.

- [ ] **Step 2: Module + service**

`billing.module.ts`: copy `authorization/authorization.module.ts`, replacing the package/service constants with `BILLING_PACKAGE_NAME`, `BILLING_SERVICE_NAME`, `BillingServiceService` from `@notopia-uit/pb/billing`, url `servicesCfg.billingUrl`, provider/export `BillingService`.

```ts
// apps/document/src/billing/billing.service.ts
import { Inject, Injectable, InternalServerErrorException, OnModuleInit } from '@nestjs/common';
import { type ClientGrpc } from '@nestjs/microservices';
import {
  BILLING_PACKAGE_NAME,
  BILLING_SERVICE_NAME,
  type BillingServiceClient,
  type Entitlements as EntitlementsPb,
  PlanTier as PlanTierPb,
} from '@notopia-uit/pb/billing';
import { Traceable } from 'nestjs-otel';
import { firstValueFrom } from 'rxjs';

import { type Entitlements, type PlanEntitlements, PlanFeature, PlanLimit, PlanTier } from './plan';

const toPlanTier = (tier: PlanTierPb): PlanTier => {
  switch (tier) {
    case PlanTierPb.PLAN_TIER_FREE:
      return PlanTier.Free;
    case PlanTierPb.PLAN_TIER_PLUS:
      return PlanTier.Plus;
    case PlanTierPb.PLAN_TIER_PRO:
      return PlanTier.Pro;
    default:
      throw new InternalServerErrorException(`Invalid plan tier from billing: ${tier}`);
  }
};

const pickKnown = <K extends string, V>(known: readonly K[], raw: Record<string, V>, map: (v: V) => V) => …;
// features: pickKnown(Object.values(PlanFeature), raw.features, Boolean)
// limits:   pickKnown(Object.values(PlanLimit), raw.limits, Number)  (int64 may come as Long)

const toEntitlements = (e: EntitlementsPb | undefined): Entitlements => {
  if (!e) throw new InternalServerErrorException('Missing entitlements from billing');
  return { plan: toPlanTier(e.plan), features: …, limits: … };
};

@Injectable()
@Traceable()
export class BillingService implements OnModuleInit {
  private client!: BillingServiceClient;

  constructor(@Inject(BILLING_PACKAGE_NAME) private readonly grpc: ClientGrpc) {}

  onModuleInit(): void {
    this.client = this.grpc.getService<BillingServiceClient>(BILLING_SERVICE_NAME);
  }

  async getEntitlements(workspaceId: string): Promise<Entitlements> {
    const res = await firstValueFrom(this.client.getWorkspaceEntitlements({ workspaceId }));
    return toEntitlements(res.entitlements);
  }

  async batchGetEntitlements(workspaceIds: string[]): Promise<Map<string, Entitlements>> {
    const out = new Map<string, Entitlements>();
    for (let i = 0; i < workspaceIds.length; i += 500) {
      const res = await firstValueFrom(
        this.client.batchGetWorkspaceEntitlements({ workspaceIds: workspaceIds.slice(i, i + 500) })
      );
      for (const [id, e] of Object.entries(res.entitlements)) out.set(id, toEntitlements(e));
    }
    return out;
  }

  async listPlans(): Promise<PlanEntitlements[]> {
    const res = await firstValueFrom(this.client.listPlans({}));
    return res.plans.map((p) => ({ tier: toPlanTier(p.tier), features: …, limits: … }));
  }
}
```

Import `BillingModule` in `app.module.ts`.

- [ ] **Step 3: Build, commit**

```bash
oxfmt apps/document/src && nx lint document --fix && nx build document
git add apps/document packages/pb/package.json
git commit -m "feat(document): billing grpc client with typed plan keys"
```

---

### Task 8: Revision metadata, contributors and `commitIfDue`

**Files:**

- Modify: `apps/document/src/document/document.entity.ts`, `apps/document/src/revision/revision.entity.ts`
- Modify: `apps/document/src/document/document.service.ts` (`commitIfDue`, `mergePendingContributors`, manual commit through `commitIfDue`)
- Create: `apps/document/src/hocuspocus/contributor-tracker.ts`, `apps/document/src/hocuspocus/contributor-flush.extension.ts`
- Modify: `apps/document/src/hocuspocus/hocuspocus.ts` (track contributors in `onChange`, register `ContributorFlushExtension` first)
- Modify: `apps/document/src/document/document.api.ts`, `apps/document/src/revision/revision.service.ts`, `revision.api.ts`, `api/document/` (revision schema + `trigger` filter)
- Create: `apps/document/src/document/commit-policy.spec.ts`

**Interfaces:**

- Produces:

```ts
export type RevisionTrigger = 'manual' | 'auto' | 'sweeper';
DocumentService.commitIfDue(p: {
  documentId: string;
  trigger: RevisionTrigger;
  authorId?: string;          // manual: HTTP user
  force: boolean;             // manual / sweeper
  minIntervalSeconds: number; // from entitlements, 0 for manual
}): Promise<{ revisionId: string | null }>;
DocumentService.mergePendingContributors(id: string, contributors: string[]): Promise<boolean>; // false = row not created yet
ContributorTracker.add(documentId: string, userId: string): void;
ContributorTracker.peek(documentId: string): string[]; // most recent last
ContributorTracker.clear(documentId: string, flushed: string[]): void; // removes only what was flushed
ContributorFlushExtension implements Extension // onStoreDocument
```

- [ ] **Step 1: Entities**

`document.entity.ts`:

```ts
  @Column({ type: 'timestamptz', nullable: true })
  lastCommittedAt!: Date | null;

  @UpdateDateColumn({ type: 'timestamptz' })
  updatedAt!: Date;

  @Column({ type: 'text', array: true, default: () => "'{}'" })
  pendingContributorIds!: string[];
```

`revision.entity.ts`:

```ts
  @Column({ type: 'text', nullable: true })
  authorId!: string | null;

  @Column({ type: 'enum', enum: ['manual', 'auto', 'sweeper'], default: 'manual' })
  trigger!: 'manual' | 'auto' | 'sweeper';

  @Column({ type: 'text', array: true, default: () => "'{}'" })
  contributorIds!: string[];

  @Column({ type: 'text', nullable: true })
  contentHash!: string | null;
```

TypeORM `synchronize` (non-production) adds the columns on start.

- [ ] **Step 2: Contributor tracker + flush extension**

```ts
// apps/document/src/hocuspocus/contributor-tracker.ts
@Injectable()
export class ContributorTracker {
  private readonly pending = new Map<string, string[]>();

  add(documentId: string, userId: string): void {
    const list = (this.pending.get(documentId) ?? []).filter((id) => id !== userId);
    list.push(userId);
    this.pending.set(documentId, list);
  }

  peek(documentId: string): string[] {
    return [...(this.pending.get(documentId) ?? [])];
  }

  clear(documentId: string, flushed: string[]): void {
    // keep ids added while the flush was running
    const rest = (this.pending.get(documentId) ?? []).filter((id) => !flushed.includes(id));
    if (rest.length === 0) this.pending.delete(documentId);
    else this.pending.set(documentId, rest);
  }
}
```

```ts
// apps/document/src/hocuspocus/contributor-flush.extension.ts
export class ContributorFlushExtension implements Extension {
  private readonly logger = new Logger(ContributorFlushExtension.name);

  constructor(
    private readonly tracker: ContributorTracker,
    private readonly documentService: DocumentService
  ) {}

  async onStoreDocument({ documentName }: onStoreDocumentPayload): Promise<void> {
    const contributors = this.tracker.peek(documentName);
    if (contributors.length === 0) return;
    try {
      const flushed = await this.documentService.mergePendingContributors(
        documentName,
        contributors
      );
      if (flushed) this.tracker.clear(documentName, contributors);
    } catch (err) {
      this.logger.error(
        { err, documentId: documentName },
        'contributor flush failed, retrying on next store'
      );
    }
  }
}
```

Why a separate extension, registered **first** (before the Hocuspocus Redis extension added in Task 10 and before `HocuspocusDatabase`): with scaling on, only the instance holding the Redis store lock runs the later store hooks; the others get `SkipFurtherHooksError`. Flushing earlier means every instance writes its own editors. On the very first store of a new document the row does not exist yet: `mergePendingContributors` returns `false`, the ids stay in memory and are flushed on the next store. Never rethrow: a store hook error keeps the document in memory.

In `hocuspocus.ts` `onChange`: when `data.connection` exists, `this.contributors.add(data.documentName, data.context.user.id)`. `upsertData` is unchanged.

- [ ] **Step 3: Write the failing test**

```ts
// apps/document/src/document/commit-policy.spec.ts
import { contentHash, mergeContributors, shouldCommit } from './commit-policy';

describe('commit policy', () => {
  const now = new Date('2026-10-11T10:00:00Z');

  it('respects the min interval unless forced', () => {
    const last = new Date('2026-10-11T09:58:00Z'); // 120 s ago
    expect(
      shouldCommit({
        modified: true,
        lastCommittedAt: last,
        now,
        force: false,
        minIntervalSeconds: 300,
      })
    ).toBe(false);
    expect(
      shouldCommit({
        modified: true,
        lastCommittedAt: last,
        now,
        force: true,
        minIntervalSeconds: 300,
      })
    ).toBe(true);
    expect(
      shouldCommit({
        modified: true,
        lastCommittedAt: last,
        now,
        force: false,
        minIntervalSeconds: 60,
      })
    ).toBe(true);
    expect(
      shouldCommit({
        modified: false,
        lastCommittedAt: null,
        now,
        force: true,
        minIntervalSeconds: 0,
      })
    ).toBe(false);
  });

  it('hashes equal content equally', () => {
    expect(contentHash([{ type: 'paragraph' }])).toBe(contentHash([{ type: 'paragraph' }]));
  });

  it('keeps the most recent editor last without duplicates', () => {
    expect(mergeContributors(['a', 'b'], ['c', 'a'])).toEqual(['b', 'c', 'a']);
  });
});
```

Run: `nx test document -- src/document/commit-policy.spec.ts` → FAIL.

- [ ] **Step 4: Commit policy + `commitIfDue`**

```ts
// apps/document/src/document/commit-policy.ts
import { createHash } from 'node:crypto';

export type RevisionTrigger = 'manual' | 'auto' | 'sweeper';

export const shouldCommit = (p: {
  modified: boolean;
  lastCommittedAt: Date | null;
  now: Date;
  force: boolean;
  minIntervalSeconds: number;
}): boolean => {
  if (!p.modified) return false;
  if (p.force || !p.lastCommittedAt) return true;
  return (p.now.getTime() - p.lastCommittedAt.getTime()) / 1000 >= p.minIntervalSeconds;
};

export const contentHash = (blocks: unknown): string =>
  createHash('sha256').update(JSON.stringify(blocks)).digest('hex');

export const mergeContributors = (existing: string[], incoming: string[]): string[] => {
  const out = existing.filter((id) => !incoming.includes(id));
  return [...out, ...incoming];
};
```

`document.service.ts` — replace `commitDocument` with `commitIfDue` (keep the Kafka emit and transaction shape):

```ts
  async commitIfDue({
    documentId,
    trigger,
    authorId,
    force,
    minIntervalSeconds,
  }: {
    documentId: string;
    trigger: RevisionTrigger;
    authorId?: string;
    force: boolean;
    minIntervalSeconds: number;
  }): Promise<{ revisionId: string | null }> {
    return this.dataSource.transaction(async (manager) => {
      const document = await manager.findOne(DocumentEntity, {
        where: { id: documentId },
        lock: { mode: 'pessimistic_write' },
      });
      if (!document) throw new DocumentNotFoundException(documentId);
      const now = new Date();
      if (!shouldCommit({ modified: document.modified, lastCommittedAt: document.lastCommittedAt, now, force, minIntervalSeconds })) {
        return { revisionId: null };
      }
      const editor = new BlocknoteEditorService({ schema: this.blocknoteSchema, initialContent: document.data });
      const blocks = editor.blocks();
      const hash = contentHash(blocks);
      const latest = await manager.findOne(RevisionEntity, {
        where: { document: { id: documentId } },
        order: { createdAt: 'DESC' },
      });
      const contributors = document.pendingContributorIds;
      const author = authorId ?? contributors.at(-1) ?? this.identity.serviceAccountUserId;
      if (latest?.contentHash === hash) {
        await manager.update(DocumentEntity, { id: documentId }, { modified: false, pendingContributorIds: [] });
        return { revisionId: null };
      }
      const revisionId = uuidv7();
      await manager.save(RevisionEntity, {
        id: revisionId, document, content: blocks, authorId: author, trigger, contributorIds: contributors, contentHash: hash,
      });
      await manager.update(DocumentEntity, { id: documentId }, { modified: false, lastCommittedAt: now, pendingContributorIds: [] });
      const { tags, outgoingLinkIds } = editor.extractTagsAndOutgoingLinkIds();
      await lastValueFrom(
        this.kafkaClient.emit('events.integration.document.document.committed', {
          id: documentId, userId: author, tags, outgoingLinkIds, content: blocks,
        } satisfies DocumentCommittedEvent)
      );
      return { revisionId };
    });
  }
```

Inject `@Inject(IDENTITY_CONFIG)`-backed identity (via `ConfigService.get<IdentityConfig>(IDENTITY_CONFIG)`) as `this.identity`.

`mergePendingContributors` merges in SQL so concurrent instances never lose ids:

```ts
  async mergePendingContributors(id: string, contributors: string[]): Promise<boolean> {
    const result = await this.repo.query(
      `UPDATE documents SET
         "pendingContributorIds" = ARRAY(
             SELECT x FROM unnest("pendingContributorIds") AS x
             WHERE x <> ALL($2::text[])
           ) || $2::text[]
       WHERE id = $1`,
      [id, contributors]
    );
    return result[1] > 0;
  }
```

The merge keeps order (older contributors first, the latest batch last) and never drops ids written by another instance. It does not touch `updatedAt` (the sweeper's idle check follows content stores only). Use the column names TypeORM generated (check `\d documents`; default naming keeps camelCase quoted) and check the affected-row shape `repo.query` returns for `UPDATE` with the pg driver.

Manual commit (`document.api.ts`): `const { revisionId } = await this.documentService.commitIfDue({ documentId, trigger: 'manual', authorId: user.id, force: true, minIntervalSeconds: 0 });` — when `revisionId` is `null` (nothing changed), return the latest revision id (existing response shape) or 204 per the current spec; keep the permission check that `commitDocument` had (move it into the API method if it lived there). After a commit, call `HocuspocusService.setModified(documentId, false)` (already done by the API today — keep it).

Revision list: `getRevisionsByDocumentId(documentId, page, limit, trigger?)` adds `...(trigger && { trigger })` to `where` and `order: { createdAt: 'DESC' }`; the revision schema in `api/document/components/schemas/` gains `authorId` (nullable string), `trigger` (enum), `contributorIds` (string[]); the list path gains an optional `trigger` query param. Run `nx run api:gen` and map the fields in `revision.api.ts`.

- [ ] **Step 5: Tests, build, commit**

```bash
nx test document
oxfmt apps/document/src && nx lint document --fix && nx run api:gen && nx build document
git add apps/document api/document packages/api-document-nestjs-server packages/api-gen
git commit -m "feat(document): revision author, trigger, contributors and commitIfDue"
```

---

### Task 9: Auto-commit extension

**Files:**

- Create: `apps/document/src/hocuspocus/auto-commit.extension.ts`
- Modify: `apps/document/src/hocuspocus/hocuspocus.ts` (extensions, debounce config), `apps/document/src/config/*` (`hocuspocusDebounceMs`, `hocuspocusMaxDebounceMs`), `apps/document/src/hocuspocus/hocuspocus.module.ts` (import `BillingModule`)

**Interfaces:**

- Consumes: `BillingService.getEntitlements`, `NoteService.getWorkspaceByNote({ noteId, userId })`, `DocumentService.commitIfDue`, `HocuspocusService.setModified`.
- Produces: `AutoCommitExtension implements Extension` (`afterStoreDocument`).

- [ ] **Step 1: Extension**

```ts
// apps/document/src/hocuspocus/auto-commit.extension.ts
import type { Extension, afterStoreDocumentPayload } from '@hocuspocus/server';
import { Logger } from '@nestjs/common';

import { DocumentService } from '../document/document.service';
import { BillingService, feature, limit, PlanFeature, PlanLimit } from '../billing';
import { NoteService } from '../note/note.service';
import { readModified } from './document-metadata';
import type { HocuspocusContext } from './hocuspocus-context';

export class AutoCommitExtension implements Extension {
  private readonly logger = new Logger(AutoCommitExtension.name);

  constructor(
    private readonly documentService: DocumentService,
    private readonly noteService: NoteService,
    private readonly billingService: BillingService,
    private readonly serviceAccountUserId: string,
    private readonly onCommitted: (documentId: string) => void
  ) {}

  async afterStoreDocument({
    documentName,
    document,
  }: afterStoreDocumentPayload<HocuspocusContext>) {
    if (!readModified(document)) return;
    try {
      const entity = await this.documentService.getById(documentName);
      if (!entity) return;
      const lookupUserId = entity.pendingContributorIds.at(-1) ?? this.serviceAccountUserId;
      const workspace = await this.noteService.getWorkspaceByNote({
        noteId: documentName,
        userId: lookupUserId,
      });
      const ent = await this.billingService.getEntitlements(workspace.id);
      if (!feature(ent, PlanFeature.AutoCommit)) return;
      const { revisionId } = await this.documentService.commitIfDue({
        documentId: documentName,
        trigger: 'auto',
        force: false,
        minIntervalSeconds: limit(ent, PlanLimit.AutoCommitMinIntervalSeconds),
      });
      if (revisionId) this.onCommitted(documentName);
    } catch (err) {
      this.logger.error({ err, documentId: documentName }, 'auto-commit skipped');
    }
  }
}
```

No forced commit when the last client leaves: `clientsCount` is per instance, so with scaling on it would fire while others still edit on another instance. The content is already stored; the sweeper (Task 10) turns the final edits into a revision after the idle time.

Never rethrow: an error in a store hook keeps the document in memory.

- [ ] **Step 2: Register + debounce config**

Config: `hocuspocusDebounceMs: z.number().default(2000)`, `hocuspocusMaxDebounceMs: z.number().default(10000)` from `NOTOPIA_DOCUMENT_HOCUSPOCUS_DEBOUNCE_MS` / `..._MAX_DEBOUNCE_MS`.

`hocuspocus.ts` constructor: inject `BillingService`, `ContributorTracker` and `ConfigService`; pass `debounce`/`maxDebounce`; extensions in order: `ContributorFlushExtension`, (Redis, Task 10), `HocuspocusDatabase`, `new AutoCommitExtension(documentService, noteService, billingService, identity.serviceAccountUserId, (id) => this.setModifiedFalse(id))`, where `setModifiedFalse` sets the metadata map to `{ modified: false }` on the loaded document with a normal origin (same as `HocuspocusService.setModified`; avoid a circular DI by implementing it inline on `this.hocuspocus.documents.get(id)`).

- [ ] **Step 3: Verify, commit**

```bash
oxfmt apps/document/src && nx lint document --fix && nx build document
```

Manual: workspace on `plus` (300 s interval): edit a note, wait 5 min while typing occasionally → an `auto` revision appears; close the tab → the sweeper commits it after the idle time; free workspace: no auto revisions, manual Commit still works.

```bash
git add apps/document
git commit -m "feat(document): auto-commit on paid plans via hocuspocus extension"
```

---

### Task 10: Sweeper + revision purge crons, redis and scaling config

**Files:**

- Create: `apps/document/src/cron/{cron.module.ts,cron-registrar.ts,sweeper.cron.ts,revision-purge.cron.ts,revision-purge.ts,revision-purge.spec.ts}`
- Modify: `apps/document/src/app.module.ts` (`ScheduleModule.forRoot()`, `LocksSetupModule.forRoot(...)`, `CronModule`), `apps/document/src/config/*` + `env.validation.ts` (`cron`, `redis`, `scaling` sections), `apps/document/src/hocuspocus/hocuspocus.ts` (Redis extension), `apps/document/package.json` (deps), `deploy/compose/services/document.yaml` (`document_redis`, always on)

**Interfaces:**

- Consumes: `@notopia-uit/lib-server/locks` (`LocksSetupModule`), `@nestjs/locks` `Locks.withLock`, `@nestjs/schedule` `SchedulerRegistry`, `NoteService.batchGetNoteWorkspaces`, `BillingService.batchGetEntitlements/listPlans`.
- Produces: `planRevisionDeletes(revisions: {id, documentId, createdAt}[], retentionDaysByDocument: Map<string, number>, now: Date): string[]` (pure, tested).

- [ ] **Step 1: Write the failing test**

```ts
// apps/document/src/cron/revision-purge.spec.ts
import { planRevisionDeletes } from './revision-purge';

const d = (iso: string) => new Date(iso);

describe('planRevisionDeletes', () => {
  const now = d('2026-10-11T00:00:00Z');

  it('deletes revisions older than retention but always keeps the latest', () => {
    const revisions = [
      { id: 'r1', documentId: 'a', createdAt: d('2026-09-01T00:00:00Z') },
      { id: 'r2', documentId: 'a', createdAt: d('2026-09-02T00:00:00Z') },
      { id: 'r3', documentId: 'b', createdAt: d('2026-09-01T00:00:00Z') },
    ];
    const retention = new Map([
      ['a', 7],
      ['b', 7],
    ]);
    expect(planRevisionDeletes(revisions, retention, now)).toEqual(['r1']);
  });

  it('skips unlimited and unknown documents', () => {
    const revisions = [
      { id: 'r1', documentId: 'a', createdAt: d('2020-01-01T00:00:00Z') },
      { id: 'r2', documentId: 'a', createdAt: d('2020-01-02T00:00:00Z') },
      { id: 'r3', documentId: 'gone', createdAt: d('2020-01-01T00:00:00Z') },
      { id: 'r4', documentId: 'gone', createdAt: d('2020-01-02T00:00:00Z') },
    ];
    expect(planRevisionDeletes(revisions, new Map([['a', -1]]), now)).toEqual([]);
  });
});
```

Run: `nx test document -- src/cron/revision-purge.spec.ts` → FAIL.

- [ ] **Step 2: Purge logic**

```ts
// apps/document/src/cron/revision-purge.ts
import { isUnlimited } from '../billing';

export interface RevisionRef {
  id: string;
  documentId: string;
  createdAt: Date;
}

export const planRevisionDeletes = (
  revisions: RevisionRef[],
  retentionDaysByDocument: Map<string, number>,
  now: Date
): string[] => {
  const byDoc = new Map<string, RevisionRef[]>();
  for (const r of revisions) byDoc.set(r.documentId, [...(byDoc.get(r.documentId) ?? []), r]);
  const out: string[] = [];
  for (const [documentId, list] of byDoc) {
    const days = retentionDaysByDocument.get(documentId);
    if (days === undefined || isUnlimited(days)) continue;
    const cutoff = now.getTime() - days * 86_400_000;
    const sorted = [...list].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
    for (const r of sorted.slice(1)) if (r.createdAt.getTime() < cutoff) out.push(r.id);
  }
  return out;
};
```

The cron loads, for documents having revisions older than `minDays`, **all** their revisions' `(id, documentId, createdAt)` so "latest" is correct.

- [ ] **Step 3: Crons (config-driven)**

`@Cron(expr)` is evaluated when the class loads, so it cannot read config. Jobs are registered at runtime instead; the lock key is the configured job name.

Config (`config.ts` / `config.factory.ts`):

```ts
export const cronJobConfigSchema = z.object({
  name: z.string().min(1),
  schedule: z.string().min(1),
  enabled: z.boolean(),
});
export const cronConfigSchema = z.object({
  sweeper: cronJobConfigSchema.extend({ idleSeconds: z.number().int().positive() }),
  revisionPurge: cronJobConfigSchema,
});
```

```
NOTOPIA_DOCUMENT_CRON_SWEEPER_NAME=document:sweeper                  # default
NOTOPIA_DOCUMENT_CRON_SWEEPER_SCHEDULE=*/10 * * * *                  # default
NOTOPIA_DOCUMENT_CRON_SWEEPER_ENABLED=true                           # default
NOTOPIA_DOCUMENT_CRON_SWEEPER_IDLE_SECONDS=600                       # default
NOTOPIA_DOCUMENT_CRON_REVISION_PURGE_NAME=document:revision-purge    # default
NOTOPIA_DOCUMENT_CRON_REVISION_PURGE_SCHEDULE=30 3 * * *             # default
NOTOPIA_DOCUMENT_CRON_REVISION_PURGE_ENABLED=true                    # default
```

`enabled` parses `process.env.X !== 'false'` so unset means `true`.

```ts
// apps/document/src/cron/cron-registrar.ts
// NOTE: move to a document worker when crons are split out of the API process.
@Injectable()
export class CronRegistrar implements OnModuleInit {
  private readonly logger = new Logger(CronRegistrar.name);

  constructor(
    private readonly schedulerRegistry: SchedulerRegistry,
    private readonly locks: Locks,
    @Inject(CRON_CONFIG) private readonly cronConfig: CronConfig,
    private readonly sweeper: SweeperCron,
    private readonly revisionPurge: RevisionPurgeCron
  ) {}

  onModuleInit(): void {
    this.register(this.cronConfig.sweeper, () => this.sweeper.sweep());
    this.register(this.cronConfig.revisionPurge, () => this.revisionPurge.purge());
  }

  private register(cfg: CronJobConfig, run: () => Promise<void>): void {
    if (!cfg.enabled) {
      this.logger.log({ job: cfg.name }, 'cron job disabled');
      return;
    }
    const job = CronJob.from({
      cronTime: cfg.schedule, // no timeZone: process zone from TZ
      onTick: () => this.runLocked(cfg.name, run),
    });
    this.schedulerRegistry.addCronJob(cfg.name, job);
    job.start();
  }

  private async runLocked(name: string, run: () => Promise<void>): Promise<void> {
    try {
      await this.locks.withLock(name, () => run());
    } catch (err) {
      if (err instanceof LockNotAcquiredError) {
        this.logger.debug({ job: name }, 'cron job held by another instance, skipping');
        return;
      }
      this.logger.error({ err, job: name }, 'cron job failed');
    }
  }
}
```

Lock TTL: `withLock(key, fn, { ttl })` and `LocksSetupModule.forRoot({ ttl })` both take a TTL (default `30s`). While held, the lock renews itself every `ttl / 3`, so a long purge keeps it; TTL only decides how fast another instance takes over after a crash. The module default is enough.

`SweeperCron` / `RevisionPurgeCron` below become plain `@Injectable()` services with `sweep()` / `purge()` and no decorators.

```ts
// apps/document/src/cron/revision-purge.cron.ts
@Injectable()
export class RevisionPurgeCron {
  private readonly logger = new Logger(RevisionPurgeCron.name);

  constructor(
    @InjectDataSource() private readonly dataSource: DataSource,
    private readonly noteService: NoteService,
    private readonly billingService: BillingService
  ) {}

  async purge(): Promise<void> {
    const plans = await this.billingService.listPlans();
    const days = plans
      .map((p) => limit(p, PlanLimit.RevisionHistoryDays))
      .filter((d) => !isUnlimited(d));
    if (days.length === 0) return;
    const minDays = Math.min(...days);
    const repo = this.dataSource.getRepository(RevisionEntity);
    const docRows: { documentId: string }[] = await repo
      .createQueryBuilder('r')
      .select('DISTINCT r."documentId"', 'documentId')
      .where('r."createdAt" < now() - make_interval(days => :d)', { d: minDays })
      .getRawMany();
    const documentIds = docRows.map((r) => r.documentId);
    for (let i = 0; i < documentIds.length; i += 500) {
      const chunk = documentIds.slice(i, i + 500);
      const workspaces = await this.noteService.batchGetNoteWorkspaces(chunk);
      const ents = await this.billingService.batchGetEntitlements([
        ...new Set(workspaces.values()),
      ]);
      const retention = new Map<string, number>();
      for (const [noteId, ws] of workspaces) {
        const e = ents.get(ws);
        if (e) retention.set(noteId, limit(e, PlanLimit.RevisionHistoryDays));
      }
      const revisions = await repo.find({
        select: { id: true, createdAt: true, document: { id: true } },
        relations: { document: true },
        where: { document: { id: In(chunk) } },
      });
      const ids = planRevisionDeletes(
        revisions.map((r) => ({ id: r.id, documentId: r.document.id, createdAt: r.createdAt })),
        retention,
        new Date()
      );
      if (ids.length > 0) await repo.delete({ id: In(ids) });
      this.logger.log({ deleted: ids.length }, 'revision purge batch done');
    }
  }
}
```

(`documentId` is the FK column TypeORM created for `RevisionEntity.document`; check the actual column name.) Revisions are hard-deleted (`delete`, not `softDelete`). `planRevisionDeletes` treats `PLAN_QUOTA_UNLIMITED` as "keep all".

```ts
// apps/document/src/cron/sweeper.cron.ts
@Injectable()
export class SweeperCron {
  private readonly logger = new Logger(SweeperCron.name);

  constructor(
    @InjectRepository(DocumentEntity) private readonly repo: Repository<DocumentEntity>,
    private readonly documentService: DocumentService,
    private readonly noteService: NoteService,
    private readonly billingService: BillingService,
    @Inject(CRON_CONFIG) private readonly cronConfig: CronConfig
  ) {}

  async sweep(): Promise<void> {
    const idleMs = this.cronConfig.sweeper.idleSeconds * 1000;
    const stale = await this.repo.find({
      select: { id: true },
      where: { modified: true, updatedAt: LessThan(new Date(Date.now() - idleMs)) },
      take: 500,
    });
    if (stale.length === 0) return;
    const workspaces = await this.noteService.batchGetNoteWorkspaces(stale.map((d) => d.id));
    const ents = await this.billingService.batchGetEntitlements([...new Set(workspaces.values())]);
    for (const [documentId, ws] of workspaces) {
      const e = ents.get(ws);
      if (!e || !feature(e, PlanFeature.AutoCommit)) continue;
      try {
        await this.documentService.commitIfDue({
          documentId,
          trigger: 'sweeper',
          force: true,
          minIntervalSeconds: 0,
        });
      } catch (err) {
        this.logger.error({ err, documentId }, 'sweeper commit failed');
      }
    }
  }
}
```

The sweeper is also what turns the last edits into a revision after everyone leaves (no forced commit on disconnect, Task 9).

`cron.module.ts`: imports `DocumentModule`, `NoteModule`, `BillingModule`, `TypeOrmModule.forFeature([DocumentEntity, RevisionEntity])`; providers `SweeperCron`, `RevisionPurgeCron`, `CronRegistrar`.

- [ ] **Step 4: Redis and scaling config + app module**

Two separate config sections (`config.ts` / `config.factory.ts`):

```ts
export const redisConfigSchema = z.object({ url: z.string().url().optional() });
export const scalingConfigSchema = z.object({ enabled: z.boolean() });
```

from `NOTOPIA_DOCUMENT_REDIS_URL` and `NOTOPIA_DOCUMENT_SCALING_ENABLED === 'true'` (default `false`). Cross-check in `env.validation.ts`: scaling enabled without a Redis URL fails start-up with "NOTOPIA_DOCUMENT_REDIS_URL is required when NOTOPIA_DOCUMENT_SCALING_ENABLED=true".

- Locks: `LocksSetupModule.forRoot({ redisUrl: redis.url })` — Redis store whenever a URL is set, in-memory otherwise (read the env through the same factory function, since `forRoot` is static).
- Hocuspocus: only when `scaling.enabled`, add `new Redis({ … })` from `@hocuspocus/extension-redis` **after** `ContributorFlushExtension` and before `HocuspocusDatabase` (parse host/port from the URL, or pass an ioredis instance if the extension accepts one), and set the server `name` to `os.hostname()`.

`app.module.ts` imports: `ScheduleModule.forRoot()`, `LocksSetupModule.forRoot(...)`, `CronModule`. `main.ts` already calls `app.enableShutdownHooks()`.

Dependencies (`apps/document/package.json`, via catalog): `@nestjs/schedule`, `cron`, `@nestjs/locks`, `@hocuspocus/extension-redis` (`4.0.0`), `ioredis`; `pnpm install`.

`apps/document/.env`: `NOTOPIA_DOCUMENT_REDIS_URL=redis://localhost:6374`, `NOTOPIA_DOCUMENT_SCALING_ENABLED=false`.

`deploy/compose/services/document.yaml`: add (always on, no profile, like `note_redis` / `billing_redis`)

```yaml
document_redis:
  container_name: notopia-document_redis
  image: redis:8.4.3-alpine3.22
  command: ['redis-server', '--appendonly', 'yes', '--maxmemory-policy', 'noeviction']
  ports:
    - 6374:6379
```

- [ ] **Step 5: Tests, build, verify, commit**

```bash
nx test document
oxfmt apps/document && nx lint document --fix && nx build document
```

Manual: without `NOTOPIA_DOCUMENT_REDIS_URL` the app starts with in-memory locks; `SCALING_ENABLED=true` without the URL fails start-up. With the URL and `NOTOPIA_DOCUMENT_SCALING_ENABLED=true`, run two instances on different ports: edits sync across instances; editors on both instances appear in the next revision's `contributorIds`; with `NOTOPIA_DOCUMENT_CRON_SWEEPER_SCHEDULE='* * * * *'` the sweeper runs on one instance per tick; `NOTOPIA_DOCUMENT_CRON_SWEEPER_ENABLED=false` → not registered.

```bash
git add apps/document deploy/compose/services/document.yaml pnpm-workspace.yaml pnpm-lock.yaml
git commit -m "feat(document): config-driven crons, separate redis and scaling config"
```

---

### Task 11: Docs, API descriptions, web outline, E2E

**Files:**

- Modify: `apps/document/AGENTS.md`, `internal/note/AGENTS.md`, `api/document/paths/*` (commit + revisions descriptions)
- Create: `docs/superpowers/notes/billing-web-outline.md` (handoff for the frontend owner)

- [ ] **Step 1: Docs**

- `apps/document/AGENTS.md`: commit pipeline now `commitIfDue` (manual/auto/sweeper), contributors (`ContributorFlushExtension` before the Redis extension), `AutoCommitExtension` (no forced commit on disconnect), config-driven crons (`CronRegistrar`), separate `redis` / `scaling` config; remove "Commit is currently manual only".
- `internal/note/AGENTS.md`: entitlement gates (members, payer guard, graph), trash purge cron, billing SSE event, TODO: gate `published_notes` when publishing is implemented.
- `api/document` descriptions: commit → "Plus/pro workspaces also auto-commit (`auto_commit`)"; revisions → "Revisions older than plan `revision_history_days` are purged daily; the latest is always kept. Filter with `trigger`."

- [ ] **Step 2: Web outline (no code)**

`docs/superpowers/notes/billing-web-outline.md`: billing settings tab (plan, status, renewal/access-until, payer + claim button, Upgrade → checkout URL, Manage → portal URL, Cancel now with `GET …/cancel-now/preview` amount); pricing page from `GET /billing/plans` (VND formatting `Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' })`); global 402 handler on `@notopia-uit/api-gen` → upgrade dialog with `details.requiredPlan`; 409 `billingPayerMustTransfer` message; `<FeatureGate>` reading entitlements, refetched on SSE `WorkspaceBillingChangedEvent` and polled every 2 s (max 20 s) after returning with `?checkout=success|payer` or `?portal=return`; downgrade warning about retention; editor "Saved automatically" when `auto_commit`; revision list collapses `trigger=auto`; admin pages for plans/prices/policy (group `notopia-admins`).

- [ ] **Step 3: End-to-end checklist**

1. Free workspace with 3 members → adding a 4th → 402 `requiredPlan: plus`.
2. Upgrade via checkout → SSE `WorkspaceBillingChangedEvent` → adding the 4th member succeeds.
3. Free workspace graph → 402; note graph depth > 5 is clamped.
4. Payer demotion → 409; another owner claims payer (plan 03) → demotion allowed.
5. Plus workspace: auto revisions every ≥ 300 s while editing; after closing the tab the sweeper commits within the idle time; identical content → no revision.
6. Trash purge and revision purge with shortened retention (admin API) and a `* * * * *` schedule from env; `…_ENABLED=false` disables a job.
7. Stop billing → member add returns 503; editing continues (auto-commit skipped, logged).

- [ ] **Step 4: Commit**

```bash
nx lint api && nx run api:gen
git add apps/document/AGENTS.md internal/note/AGENTS.md api/document docs/superpowers/notes packages/api-document-nestjs-server packages/api-gen
git commit -m "docs: enforcement notes, api descriptions and web outline"
```
