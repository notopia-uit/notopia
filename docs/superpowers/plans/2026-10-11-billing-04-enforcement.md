# Billing 04 — Enforcement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make plans matter: note and document read entitlements from billing over gRPC (no cache), enforce gates with 402, run plan-aware trash/revision retention, auto-commit documents on paid plans, and push billing changes to clients over note SSE.

**Architecture:** note gets an `EntitlementSvc` port (gRPC adapter to billing) used by command/query handlers, a trash-purge cron (`pkg/cron`), and a Kafka consumer that forwards billing notifications to the existing `WorkspaceEventHub`. document gets a billing gRPC client, revision/document columns, a transactional `commitIfDue`, a Hocuspocus `AutoCommitExtension`, sweeper + revision-purge crons (`@nestjs/locks`), and an optional Redis scaling mode.

**Tech Stack:** Go (gin strict server, grpc, Watermill, `pkg/cron`), NestJS 11 (TypeORM synchronize in dev, `@nestjs/schedule` 6.1.3, `@nestjs/locks` 0.0.2 via `@notopia-uit/lib-server/locks`, Hocuspocus v4.0.0, `@hocuspocus/extension-redis` 4.0.0), vitest + @suites.

**Spec:** `docs/superpowers/specs/2026-10-11-billing-04-enforcement-design.md`

**Depends on:** plans 01, 02, 03 (billing gRPC + `events.integration.billing.*`).

## Decisions made while planning (confirm in review)

1. **`published_notes` is not enforced yet.** `cmd_publishnote.go` / `cmd_publishworkspace.go` are stubs (`return nil`) and there is no publish column. The limit stays in billing; the gate lands with the publish feature (TODO in `internal/note/AGENTS.md`).
2. **Plan-change broadcast** uses a reserved hub key `workspace:broadcast` (metadata workspace id = `uuid.Nil`); every SSE subscriber also listens on it.
3. **Members limit** is checked against the *requested* member list of `UpdateWorkspaceMembers` (it replaces the whole list): blocked only when the new count exceeds the limit **and** grows past the current count, so shrinking or role changes on an over-limit workspace still work.
4. **Document schema** changes rely on TypeORM `synchronize` (dev) like the existing entities; no migration files exist in `apps/document` today.

## Global Constraints

- Billing calls are uncached; billing unreachable → gates return 503 (`billingServiceUnavailable`), auto-commit skips, crons skip the batch.
- 402 body = `PaymentRequiredError` (`api/common`): `{code: featureNotInPlan|planLimitExceeded, message, details: {kind, key, current?, max?, requiredPlan}}`; `requiredPlan` = cheapest tier (`plus` < `pro`) that unlocks it.
- `-1` = unlimited; missing feature = false; missing limit = 0.
- Cron actors: note `NOTOPIA_NOTE_IDENTITY_SERVICE_ACCOUNT_USER_ID=1001`, document `NOTOPIA_DOCUMENT_IDENTITY_SERVICE_ACCOUNT_USER_ID=1002`.
- Cron time zone `Asia/Ho_Chi_Minh`: note trash purge `0 3 * * *`; document revision purge `30 3 * * *`; document sweeper `*/10 * * * *`.
- Hocuspocus facts (v4.0.0): store payload `{clientsCount, document, documentName, instance, lastContext, lastTransactionOrigin}`; `afterStoreDocument` runs in the same `saveMutex` right after `onStoreDocument`; debounce/maxDebounce are server-wide only (defaults 2000/10000).
- Skip obvious tests. `oxfmt` changed JS; `nx lint <project> --fix`; conventional commits.

## Review Focus

- Over-limit workspace after a downgrade (5 members on free, limit 3): removing a member or changing a role must still succeed (Task 2 test).
- Payer demoted/removed through `UpdateWorkspaceMembers` or leaving: blocked with 409 while billing says `requires_payer`; allowed when the subscription is lifetime or cancelling (Task 2 test).
- Two auto-commit triggers racing (afterStoreDocument + sweeper + manual): exactly one revision (row lock + interval/hash check in `commitIfDue`, Task 7 test).
- Edits that end with identical content (type then undo): no new revision, `modified` cleared (Task 7 test).
- Revision purge on a document whose only revision is older than retention: the latest revision is kept (Task 9 test).

---

## Part A — note

### Task 1: Entitlement port, billing client, 402/409/503 errors

**Files:**
- Create: `internal/note/app/svc_entitlement.go`, `internal/note/infra/service/billing.go`, `internal/note/errs/billing.go`
- Modify: `internal/note/config/config.go` (`Services.Billing`), `internal/note/infra/service/wire.go` (or the file holding `service.ProviderSet`), `internal/note/controller/http/err.go`, `cmd/note/.env`

**Interfaces:**
- Produces:

```go
type Entitlements struct {
	Plan     string
	Features map[string]bool
	Limits   map[string]int64
}
func (e Entitlements) Feature(key string) bool
func (e Entitlements) Limit(key string) int64

type PlanEntitlements struct {
	Tier     string
	Features map[string]bool
	Limits   map[string]int64
}

type EntitlementSvc interface {
	Get(ctx context.Context, workspaceID uuid.UUID) (Entitlements, error)
	BatchGet(ctx context.Context, workspaceIDs []uuid.UUID) (map[uuid.UUID]Entitlements, error)
	GetPayer(ctx context.Context, workspaceID uuid.UUID) (payerID *string, requiresPayer bool, err error)
	ListPlans(ctx context.Context) ([]PlanEntitlements, error)
}
func RequiredPlanForFeature(plans []PlanEntitlements, key string) string
func RequiredPlanForLimit(plans []PlanEntitlements, key string, needed int64) string
```
- errs: `NewFeatureNotInPlan(key, requiredPlan string)`, `NewPlanLimitExceeded(key string, current, max int64, requiredPlan string)`, `NewBillingPayerMustTransfer(workspaceID uuid.UUID)`, `NewBillingServiceUnavailable(err error)`.

- [ ] **Step 1: Port + helpers**

```go
// internal/note/app/svc_entitlement.go
package app

import (
	"context"

	"github.com/google/uuid"
)

const (
	FeatureWorkspaceGraph    = "workspace_graph"
	LimitMembers             = "members"
	LimitNoteGraphMaxDepth   = "note_graph_max_depth"
	LimitTrashRetentionDays  = "trash_retention_days"
	Unlimited          int64 = -1
)

type Entitlements struct {
	Plan     string
	Features map[string]bool
	Limits   map[string]int64
}

func (e Entitlements) Feature(key string) bool { return e.Features[key] }
func (e Entitlements) Limit(key string) int64  { return e.Limits[key] }

type PlanEntitlements struct {
	Tier     string
	Features map[string]bool
	Limits   map[string]int64
}

type EntitlementSvc interface {
	Get(ctx context.Context, workspaceID uuid.UUID) (Entitlements, error)
	BatchGet(ctx context.Context, workspaceIDs []uuid.UUID) (map[uuid.UUID]Entitlements, error)
	GetPayer(ctx context.Context, workspaceID uuid.UUID) (payerID *string, requiresPayer bool, err error)
	ListPlans(ctx context.Context) ([]PlanEntitlements, error)
}

var paidTiers = []string{"plus", "pro"}

func planByTier(plans []PlanEntitlements, tier string) (PlanEntitlements, bool) {
	for _, p := range plans {
		if p.Tier == tier {
			return p, true
		}
	}
	return PlanEntitlements{}, false
}

func RequiredPlanForFeature(plans []PlanEntitlements, key string) string {
	for _, tier := range paidTiers {
		if p, ok := planByTier(plans, tier); ok && p.Features[key] {
			return tier
		}
	}
	return "pro"
}

func RequiredPlanForLimit(plans []PlanEntitlements, key string, needed int64) string {
	for _, tier := range paidTiers {
		if p, ok := planByTier(plans, tier); ok {
			if l := p.Limits[key]; l == Unlimited || l >= needed {
				return tier
			}
		}
	}
	return "pro"
}
```

- [ ] **Step 2: Errors**

```go
// internal/note/errs/billing.go
package errs

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	CodeFeatureNotInPlan          Code = "featureNotInPlan"
	CodePlanLimitExceeded         Code = "planLimitExceeded"
	CodeBillingPayerMustTransfer  Code = "billingPayerMustTransfer"
	CodeBillingServiceUnavailable Code = "billingServiceUnavailable"
)

type PaymentRequired struct {
	Err
	Kind         string
	Key          string
	Current      *int64
	Max          *int64
	RequiredPlan string
}

func NewFeatureNotInPlan(key, requiredPlan string) *PaymentRequired {
	return &PaymentRequired{
		Err:  Err{code: CodeFeatureNotInPlan, message: fmt.Sprintf("feature %q requires plan %s", key, requiredPlan)},
		Kind: "feature", Key: key, RequiredPlan: requiredPlan,
	}
}

func NewPlanLimitExceeded(key string, current, max int64, requiredPlan string) *PaymentRequired {
	return &PaymentRequired{
		Err:  Err{code: CodePlanLimitExceeded, message: fmt.Sprintf("limit %q exceeded (%d/%d), requires plan %s", key, current, max, requiredPlan)},
		Kind: "limit", Key: key, Current: &current, Max: &max, RequiredPlan: requiredPlan,
	}
}

func NewBillingPayerMustTransfer(workspaceID uuid.UUID) *Err {
	return &Err{code: CodeBillingPayerMustTransfer, message: fmt.Sprintf("the billing payer of workspace %s cannot leave or lose owner role; another owner must claim billing first", workspaceID)}
}

func NewBillingServiceUnavailable(err error) *Err {
	return &Err{code: CodeBillingServiceUnavailable, message: "billing service unavailable", err: err}
}
```

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
NOTOPIA_NOTE_SERVICES_BILLING_URL=localhost:18083
NOTOPIA_NOTE_SERVICES_BILLING_LIVE_URL=http://localhost:28083/billing/health/live
```

`internal/note/infra/service/billing.go`: copy the connection setup of `infra/service/authorization.go` (otelgrpc stats handler, logging interceptor, insecure creds) with `servicesCfg.Billing.URL` and an error interceptor returning `errs.NewBillingServiceUnavailable(err)` for `codes.Unavailable`/`DeadlineExceeded` and wrapping others as internal. Methods:

```go
func toEntitlements(e *pb.Entitlements) app.Entitlements {
	return app.Entitlements{Plan: e.GetPlan(), Features: e.GetFeatures(), Limits: e.GetLimits()}
}

func (b *Billing) Get(ctx context.Context, workspaceID uuid.UUID) (app.Entitlements, error) {
	res, err := b.client.GetWorkspaceEntitlements(ctx, &pb.GetWorkspaceEntitlementsRequest{WorkspaceId: workspaceID.String()})
	if err != nil {
		return app.Entitlements{}, err
	}
	return toEntitlements(res.GetEntitlements()), nil
}

func (b *Billing) BatchGet(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]app.Entitlements, error) {
	out := make(map[uuid.UUID]app.Entitlements, len(ids))
	for start := 0; start < len(ids); start += 500 {
		chunk := ids[start:min(start+500, len(ids))]
		raw := make([]string, len(chunk))
		for i, id := range chunk {
			raw[i] = id.String()
		}
		res, err := b.client.BatchGetWorkspaceEntitlements(ctx, &pb.BatchGetWorkspaceEntitlementsRequest{WorkspaceIds: raw})
		if err != nil {
			return nil, err
		}
		for k, v := range res.GetEntitlements() {
			id, err := uuid.Parse(k)
			if err != nil {
				return nil, err
			}
			out[id] = toEntitlements(v)
		}
	}
	return out, nil
}

func (b *Billing) GetPayer(ctx context.Context, workspaceID uuid.UUID) (*string, bool, error) {
	res, err := b.client.GetWorkspacePayer(ctx, &pb.GetWorkspacePayerRequest{WorkspaceId: workspaceID.String()})
	if err != nil {
		return nil, false, err
	}
	return res.PayerId, res.GetRequiresPayer(), nil
}

func (b *Billing) ListPlans(ctx context.Context) ([]app.PlanEntitlements, error) {
	res, err := b.client.ListPlans(ctx, &pb.ListPlansRequest{})
	if err != nil {
		return nil, err
	}
	out := make([]app.PlanEntitlements, 0, len(res.GetPlans()))
	for _, p := range res.GetPlans() {
		out = append(out, app.PlanEntitlements{Tier: p.GetTier(), Features: p.GetFeatures(), Limits: p.GetLimits()})
	}
	return out, nil
}
```

`var ProvideBilling = NewBilling`, `var _ app.EntitlementSvc = (*Billing)(nil)`; add to the service `ProviderSet` with `wire.Bind(new(app.EntitlementSvc), new(*Billing))`. Health (`controller/health/health.go`): add a `billingService` readiness check on `servicesCfg.Billing.LiveURL` like `authorizationService`.

- [ ] **Step 4: Build, commit**

```bash
nx run note:gen:wire && nx build note && nx lint note
git add internal/note cmd/note/.env cmd/note/wire_gen.go
git commit -m "feat(note): billing entitlement client and payment-required errors"
```

---

### Task 2: Gates — members limit, payer guard, graph

**Files:**
- Modify: `internal/note/app/cmd_updateworkspacemembers.go`, `internal/note/app/cmd_leaveworkspace.go`, `internal/note/app/q_getnotegraph.go`, `internal/note/app/q_getworkspacegraph.go`
- Create: `internal/note/app/billinggate.go`, `internal/note/app/billinggate_test.go`

**Interfaces:**
- Consumes: `EntitlementSvc`, `AuthorizationSvc.GetWorkspaceMembers(ctx, userID, workspaceID) ([]WorkspaceMember, error)` (existing; `WorkspaceMember{ID string; Role WorkspaceRole}` — check field names in `internal/note/app/model.go`).
- Produces: `checkMembersLimit(ctx, ent EntitlementSvc, workspaceID uuid.UUID, current, requested int) error`, `checkPayerKept(ctx, ent EntitlementSvc, workspaceID uuid.UUID, members []WorkspaceMemberUpdate) error`, `checkPayerMayLeave(ctx, ent EntitlementSvc, workspaceID uuid.UUID, userID string) error`.

- [ ] **Step 1: Write the failing test**

```go
// internal/note/app/billinggate_test.go
package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type fakeEntitlements struct {
	EntitlementSvc
	ent           Entitlements
	payer         *string
	requiresPayer bool
}

func (f fakeEntitlements) Get(context.Context, uuid.UUID) (Entitlements, error) { return f.ent, nil }
func (f fakeEntitlements) GetPayer(context.Context, uuid.UUID) (*string, bool, error) {
	return f.payer, f.requiresPayer, nil
}
func (f fakeEntitlements) ListPlans(context.Context) ([]PlanEntitlements, error) {
	return []PlanEntitlements{
		{Tier: "plus", Limits: map[string]int64{LimitMembers: 10}},
		{Tier: "pro", Limits: map[string]int64{LimitMembers: Unlimited}},
	}, nil
}

func TestMembersLimit(t *testing.T) {
	free := fakeEntitlements{ent: Entitlements{Plan: "free", Limits: map[string]int64{LimitMembers: 3}}}
	ws := uuid.New()

	var pr *errs.PaymentRequired
	err := checkMembersLimit(context.Background(), free, ws, 3, 4)
	if !errors.As(err, &pr) || pr.RequiredPlan != "plus" || *pr.Max != 3 {
		t.Fatalf("growing past limit: err = %v", err)
	}
	if err := checkMembersLimit(context.Background(), free, ws, 5, 4); err != nil {
		t.Fatalf("shrinking an over-limit workspace must pass: %v", err)
	}
	if err := checkMembersLimit(context.Background(), free, ws, 5, 5); err != nil {
		t.Fatalf("same count (role change) must pass: %v", err)
	}
}

func TestPayerGuard(t *testing.T) {
	payer := "112"
	ws := uuid.New()
	g := fakeEntitlements{payer: &payer, requiresPayer: true}

	demoted := []WorkspaceMemberUpdate{{ID: "112", Role: WorkspaceRoleEditor}, {ID: "111", Role: WorkspaceRoleOwner}}
	if err := checkPayerKept(context.Background(), g, ws, demoted); !isCode(err, errs.CodeBillingPayerMustTransfer) {
		t.Fatalf("demoting payer: err = %v", err)
	}
	if err := checkPayerMayLeave(context.Background(), g, ws, "112"); !isCode(err, errs.CodeBillingPayerMustTransfer) {
		t.Fatalf("payer leaving: err = %v", err)
	}
	lifetime := fakeEntitlements{payer: &payer, requiresPayer: false}
	if err := checkPayerMayLeave(context.Background(), lifetime, ws, "112"); err != nil {
		t.Fatalf("no recurring charge: %v", err)
	}
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

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

func checkMembersLimit(ctx context.Context, ent EntitlementSvc, workspaceID uuid.UUID, current, requested int) error {
	if requested <= current {
		return nil
	}
	e, err := ent.Get(ctx, workspaceID)
	if err != nil {
		return err
	}
	limit := e.Limit(LimitMembers)
	if limit == Unlimited || int64(requested) <= limit {
		return nil
	}
	plans, err := ent.ListPlans(ctx)
	if err != nil {
		return err
	}
	return errs.NewPlanLimitExceeded(LimitMembers, int64(requested), limit, RequiredPlanForLimit(plans, LimitMembers, int64(requested)))
}

func checkPayerKept(ctx context.Context, ent EntitlementSvc, workspaceID uuid.UUID, members []WorkspaceMemberUpdate) error {
	payer, requires, err := ent.GetPayer(ctx, workspaceID)
	if err != nil || payer == nil || !requires {
		return err
	}
	for _, m := range members {
		if m.ID == *payer && m.Role == WorkspaceRoleOwner {
			return nil
		}
	}
	return errs.NewBillingPayerMustTransfer(workspaceID)
}

func checkPayerMayLeave(ctx context.Context, ent EntitlementSvc, workspaceID uuid.UUID, userID string) error {
	payer, requires, err := ent.GetPayer(ctx, workspaceID)
	if err != nil {
		return err
	}
	if requires && payer != nil && *payer == userID {
		return errs.NewBillingPayerMustTransfer(workspaceID)
	}
	return nil
}
```

Wire the gates (inject `entitlementSvc EntitlementSvc` into each handler's constructor):

- `UpdateWorkspaceMembersHandler.Handle`, after the existing owner validation:

```go
	current, err := h.authorizationSvc.GetWorkspaceMembers(ctx, cmd.UserID, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	if err := checkMembersLimit(ctx, h.entitlementSvc, cmd.WorkspaceID, len(current), len(cmd.Members)); err != nil {
		return err
	}
	if err := checkPayerKept(ctx, h.entitlementSvc, cmd.WorkspaceID, cmd.Members); err != nil {
		return err
	}
```

- `LeaveWorkspaceHandler.Handle`, before `LeaveWorkspace`: `if err := checkPayerMayLeave(ctx, h.entitlementSvc, cmd.WorkspaceID, cmd.UserID); err != nil { return err }`.
- `GetNoteGraphHandler.Handle`, after the permission check and **before** `if query.Depth <= 0`:

```go
	ent, err := h.entitlementSvc.Get(ctx, workspaceID)
	if err != nil {
		return Graph{}, err
	}
	if maxDepth := ent.Limit(LimitNoteGraphMaxDepth); maxDepth != Unlimited && (query.Depth <= 0 || int64(query.Depth) > maxDepth) {
		query.Depth = int(maxDepth)
	}
```

- `GetWorkspaceGraphHandler.Handle`, after the permission check:

```go
	ent, err := h.entitlementSvc.Get(ctx, query.ID)
	if err != nil {
		return Graph{}, err
	}
	if !ent.Feature(FeatureWorkspaceGraph) {
		plans, err := h.entitlementSvc.ListPlans(ctx)
		if err != nil {
			return Graph{}, err
		}
		return Graph{}, errs.NewFeatureNotInPlan(FeatureWorkspaceGraph, RequiredPlanForFeature(plans, FeatureWorkspaceGraph))
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

### Task 3: Trash purge cron

**Files:**
- Create: `internal/note/app/trashdelete.go` (shared delete helper), `internal/note/app/cmd_purgeexpiredtrash.go`
- Modify: `internal/note/app/cmd_permanentlydeleteworkspaceitems.go`, `internal/note/app/cmd_emptytrash.go` (use the helper)
- Modify: `internal/note/infra/persistence/pgsqlc/workspace_read.sql`, `internal/note/app/readmodel.go`, create `internal/note/infra/persistence/pgreadmodel/trashworkspaces.go`, modify `internal/note/infra/persistence/wire.go`
- Create: `internal/note/controller/cron/{cron.go,wire.go}`; modify `internal/note/controller/wire.go`, `internal/note/server.go`, `internal/note/app/app.go`, `wire.go`

**Interfaces:**
- Consumes: `pkg/cron`, `EntitlementSvc.BatchGet`, `commonconfig.Identity.ServiceAccountUserID`.
- Produces: `permanentlyDeleteItems(ctx, r domain.RepoRegistry, noteIDs, folderIDs []uuid.UUID, userID string) error`; `Cmds.PurgeExpiredTrash`; `TrashWorkspacesReadModel.Handle(ctx, before time.Time) ([]uuid.UUID, error)`; job `note:purge-expired-trash`.

- [ ] **Step 1: Extract the delete helper**

Move the body written in plan 01 Task 4 (notes incl. folder descendants first, then folders) into:

```go
// internal/note/app/trashdelete.go
package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

// permanentlyDeleteItems deletes notes (and every note under the folders) before the folders,
// so each removed note emits NoteDeletedEvent instead of disappearing through the SQL cascade.
func permanentlyDeleteItems(ctx context.Context, r domain.RepoRegistry, noteIDs, folderIDs []uuid.UUID, userID string) error {
	folderRepo, noteRepo := r.Folder(), r.Note()
	notesByID := map[uuid.UUID]*domain.Note{}
	if len(noteIDs) > 0 {
		notes, err := noteRepo.GetMany(ctx,
			//exhaustruct:ignore
			&domain.NoteRepoGetManyParams{IDs: noteIDs, ForUpdate: true})
		if err != nil {
			return err
		}
		for _, n := range notes {
			notesByID[n.ID()] = n
		}
	}
	var folders []*domain.Folder
	if len(folderIDs) > 0 {
		var err error
		folders, err = folderRepo.GetMany(ctx,
			//exhaustruct:ignore
			&domain.FolderRepoGetManyParams{IDs: folderIDs, ForUpdate: true})
		if err != nil {
			return err
		}
		for _, f := range folders {
			children, err := noteRepo.GetRecursiveChildrenFromFolder(ctx, f.ID(), true)
			if err != nil {
				return err
			}
			for _, n := range children {
				notesByID[n.ID()] = n
			}
		}
	}
	notes := make([]*domain.Note, 0, len(notesByID))
	for _, n := range notesByID {
		n.PermanentlyDelete(userID)
		notes = append(notes, n)
	}
	if err := noteRepo.SaveMany(ctx, notes); err != nil {
		return err
	}
	for _, f := range folders {
		f.PermanentlyDelete(userID)
	}
	return folderRepo.SaveMany(ctx, folders)
}
```

`PermanentlyDeleteWorkspaceItemsHandler.Handle`: keep the permission check, then
`return h.uow.Execute(ctx, func(r domain.RepoRegistry) error { return permanentlyDeleteItems(ctx, r, cmd.NoteIDs, cmd.FolderIDs, cmd.UserID) })`.
`EmptyTrashHandler.Handle`: load trashed folder ids and trashed note ids (existing `GetMany` calls with `TrashOnly`) and call the helper with them.

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
package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/domain"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type PurgeExpiredTrash struct{}

type PurgeExpiredTrashHandler struct {
	uow            domain.UnitOfWork
	readModel      TrashWorkspacesReadModel
	entitlementSvc EntitlementSvc
	identity       *commonconfig.Identity
}

func NewPurgeExpiredTrashHandler(uow domain.UnitOfWork, readModel TrashWorkspacesReadModel, entitlementSvc EntitlementSvc, identity *commonconfig.Identity) *PurgeExpiredTrashHandler {
	return &PurgeExpiredTrashHandler{uow: uow, readModel: readModel, entitlementSvc: entitlementSvc, identity: identity}
}

var ProvidePurgeExpiredTrashHandler = NewPurgeExpiredTrashHandler

type PurgeExpiredTrashCmd commonhandler.Cmd[PurgeExpiredTrash]

var _ PurgeExpiredTrashCmd = (*PurgeExpiredTrashHandler)(nil)

// No permission check: cron-only, runs as the note service account.
func (h *PurgeExpiredTrashHandler) Handle(ctx context.Context, _ *PurgeExpiredTrash) error {
	now := time.Now()
	workspaceIDs, err := h.readModel.Handle(ctx, now.Add(-24*time.Hour))
	if err != nil || len(workspaceIDs) == 0 {
		return err
	}
	ents, err := h.entitlementSvc.BatchGet(ctx, workspaceIDs)
	if err != nil {
		return err
	}
	for _, ws := range workspaceIDs {
		days := ents[ws].Limit(LimitTrashRetentionDays)
		if days == Unlimited {
			continue
		}
		cutoff := now.Add(-time.Duration(days) * 24 * time.Hour)
		if err := h.purgeWorkspace(ctx, ws, cutoff); err != nil {
			slog.WarnContext(ctx, "trash purge failed", slog.String("workspace_id", ws.String()), slog.Any("error", err))
		}
	}
	return nil
}

func (h *PurgeExpiredTrashHandler) purgeWorkspace(ctx context.Context, ws uuid.UUID, cutoff time.Time) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		folders, err := r.Folder().GetMany(ctx, &domain.FolderRepoGetManyParams{
			WorkspaceID: ws, TrashOnly: true, TrashedBy: domain.TrashedByPurpose, ForUpdate: false,
		})
		if err != nil {
			return err
		}
		notes, err := r.Note().GetMany(ctx, &domain.NoteRepoGetManyParams{
			WorkspaceID: ws, TrashOnly: true, TrashedBy: domain.TrashedByPurpose, ForUpdate: false,
		})
		if err != nil {
			return err
		}
		var folderIDs, noteIDs []uuid.UUID
		for _, f := range folders {
			if f.TrashedAt().Before(cutoff) {
				folderIDs = append(folderIDs, f.ID())
			}
		}
		for _, n := range notes {
			if n.TrashedAt().Before(cutoff) {
				noteIDs = append(noteIDs, n.ID())
			}
		}
		if len(folderIDs) == 0 && len(noteIDs) == 0 {
			return nil
		}
		return permanentlyDeleteItems(ctx, r, noteIDs, folderIDs, h.identity.ServiceAccountUserID)
	})
}
```

Register in `app.go` (`Cmds.PurgeExpiredTrash`) and `wire.go`.

- [ ] **Step 4: Cron controller**

```go
// internal/note/controller/cron/cron.go
package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/workspaceevent"
	pkgcron "github.com/notopia-uit/notopia/pkg/cron"
)

// NOTE: move this controller to cmd/note-worker when crons are split out of the API process.
type Cron struct {
	scheduler *pkgcron.Scheduler
}

func New(a *app.Server, redis *workspaceevent.RedisClient, logger *slog.Logger) (*Cron, error) {
	locker, err := pkgcron.NewRedisLocker(redis.Client)
	if err != nil {
		return nil, err
	}
	s, err := pkgcron.NewScheduler(locker, logger)
	if err != nil {
		return nil, err
	}
	if err := s.Register(pkgcron.Job{
		Name: "note:purge-expired-trash", Schedule: "0 3 * * *", Timeout: 30 * time.Minute,
		Run: func(ctx context.Context) error { return a.Cmds.PurgeExpiredTrash.Handle(ctx, &app.PurgeExpiredTrash{}) },
	}); err != nil {
		return nil, err
	}
	return &Cron{scheduler: s}, nil
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

Manual: as admin set free `trash_retention_days` to `0` (`PATCH /billing/admin/plans/free`), trash a note in a free workspace, backdate it (`UPDATE notes SET trashed_at = now() - interval '2 days' WHERE id = …`), temporarily schedule the job `* * * * *` → the note is gone, `note.deleted` published, document row and Meilisearch entry removed (plan 01 consumers). Restore the schedule and the plan limit.

```bash
git add internal/note cmd/note/wire_gen.go
git commit -m "feat(note): plan-aware trash purge cron"
```

---

### Task 4: Billing change notifications over SSE

**Files:**
- Create: `api/note/components/schemas/WorkspaceBillingChangedEvent.yaml`
- Modify: `api/note/paths/workspaces_{workspaceId}_events.yaml`
- Modify: `internal/note/app/workspaceevent.go`, `internal/note/infra/workspaceevent/workspaceevent.go`
- Create: `internal/note/app/e_billingchanged.go`, `internal/note/controller/event/billingchanged.go`
- Modify: `internal/note/controller/event/event.go`, `app.go`, `wire.go`

**Interfaces:**
- Consumes: `share.SubscriptionChangedEvent`, `share.PlanChangedEvent` (plan 02 Task 5b) on `events.integration.billing.subscription.changed` / `events.integration.billing.plan.changed`.
- Produces: SSE event `WorkspaceBillingChangedEvent {id, event, data: {workspaceId?, plan?, tier?}}`; hub broadcast key `workspace:broadcast` for `uuid.Nil`.

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
    description: Refetch /billing/workspaces/{id}/entitlements. workspaceId is absent for plan-wide changes.
    properties:
      workspaceId:
        $ref: ./Workspace.yaml#/properties/id
      plan:
        type: string
        example: plus
      tier:
        type: string
        example: free

required:
  - id
  - event
  - data
```

Add it to the `oneOf` and the discriminator `mapping` in `workspaces_{workspaceId}_events.yaml`. Run `nx lint api && nx run api:gen`.

- [ ] **Step 2: Event type + hub broadcast**

`app/workspaceevent.go`:

```go
type WorkspaceEventBillingChanged struct {
	workspaceEvent[note.WorkspaceBillingChangedEventEvent]
}
```

and a case in `NewEmptyWorkspaceEventFromType` for `note.WorkspaceBillingChangedEventEventWorkspaceBillingChangedEvent`.

`infra/workspaceevent/workspaceevent.go`:
- `setupRouting`: `internalTopic := fmt.Sprintf("workspace:%s", workspaceID)`; if `workspaceID == uuid.Nil.String()` use `"workspace:broadcast"`.
- `Subscribe`: subscribe to both `workspace:{id}` and `workspace:broadcast`; run the existing loop for each channel in its own goroutine, close `eventCh` after both exit (`sync.WaitGroup`).

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
	WorkspaceID *uuid.UUID // nil = plan-wide change, broadcast
	Plan        string
	Tier        string
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
	target := uuid.Nil
	data := note.WorkspaceBillingChangedEventData{}
	if ev.WorkspaceID != nil {
		target = *ev.WorkspaceID
		data.WorkspaceId = ev.WorkspaceID
	}
	if ev.Plan != "" {
		data.Plan = &ev.Plan
	}
	if ev.Tier != "" {
		data.Tier = &ev.Tier
	}
	return h.publisher.Publish(ctx, target, h.identity.ServiceAccountUserID, &WorkspaceEventBillingChanged{
		workspaceEvent[note.WorkspaceBillingChangedEventEvent]{
			Id: uuid.Must(uuid.NewV7()), Event: note.WorkspaceBillingChangedEventEventWorkspaceBillingChangedEvent, Data: data,
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

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/pkg/api/share"
)

func (e *Event) subscriptionChangedHandler(msg *message.Message) error {
	var ev share.SubscriptionChangedEvent
	if err := json.Unmarshal(msg.Payload, &ev); err != nil {
		return fmt.Errorf("failed to unmarshal SubscriptionChangedEvent: %w", err)
	}
	ws := ev.WorkspaceId
	return e.app.Events.BillingChanged.Handle(msg.Context(), &app.BillingChanged{WorkspaceID: &ws, Plan: string(ev.Plan)})
}

func (e *Event) planChangedHandler(msg *message.Message) error {
	var ev share.PlanChangedEvent
	if err := json.Unmarshal(msg.Payload, &ev); err != nil {
		return fmt.Errorf("failed to unmarshal PlanChangedEvent: %w", err)
	}
	return e.app.Events.BillingChanged.Handle(msg.Context(), &app.BillingChanged{Tier: string(ev.Tier)})
}
```

`event.go` `setup()`:

```go
	e.router.AddConsumerHandler("BillingSubscriptionChangedHandler", "events.integration.billing.subscription.changed", e.subcriber, e.subscriptionChangedHandler)
	e.router.AddConsumerHandler("BillingPlanChangedHandler", "events.integration.billing.plan.changed", e.subcriber, e.planChangedHandler)
```

- [ ] **Step 4: Verify, commit**

```bash
nx run note:gen:wire && nx build note && nx lint note
```

Manual: open the workspace SSE stream (`curl -N -H 'Authorization: Bearer …' http://api.notopia.localhost/note/workspaces/W/events`), complete a checkout for W → a `WorkspaceBillingChangedEvent` arrives with `workspaceId: W`. Admin plan edit → the event arrives without `workspaceId`.

```bash
git add api/note pkg/api/note packages/api-gen internal/note cmd/note/wire_gen.go
git commit -m "feat(note): forward billing changes to workspace SSE"
```

---

### Task 5: Plan notes in the note OpenAPI

**Files:**
- Modify: `api/note/paths/workspaces_{workspaceId}_members.yaml`, `workspaces_{workspaceId}_leave.yaml`, `notes_{noteId}_graph.yaml`, `workspaces_{workspaceId}_graph.yaml`, `workspaces_{workspaceId}_show-trash.yaml`

- [ ] **Step 1: Descriptions + responses**

Append to each operation's `description` (add `description` if missing):

| File | Line |
| --- | --- |
| members (PUT/PATCH) | "Plan limit `members` (free 3, plus 10, pro unlimited). Adding beyond the limit returns 402. The billing payer cannot lose the owner role while a recurring subscription is active (409 `billingPayerMustTransfer`)." |
| leave | "The billing payer cannot leave while a recurring subscription is active (409 `billingPayerMustTransfer`)." |
| note graph | "`depth` is capped by plan limit `note_graph_max_depth` (free 5)." |
| workspace graph | "Requires feature `workspace_graph` (plus, pro); free returns 402." |
| show-trash | "Trashed items older than the plan's `trash_retention_days` are purged daily." |

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

### Task 6: Billing gRPC client in document

**Files:**
- Create: `apps/document/src/entitlement/{entitlement.module.ts,entitlement.service.ts,index.ts}`
- Modify: `apps/document/src/config/config.ts`, `config.factory.ts` (`billingUrl`), `apps/document/.env`, `apps/document/src/app.module.ts`
- Modify: `packages/pb/package.json` (export `./billing` if `nx run proto:gen` did not add it — mirror `./note`)

**Interfaces:**
- Produces:

```ts
export interface WorkspaceEntitlements {
  plan: string;
  features: Record<string, boolean>;
  limits: Record<string, number>;
}
class EntitlementService {
  get(workspaceId: string): Promise<WorkspaceEntitlements>;
  batchGet(workspaceIds: string[]): Promise<Map<string, WorkspaceEntitlements>>; // chunks of 500
  listPlans(): Promise<{ tier: string; limits: Record<string, number> }[]>;
}
export const feature = (e: WorkspaceEntitlements, key: string) => e.features[key] ?? false;
export const limit = (e: WorkspaceEntitlements, key: string) => e.limits[key] ?? 0;
export const UNLIMITED = -1;
```

- [ ] **Step 1: Config**

`servicesConfigSchema`: add `billingUrl: z.string()`; factory: `billingUrl: process.env.NOTOPIA_DOCUMENT_SERVICES_BILLING_GRPC_URL!`; `.env`: `NOTOPIA_DOCUMENT_SERVICES_BILLING_GRPC_URL=localhost:18083`.

- [ ] **Step 2: Module + service**

`entitlement.module.ts`: copy `authorization/authorization.module.ts`, replacing the package/service constants with `BILLING_PACKAGE_NAME`, `BILLING_SERVICE_NAME`, `BillingServiceService` from `@notopia-uit/pb/billing`, url `servicesCfg.billingUrl`, provider/export `EntitlementService`.

```ts
// apps/document/src/entitlement/entitlement.service.ts
import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { type ClientGrpc } from '@nestjs/microservices';
import {
  BILLING_PACKAGE_NAME,
  BILLING_SERVICE_NAME,
  type BillingServiceClient,
  type Entitlements,
} from '@notopia-uit/pb/billing';
import { Traceable } from 'nestjs-otel';
import { firstValueFrom } from 'rxjs';

export interface WorkspaceEntitlements {
  plan: string;
  features: Record<string, boolean>;
  limits: Record<string, number>;
}

export const UNLIMITED = -1;
export const feature = (e: WorkspaceEntitlements, key: string) => e.features[key] ?? false;
export const limit = (e: WorkspaceEntitlements, key: string) => e.limits[key] ?? 0;

const toEntitlements = (e?: Entitlements): WorkspaceEntitlements => ({
  plan: e?.plan ?? 'free',
  features: e?.features ?? {},
  limits: Object.fromEntries(Object.entries(e?.limits ?? {}).map(([k, v]) => [k, Number(v)])),
});

@Injectable()
@Traceable()
export class EntitlementService implements OnModuleInit {
  private client!: BillingServiceClient;

  constructor(@Inject(BILLING_PACKAGE_NAME) private readonly grpc: ClientGrpc) {}

  onModuleInit(): void {
    this.client = this.grpc.getService<BillingServiceClient>(BILLING_SERVICE_NAME);
  }

  async get(workspaceId: string): Promise<WorkspaceEntitlements> {
    const res = await firstValueFrom(this.client.getWorkspaceEntitlements({ workspaceId }));
    return toEntitlements(res.entitlements);
  }

  async batchGet(workspaceIds: string[]): Promise<Map<string, WorkspaceEntitlements>> {
    const out = new Map<string, WorkspaceEntitlements>();
    for (let i = 0; i < workspaceIds.length; i += 500) {
      const res = await firstValueFrom(
        this.client.batchGetWorkspaceEntitlements({ workspaceIds: workspaceIds.slice(i, i + 500) })
      );
      for (const [id, e] of Object.entries(res.entitlements)) out.set(id, toEntitlements(e));
    }
    return out;
  }

  async listPlans(): Promise<{ tier: string; limits: Record<string, number> }[]> {
    const res = await firstValueFrom(this.client.listPlans({}));
    return res.plans.map((p) => ({
      tier: p.tier,
      limits: Object.fromEntries(Object.entries(p.limits).map(([k, v]) => [k, Number(v)])),
    }));
  }
}
```

(int64 maps may be generated as `number` or `Long`; `Number(v)` covers both.) Import `EntitlementModule` in `app.module.ts`.

- [ ] **Step 3: Build, commit**

```bash
oxfmt apps/document/src && nx lint document --fix && nx build document
git add apps/document packages/pb/package.json
git commit -m "feat(document): billing entitlement grpc client"
```

---

### Task 7: Revision metadata, contributors and `commitIfDue`

**Files:**
- Modify: `apps/document/src/document/document.entity.ts`, `apps/document/src/revision/revision.entity.ts`
- Modify: `apps/document/src/document/document.service.ts` (`commitIfDue`, `upsertData` with contributors, manual commit through `commitIfDue`)
- Create: `apps/document/src/hocuspocus/contributor-tracker.ts`
- Modify: `apps/document/src/hocuspocus/hocuspocus.ts` (track contributors, pass them to `upsertData`)
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
  force: boolean;             // manual / last client left / sweeper
  minIntervalSeconds: number; // from entitlements, 0 for manual
}): Promise<{ revisionId: string | null }>;
DocumentService.upsertData(id: string, data: Buffer, modified: boolean, contributors: string[]): Promise<void>;
ContributorTracker.add(documentId: string, userId: string): void;
ContributorTracker.drain(documentId: string): string[]; // most recent last
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

- [ ] **Step 2: Contributor tracker**

```ts
// apps/document/src/hocuspocus/contributor-tracker.ts
import { Injectable } from '@nestjs/common';

@Injectable()
export class ContributorTracker {
  private readonly pending = new Map<string, string[]>();

  add(documentId: string, userId: string): void {
    const list = (this.pending.get(documentId) ?? []).filter((id) => id !== userId);
    list.push(userId);
    this.pending.set(documentId, list);
  }

  drain(documentId: string): string[] {
    const list = this.pending.get(documentId) ?? [];
    this.pending.delete(documentId);
    return list;
  }
}
```

Register in `HocuspocusModule` providers. In `hocuspocus.ts` `onChange`: when `data.connection` exists, `this.contributors.add(data.documentName, data.context.user.id)`. In the store callback: `await documentService.upsertData(id, state, readModified(document), this.contributors.drain(id));`.

- [ ] **Step 3: Write the failing test**

```ts
// apps/document/src/document/commit-policy.spec.ts
import { contentHash, mergeContributors, shouldCommit } from './commit-policy';

describe('commit policy', () => {
  const now = new Date('2026-10-11T10:00:00Z');

  it('respects the min interval unless forced', () => {
    const last = new Date('2026-10-11T09:58:00Z'); // 120 s ago
    expect(shouldCommit({ modified: true, lastCommittedAt: last, now, force: false, minIntervalSeconds: 300 })).toBe(false);
    expect(shouldCommit({ modified: true, lastCommittedAt: last, now, force: true, minIntervalSeconds: 300 })).toBe(true);
    expect(shouldCommit({ modified: true, lastCommittedAt: last, now, force: false, minIntervalSeconds: 60 })).toBe(true);
    expect(shouldCommit({ modified: false, lastCommittedAt: null, now, force: true, minIntervalSeconds: 0 })).toBe(false);
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

`upsertData` merges contributors in SQL so concurrent instances never lose ids:

```ts
  async upsertData(id: string, data: Buffer, modified: boolean, contributors: string[]): Promise<void> {
    await this.repo.query(
      `INSERT INTO documents (id, data, modified, "pendingContributorIds", "updatedAt")
       VALUES ($1, $2, $3, $4::text[], now())
       ON CONFLICT (id) DO UPDATE SET
         data = EXCLUDED.data,
         modified = EXCLUDED.modified,
         "pendingContributorIds" = ARRAY(
             SELECT x FROM unnest(documents."pendingContributorIds") AS x
             WHERE x <> ALL(EXCLUDED."pendingContributorIds")
           ) || EXCLUDED."pendingContributorIds",
         "updatedAt" = now()`,
      [id, data, modified, contributors]
    );
  }
```

The merge keeps order (older contributors first, the latest batch last) and never drops ids written by another instance. Use the column names TypeORM generated (check `\d documents`; default naming keeps camelCase quoted).

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

### Task 8: Auto-commit extension

**Files:**
- Create: `apps/document/src/hocuspocus/auto-commit.extension.ts`
- Modify: `apps/document/src/hocuspocus/hocuspocus.ts` (extensions, debounce config), `apps/document/src/config/*` (`hocuspocusDebounceMs`, `hocuspocusMaxDebounceMs`), `apps/document/src/hocuspocus/hocuspocus.module.ts` (import `EntitlementModule`)

**Interfaces:**
- Consumes: `EntitlementService.get`, `NoteService.getWorkspaceByNote({ noteId, userId })`, `DocumentService.commitIfDue`, `HocuspocusService.setModified`.
- Produces: `AutoCommitExtension implements Extension` (`afterStoreDocument`).

- [ ] **Step 1: Extension**

```ts
// apps/document/src/hocuspocus/auto-commit.extension.ts
import type { Extension, afterStoreDocumentPayload } from '@hocuspocus/server';
import { Logger } from '@nestjs/common';

import { DocumentService } from '../document/document.service';
import { EntitlementService, feature, limit } from '../entitlement/entitlement.service';
import { NoteService } from '../note/note.service';
import { readModified } from './document-metadata';
import type { HocuspocusContext } from './hocuspocus-context';

export class AutoCommitExtension implements Extension {
  private readonly logger = new Logger(AutoCommitExtension.name);

  constructor(
    private readonly documentService: DocumentService,
    private readonly noteService: NoteService,
    private readonly entitlementService: EntitlementService,
    private readonly onCommitted: (documentId: string) => void
  ) {}

  async afterStoreDocument({ documentName, document, clientsCount, lastContext }: afterStoreDocumentPayload<HocuspocusContext>) {
    if (!readModified(document)) return;
    const userId = lastContext?.user?.id;
    try {
      const entity = await this.documentService.getById(documentName);
      if (!entity) return;
      const force = clientsCount === 0;
      const authorId = entity.pendingContributorIds.at(-1) ?? userId;
      if (!authorId) return;
      const workspace = await this.noteService.getWorkspaceByNote({ noteId: documentName, userId: authorId });
      const ent = await this.entitlementService.get(workspace.id);
      if (!feature(ent, 'auto_commit')) return;
      const { revisionId } = await this.documentService.commitIfDue({
        documentId: documentName,
        trigger: 'auto',
        force,
        minIntervalSeconds: limit(ent, 'auto_commit_min_interval_seconds'),
      });
      if (revisionId) this.onCommitted(documentName);
    } catch (err) {
      this.logger.warn({ err, documentId: documentName }, 'auto-commit skipped');
    }
  }
}
```

Never rethrow: an error in a store hook keeps the document in memory.

- [ ] **Step 2: Register + debounce config**

Config: `hocuspocusDebounceMs: z.number().default(2000)`, `hocuspocusMaxDebounceMs: z.number().default(10000)` from `NOTOPIA_DOCUMENT_HOCUSPOCUS_DEBOUNCE_MS` / `..._MAX_DEBOUNCE_MS`.

`hocuspocus.ts` constructor: inject `EntitlementService` and `ConfigService`; pass `debounce`/`maxDebounce`; add `new AutoCommitExtension(documentService, noteService, entitlementService, (id) => this.setModifiedFalse(id))` **after** the `HocuspocusDatabase` extension, where `setModifiedFalse` sets the metadata map to `{ modified: false }` on the loaded document with a normal origin (same as `HocuspocusService.setModified`; avoid a circular DI by implementing it inline on `this.hocuspocus.documents.get(id)`).

- [ ] **Step 3: Verify, commit**

```bash
oxfmt apps/document/src && nx lint document --fix && nx build document
```

Manual: workspace on `plus` (300 s interval): edit a note, wait 5 min while typing occasionally → an `auto` revision appears; close the tab → a revision is forced immediately; free workspace: no auto revisions, manual Commit still works.

```bash
git add apps/document
git commit -m "feat(document): auto-commit on paid plans via hocuspocus extension"
```

---

### Task 9: Sweeper + revision purge crons, scaling toggle

**Files:**
- Create: `apps/document/src/cron/{cron.module.ts,sweeper.cron.ts,revision-purge.cron.ts,revision-purge.ts,revision-purge.spec.ts}`
- Modify: `apps/document/src/app.module.ts` (`ScheduleModule.forRoot()`, `LocksSetupModule.forRoot(...)`, `CronModule`), `apps/document/src/config/*` (`scalingEnabled`, `redisUrl`), `apps/document/src/hocuspocus/hocuspocus.ts` (Redis extension), `apps/document/package.json` (deps), `deploy/compose/services/document.yaml` (optional redis)

**Interfaces:**
- Consumes: `@notopia-uit/lib-server/locks` (`LocksSetupModule`), `NoteService.batchGetNoteWorkspaces`, `EntitlementService.batchGet/listPlans`.
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
    const retention = new Map([['a', 7], ['b', 7]]);
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
    if (days === undefined || days < 0) continue;
    const cutoff = now.getTime() - days * 86_400_000;
    const sorted = [...list].sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
    for (const r of sorted.slice(1)) if (r.createdAt.getTime() < cutoff) out.push(r.id);
  }
  return out;
};
```

The cron loads, for documents having revisions older than `minDays`, **all** their revisions' `(id, documentId, createdAt)` so "latest" is correct.

- [ ] **Step 3: Crons**

```ts
// apps/document/src/cron/revision-purge.cron.ts
import { Injectable, Logger } from '@nestjs/common';
import { OnOneInstance, WithoutOverlapping } from '@nestjs/locks';
import { Cron } from '@nestjs/schedule';
import { InjectDataSource } from '@nestjs/typeorm';
import { DataSource, In } from 'typeorm';

import { EntitlementService, limit } from '../entitlement/entitlement.service';
import { NoteService } from '../note/note.service';
import { RevisionEntity } from '../revision/revision.entity';
import { planRevisionDeletes } from './revision-purge';

// NOTE: move to a document worker when crons are split out of the API process.
@Injectable()
export class RevisionPurgeCron {
  private readonly logger = new Logger(RevisionPurgeCron.name);

  constructor(
    @InjectDataSource() private readonly dataSource: DataSource,
    private readonly noteService: NoteService,
    private readonly entitlementService: EntitlementService
  ) {}

  @Cron('30 3 * * *', { name: 'document-revision-purge', timeZone: 'Asia/Ho_Chi_Minh' })
  @OnOneInstance({ key: 'document:revision-purge' })
  @WithoutOverlapping()
  async purge(): Promise<void> {
    const plans = await this.entitlementService.listPlans();
    const days = plans.map((p) => limit({ plan: p.tier, features: {}, limits: p.limits }, 'revision_history_days')).filter((d) => d >= 0);
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
      const ents = await this.entitlementService.batchGet([...new Set(workspaces.values())]);
      const retention = new Map<string, number>();
      for (const [noteId, ws] of workspaces) {
        const e = ents.get(ws);
        if (e) retention.set(noteId, limit(e, 'revision_history_days'));
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

(`documentId` is the FK column TypeORM created for `RevisionEntity.document`; check the actual column name.) Revisions are hard-deleted (`delete`, not `softDelete`).

```ts
// apps/document/src/cron/sweeper.cron.ts
import { Injectable, Logger } from '@nestjs/common';
import { OnOneInstance, WithoutOverlapping } from '@nestjs/locks';
import { Cron } from '@nestjs/schedule';
import { InjectRepository } from '@nestjs/typeorm';
import { LessThan, Repository } from 'typeorm';

import { DocumentEntity } from '../document/document.entity';
import { DocumentService } from '../document/document.service';
import { EntitlementService, feature } from '../entitlement/entitlement.service';
import { NoteService } from '../note/note.service';

// NOTE: move to a document worker when crons are split out of the API process.
@Injectable()
export class SweeperCron {
  private readonly logger = new Logger(SweeperCron.name);

  constructor(
    @InjectRepository(DocumentEntity) private readonly repo: Repository<DocumentEntity>,
    private readonly documentService: DocumentService,
    private readonly noteService: NoteService,
    private readonly entitlementService: EntitlementService
  ) {}

  @Cron('*/10 * * * *', { name: 'document-sweeper', timeZone: 'Asia/Ho_Chi_Minh' })
  @OnOneInstance({ key: 'document:sweeper' })
  @WithoutOverlapping()
  async sweep(): Promise<void> {
    const stale = await this.repo.find({
      select: { id: true },
      where: { modified: true, updatedAt: LessThan(new Date(Date.now() - 10 * 60_000)) },
      take: 500,
    });
    if (stale.length === 0) return;
    const workspaces = await this.noteService.batchGetNoteWorkspaces(stale.map((d) => d.id));
    const ents = await this.entitlementService.batchGet([...new Set(workspaces.values())]);
    for (const [documentId, ws] of workspaces) {
      const e = ents.get(ws);
      if (!e || !feature(e, 'auto_commit')) continue;
      try {
        await this.documentService.commitIfDue({ documentId, trigger: 'sweeper', force: true, minIntervalSeconds: 0 });
      } catch (err) {
        this.logger.warn({ err, documentId }, 'sweeper commit failed');
      }
    }
  }
}
```

`cron.module.ts`: imports `DocumentModule`, `NoteModule`, `EntitlementModule`, `TypeOrmModule.forFeature([DocumentEntity, RevisionEntity])`; providers `SweeperCron`, `RevisionPurgeCron`.

- [ ] **Step 4: Scaling toggle + app module**

Config (`config.ts` / `config.factory.ts`):

```ts
export const scalingConfigSchema = z
  .object({ enabled: z.boolean(), redisUrl: z.string().optional() })
  .refine((c) => !c.enabled || !!c.redisUrl, { message: 'NOTOPIA_DOCUMENT_REDIS_URL is required when scaling is enabled' });
```

from `NOTOPIA_DOCUMENT_SCALING_ENABLED === 'true'` and `NOTOPIA_DOCUMENT_REDIS_URL`.

`app.module.ts` imports: `ScheduleModule.forRoot()`, `LocksSetupModule.forRoot({ redisUrl: scaling.enabled ? scaling.redisUrl : undefined })` (read the env through the same factory function, since `forRoot` is static), `CronModule`. `main.ts` already calls `app.enableShutdownHooks()`.

`hocuspocus.ts`: when scaling is enabled, prepend `new Redis({ ... })` from `@hocuspocus/extension-redis` (parse host/port from the URL, or pass an ioredis instance if the extension accepts one) and set the server `name` to `os.hostname()`.

Dependencies (`apps/document/package.json`, via catalog): `@nestjs/schedule`, `@nestjs/locks`, `@hocuspocus/extension-redis` (`4.0.0`), `ioredis`; `pnpm install`.

`deploy/compose/services/document.yaml`: add

```yaml
  document_redis:
    container_name: notopia-document_redis
    image: redis:8.4.3-alpine3.22
    command: ["redis-server", "--appendonly", "yes", "--maxmemory-policy", "noeviction"]
    profiles: ["scaling"]
    ports:
      - 6374:6379
```

- [ ] **Step 5: Tests, build, verify, commit**

```bash
nx test document
oxfmt apps/document && nx lint document --fix && nx build document
```

Manual: with `NOTOPIA_DOCUMENT_SCALING_ENABLED=false` the app starts with no Redis. With `docker compose --profile scaling up -d document_redis` and `NOTOPIA_DOCUMENT_SCALING_ENABLED=true NOTOPIA_DOCUMENT_REDIS_URL=redis://localhost:6374`, run two instances on different ports: edits sync across instances; temporarily set the sweeper to `* * * * *` and see `document:sweeper` run on one instance per tick.

```bash
git add apps/document deploy/compose/services/document.yaml pnpm-workspace.yaml pnpm-lock.yaml
git commit -m "feat(document): sweeper and revision purge crons with optional redis scaling"
```

---

### Task 10: Docs, API descriptions, web outline, E2E

**Files:**
- Modify: `apps/document/AGENTS.md`, `internal/note/AGENTS.md`, `api/document/paths/*` (commit + revisions descriptions)
- Create: `docs/superpowers/notes/billing-web-outline.md` (handoff for the frontend owner)

- [ ] **Step 1: Docs**

- `apps/document/AGENTS.md`: commit pipeline now `commitIfDue` (manual/auto/sweeper), contributors, `AutoCommitExtension`, crons, scaling toggle; remove "Commit is currently manual only".
- `internal/note/AGENTS.md`: entitlement gates (members, payer guard, graph), trash purge cron, billing SSE event, TODO: gate `published_notes` when publishing is implemented.
- `api/document` descriptions: commit → "Plus/pro workspaces also auto-commit (`auto_commit`)"; revisions → "Revisions older than plan `revision_history_days` are purged daily; the latest is always kept. Filter with `trigger`."

- [ ] **Step 2: Web outline (no code)**

`docs/superpowers/notes/billing-web-outline.md`: billing settings tab (plan, status, renewal/access-until, payer + claim button, Upgrade → checkout URL, Manage → portal URL, Cancel now with `GET …/cancel-now/preview` amount); pricing page from `GET /billing/plans` (VND formatting `Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' })`); global 402 handler on `@notopia-uit/api-gen` → upgrade dialog with `details.requiredPlan`; 409 `billingPayerMustTransfer` message; `<FeatureGate>` reading entitlements, refetched on SSE `WorkspaceBillingChangedEvent` and polled every 2 s (max 20 s) after returning with `?checkout=success|payer` or `?portal=return`; downgrade warning about retention; editor "Saved automatically" when `auto_commit`; revision list collapses `trigger=auto`; admin pages for plans/prices/policy (group `notopia-admins`).

- [ ] **Step 3: End-to-end checklist**

1. Free workspace with 3 members → adding a 4th → 402 `requiredPlan: plus`.
2. Upgrade via checkout → SSE `WorkspaceBillingChangedEvent` → adding the 4th member succeeds.
3. Free workspace graph → 402; note graph depth > 5 is clamped.
4. Payer demotion → 409; another owner claims payer (plan 03) → demotion allowed.
5. Plus workspace: auto revisions every ≥ 300 s while editing; closing the tab forces one; identical content → no revision.
6. Trash purge and revision purge with shortened retention (admin API) and a temporary `* * * * *` schedule.
7. Stop billing → member add returns 503; editing continues (auto-commit skipped, logged).

- [ ] **Step 4: Commit**

```bash
nx lint api && nx run api:gen
git add apps/document/AGENTS.md internal/note/AGENTS.md api/document docs/superpowers/notes packages/api-document-nestjs-server packages/api-gen
git commit -m "docs: enforcement notes, api descriptions and web outline"
```
