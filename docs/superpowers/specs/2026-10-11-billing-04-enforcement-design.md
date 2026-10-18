# Billing 04 — Enforcement in note and document

> Sub-project 4 of 4. Depends on 01, 02, 03.
> Order: 01 + 02 (parallel) → 03 → 04. Progress is tracked by the task checkboxes in each plan under `docs/superpowers/plans/`.

## Goal

Make plans matter: note and document read entitlements from billing, gate features and limits,
return 402, and run plan-aware retention. Document gets auto-commit with contributors and an
optional Redis scaling mode.

## Shared rules

- Entitlements come only from billing gRPC (`GetWorkspaceEntitlements`,
  `BatchGetWorkspaceEntitlements`, `GetWorkspacePayer`, `ListPlans`). No plan data stored in
  note or document.
- **No cache.** Every check calls billing gRPC directly; upgrades apply immediately. Callers
  keep hot paths cheap by checking local preconditions first (e.g. auto-commit checks
  `modified` and `lastCommittedAt` before calling billing).
- Billing unreachable: gates return 503, auto-commit skips (the sweeper catches up), crons skip
  the batch and retry next tick.
- Billing's proto `PlanTier` enum is mapped at each client into a local type (note: Go `app.PlanTier`
  `uint8`; document: TS union `'free' | 'plus' | 'pro'` from the `@notopia-uit/pb/billing` enum).
  `PLAN_TIER_UNSPECIFIED` (or any unknown value) is an error in both services, no fallback plan.
- Feature/limit keys are typed in each consumer, never raw strings. note: `app.PlanFeature`,
  `app.PlanLimit` (`uint8` enums) and `app.PlanQuota` (`int64`, `PlanQuotaUnlimited = -1`,
  `Allows(n)`); the adapter parses the proto's string-keyed maps and skips keys the service does
  not use. document: TS const objects for the keys it uses.
- `-1` = unlimited. Missing feature key = false; missing limit key = 0.
- Downgrade policy: block growth, never delete user data except via retention crons. No
  post-downgrade buffer; the web warns before downgrade.
- 402 body: `PaymentRequiredError` from `api/common` (`featureNotInPlan` / `planLimitExceeded`,
  `details.requiredPlan` = cheapest tier unlocking it, from `ListPlans`).
- Cron actor: the service account id from `Identity.ServiceAccountUserID`
  (note `1001`, document `1002`).
- Cron jobs are configured from env, every field with a default: `name` (also the lock key),
  `schedule`, `timeout`, `enabled` (default `true`). `enabled=false` lets a future
  `{svc}-worker` own the jobs while the API process turns them off, with no code change.
  Go: shared `commonconfig.CronJob`; env `NOTOPIA_{SVC}_CRON_{JOB}_{NAME|SCHEDULE|TIMEOUT|ENABLED}`.
- Cron controllers never import Redis. `pkg/cron` exports `Locker` (alias of `gocron.Locker`);
  infra builds it from the service's shared Redis client and wire injects it. Applies to
  billing's existing cron too (refactored in this sub-project).

## note (`internal/note`)

| Item                                                                  | Change                                                                                                                                                                                                          |
| --------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `app/svc_billing.go` + `infra/service/billing.go`                     | `BillingSvc` port + gRPC adapter, mirrors `svc_authorization.go` / `infra/service/authorization.go`. Named after the service it calls (it also serves payer info).                                              |
| `errs/`                                                               | `FeatureNotInPlan`, `PlanLimitExceeded` → 402 (`PaymentRequiredError`) in `controller/http/err.go`; `BillingPayerMustTransfer` → 409.                                                                           |
| `cmd_updateworkspacemembers.go`                                       | `members` limit when the resulting count grows past it (existing members never removed). Payer guard: demoting or removing the payer while `GetWorkspacePayer.requires_payer` → 409 `billingPayerMustTransfer`. |
| `cmd_leaveworkspace.go`                                               | Payer guard.                                                                                                                                                                                                    |
| `cmd_publishnote.go`                                                  | `published_notes` limit (count published notes in the workspace). **Deferred**: publishing is a stub today; the gate lands with the publish feature (plan 04, decision 1).                                      |
| `q_getnotegraph.go`                                                   | Clamp `Depth` to `note_graph_max_depth` before the existing `<= 0 → unlimited` rule.                                                                                                                            |
| `q_getworkspacegraph.go`                                              | Require `workspace_graph`.                                                                                                                                                                                      |
| `cmd_changeworkspaceslug.go`, `cmd_publishworkspace.go`, search token | Not gated.                                                                                                                                                                                                      |
| `api/note/paths/*`                                                    | One-line plan note in each gated operation's description + `402` response ref.                                                                                                                                  |

**Permanent delete refactor (before the purge)**

The "notes + descendant notes first, then folders" delete is duplicated in
`PermanentlyDeleteFolder`, `PermanentlyDeleteWorkspaceItems` and `EmptyTrash`.

- Domain: `TrashService.PermanentlyDeleteNotes(notes, userID)` and
  `PermanentlyDeleteFoldersWithChildren(folders, childNotes, userID)`, pure like the existing
  `TrashFoldersWithChildren` / `RestoreFoldersWithChildren`.
- App: one loader (ids → notes, folders, descendant notes, all `ForUpdate`) shared by those
  three handlers and the purge.
- The purge cannot reuse `PermanentlyDeleteWorkspaceItemsHandler`: its permission check would
  reject the service account, and handlers do not call handlers.

**Trash purge cron**

- `app/cmd_purgeexpiredtrash.go`: no permission check, actor = note service account, uses the
  shared loader + `TrashService` so `NoteDeletedEvent` (and, via 01, the `note.deleted`
  integration event) fire.
- Flow: `SELECT DISTINCT workspace_id` of root trashed items (`trashed_by = 'purpose'`) with
  `trashed_at < now − 1 day` → batches of 500 → `BatchGetWorkspaceEntitlements` → for each
  workspace delete root items older than `trash_retention_days` (children cascade as today).
- `controller/cron/`, default daily 03:00 in the process zone (`TZ` env, passed as `time.Local`),
  `pkg/cron` with the injected `Locker`, `// NOTE:` move to `cmd/note-worker`.
- note's Redis client moves from `infra/workspaceevent` into `infra/common` (next to the Kafka
  publisher), shared by the SSE hub and the cron locker. The hard-coded `DB: 1` goes away (SSE
  uses pub/sub, which ignores the database number).

### Billing change notifications (SSE)

- Consume `events.integration.billing.subscription.changed` → publish SSE
  `WorkspaceBillingChangedEvent {workspaceId}` on that workspace's channel through
  `WorkspaceEventHub`.
- Schema `api/note/components/schemas/WorkspaceBillingChangedEvent.yaml`, added to the SSE
  events union in `workspaces_{workspaceId}_events.yaml`.
- Clients refetch `/billing/workspaces/{id}/entitlements` on that event.

## document (`apps/document`)

### Entitlement client

`src/entitlement/` gRPC client module to billing (mirrors `src/authorization/`). Workspace of a
document comes from `NoteService.getWorkspaceByNote`.

### Entities (TypeORM migration)

| Entity           | New columns                                                                                                                                    |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `DocumentEntity` | `lastCommittedAt timestamptz null`, `updatedAt timestamptz` (`@UpdateDateColumn`), `pendingContributorIds text[] default '{}'`                 |
| `RevisionEntity` | `authorId text null`, `trigger` enum `manual`/`auto`/`sweeper` default `manual`, `contributorIds text[] default '{}'`, `contentHash text null` |

### Contributors and author

- `onChange` with `connection` set: append `context.user.id` to an in-memory
  `Map<docId, string[]>` (move to end if present; last = most recent editor).
- `ContributorFlushExtension.onStoreDocument`, registered **before** the Hocuspocus Redis
  extension: if the local list is non-empty, merge it into `pendingContributorIds` with a
  merge-only `UPDATE` (SQL dedupe keeping order), then clear the entry.
- Why before the Redis extension: with scaling on, only the instance holding the Redis store
  lock runs the later store hooks; the others get `SkipFurtherHooksError`. Flushing earlier
  means every instance writes its own editors, so none are lost when a document unloads on an
  instance that never won the lock. Same code path with scaling off.
- Commit: copy `pendingContributorIds` to `RevisionEntity.contributorIds`, clear it.
- `authorId`: manual → HTTP user; auto / sweeper → last element of `pendingContributorIds`,
  fallback document service account. `lastContext` is not used (local-origin changes carry
  an empty context).
- Kafka `document.committed.userId` = `authorId`.

### Commit

`DocumentService.commitIfDue({documentId, trigger, authorId?, force})`, one transaction with the
`pessimistic_write` lock:

1. Skip unless `modified`.
2. Skip unless `force` or `trigger = manual` or `now − lastCommittedAt ≥ auto_commit_min_interval_seconds`.
3. Compute `contentHash` (sha256 of blocks JSON). Equal to the latest revision's hash → only
   clear `modified`, no revision, no event.
4. Else save revision (`authorId`, `trigger`, `contributorIds`, `contentHash`), set
   `lastCommittedAt`, clear `modified` and `pendingContributorIds`, emit `document.committed`.
5. After commit, `HocuspocusService.setModified(id, false)` if loaded.

The existing manual `commitDocument` becomes `commitIfDue({trigger: 'manual', force: true})`
and is available on every plan.

### Auto-commit extension

`AutoCommitExtension implements Extension`, `afterStoreDocument` (runs in the same `saveMutex`
right after the Database extension's store):

- If workspace has `auto_commit`: `commitIfDue({trigger: 'auto', force: false})`.
- No forced commit when the last client leaves: `clientsCount` is per instance, so with scaling
  on it would fire while others still edit on another instance. The content is already
  stored; the sweeper turns the final edits into a revision within one sweeper interval.
- Errors are logged, never thrown (a throw would keep the document in memory).
- Hocuspocus `debounce` / `maxDebounce` from config (defaults 2 000 / 10 000). They are
  server-wide only; per-plan behaviour is the commit interval, not the debounce.

### Sweeper cron (default every 10 min)

Documents with `modified = true` and `updatedAt < now − idle` (config, default 10 min) → note `BatchGetNoteWorkspaces`
→ billing batch entitlements → for workspaces with `auto_commit`:
`commitIfDue({trigger: 'sweeper', force: true})`.

### Revision purge cron (default daily 03:30)

1. billing `ListPlans` → `minDays = min(revision_history_days where ≥ 0)`; none → stop.
2. `SELECT DISTINCT document_id FROM revisions WHERE created_at < now − minDays`.
3. Batches of 500 → note `BatchGetNoteWorkspaces` → billing `BatchGetWorkspaceEntitlements`.
4. Per document: hard delete revisions older than its plan's days, **always keeping the
   latest**. Unlimited (`-1`) → skip. Documents missing from the note map → skip.

### Redis and scaling (separate config)

```
NOTOPIA_DOCUMENT_REDIS_URL=redis://localhost:6374   # redis config, optional
NOTOPIA_DOCUMENT_SCALING_ENABLED=false              # scaling config, default false
```

- `redis` and `scaling` are separate config sections. Cross-check in `env.validation.ts`:
  `scaling.enabled` requires `redis.url`.
- Locks (`LocksSetupModule`) use Redis whenever `redis.url` is set, else the in-memory store.
- `scaling.enabled` only adds `@hocuspocus/extension-redis` (server `name` = hostname).
- Compose: `document_redis` always runs (like `note_redis` / `billing_redis`), no profile.

### Crons and locks

`@Cron(expr)` is evaluated at class load, so it cannot read config. Jobs are registered in
`onModuleInit` through `SchedulerRegistry.addCronJob(cfg.name, new CronJob(cfg.schedule, …))`
(no `timeZone`; process zone from `TZ`), skipped when `cfg.enabled` is false. The body runs in
`Locks.withLock(cfg.name, …)`; `LockNotAcquiredError` = another instance has it, skip the tick.
Env: `NOTOPIA_DOCUMENT_CRON_{SWEEPER|REVISION_PURGE}_{NAME|SCHEDULE|ENABLED}`, plus
`NOTOPIA_DOCUMENT_CRON_SWEEPER_IDLE_SECONDS`. `app.enableShutdownHooks()`.
`// NOTE:` move to a document worker when split.

### API

- Revision list: optional `trigger` filter; response adds `authorId`, `trigger`,
  `contributorIds`.
- `api/document/` descriptions: auto-commit (plus/pro), revision retention per plan.

## Web (`apps/web`) — outline only, not implemented here

Billing settings tab (plan, status, renewal/access-until, payer, Upgrade, Manage, Cancel now
with refund preview, Transfer payer); "Claim billing" banner; pricing page from
`GET /billing/plans` in VND; global 402 handler → upgrade dialog using `details.requiredPlan`;
`<FeatureGate>` from `/entitlements`, refetched on SSE `WorkspaceBillingChangedEvent` and polled every 2 s for up to 20 s after returning from Checkout/Portal; downgrade warning; editor shows "Saved automatically" with
`auto_commit`, else manual Commit; revision list collapses `auto`; admin pages for plans,
prices, policy.

## Future AI service

Reads `ai_credits_monthly` via billing gRPC and keeps its own usage counter. Not designed here.

## Testing

Unit tests only where logic is non-obvious.

| Item             | Test                                                                                                                                                                                                                                                                                                 |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `commitIfDue`    | interval vs force, hash-equal skip, author fallback                                                                                                                                                                                                                                                  |
| revision purge   | keeps latest, per-plan days, skips unlimited                                                                                                                                                                                                                                                         |
| note limit gates | at limit → `PlanLimitExceeded` with `requiredPlan`; payer guard → `BillingPayerMustTransfer`                                                                                                                                                                                                         |
| E2E (manual)     | free workspace: 4th member → 402 `requiredPlan: plus`; upgrade in sandbox → allowed immediately; plus workspace: edit, wait > 300 s → auto revision; close tab → the sweeper commits it after the idle time; trash item past retention purged by cron (temporarily set retention to 0 via admin API) |

## Acceptance

- All tests pass; `nx lint note document` clean; `nx gen api` no diff.
- `apps/document/AGENTS.md` and `internal/note/AGENTS.md` updated (auto-commit, gates, crons).
- Manual E2E scenarios pass against the sandbox from sub-project 03.
