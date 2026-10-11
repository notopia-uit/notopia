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
- Cache: in-memory, 60 s TTL per workspace. Flush one workspace on
  `events.integration.billing.subscription.changed`; flush all on
  `events.integration.billing.plan.changed`.
- Billing unreachable: fail closed for gates that block growth (return 503), fail open to the
  last cached value when present. Crons skip the batch and retry next tick.
- `-1` = unlimited. Missing feature key = false; missing limit key = 0.
- Downgrade policy: block growth, never delete user data except via retention crons. No
  post-downgrade buffer; the web warns before downgrade.
- 402 body: `PaymentRequiredError` from `api/common` (`featureNotInPlan` / `planLimitExceeded`,
  `details.requiredPlan` = cheapest tier unlocking it, from `ListPlans`).
- Cron actor: the service account id from `Identity.ServiceAccountUserID`
  (note `1001`, document `1002`).

## note (`internal/note`)

| Item | Change |
| --- | --- |
| `app/svc_entitlement.go` + `infra/service/billing.go` | Port + gRPC adapter, mirrors `svc_authorization.go` / `infra/service/authorization.go`. Includes the cache. |
| `controller/event/` | Consume `billing.subscription.changed`, `billing.plan.changed` → cache flush. |
| `errs/` | `FeatureNotInPlan`, `PlanLimitExceeded` → 402 (`PaymentRequiredError`) in `controller/http/err.go`; `BillingPayerMustTransfer` → 409. |
| `cmd_updateworkspacemembers.go` | `members` limit when the resulting count grows past it (existing members never removed). Payer guard: demoting or removing the payer while `GetWorkspacePayer.requires_payer` → 409 `billingPayerMustTransfer`. |
| `cmd_leaveworkspace.go` | Payer guard. |
| `cmd_publishnote.go` | `published_notes` limit (count published notes in the workspace). |
| `q_getnotegraph.go` | Clamp `Depth` to `note_graph_max_depth` before the existing `<= 0 → unlimited` rule. |
| `q_getworkspacegraph.go` | Require `workspace_graph`. |
| `cmd_changeworkspaceslug.go`, `cmd_publishworkspace.go`, search token | Not gated. |
| `api/note/paths/*` | One-line plan note in each gated operation's description + `402` response ref. |

**Trash purge cron**

- `app/cmd_purgeexpiredtrash.go`: no permission check, actor = note service account, reuses
  the domain trash / `Note.PermanentlyDelete` path so `NoteDeletedEvent` (and, via 01, the
  `note.deleted` integration event) fire.
- Flow: `SELECT DISTINCT workspace_id` of root trashed items (`trashed_by = 'purpose'`) with
  `trashed_at < now − 1 day` → batches of 500 → `BatchGetWorkspaceEntitlements` → for each
  workspace delete root items older than `trash_retention_days` (children cascade as today).
- `controller/cron/`, daily 03:00 `Asia/Ho_Chi_Minh`, `pkg/cron` with Redis locker on
  `note_redis`, `// NOTE:` move to `cmd/note-worker`.

## document (`apps/document`)

### Entitlement client

`src/entitlement/` gRPC client module to billing (mirrors `src/authorization/`), cache as
above, Kafka handlers for the two billing topics. Workspace of a document comes from
`NoteService.getWorkspaceByNote` (cached alongside).

### Entities (TypeORM migration)

| Entity | New columns |
| --- | --- |
| `DocumentEntity` | `lastCommittedAt timestamptz null`, `updatedAt timestamptz` (`@UpdateDateColumn`), `pendingContributorIds text[] default '{}'` |
| `RevisionEntity` | `authorId text null`, `trigger` enum `manual`/`auto`/`sweeper` default `manual`, `contributorIds text[] default '{}'`, `contentHash text null` |

### Contributors and author

- `onChange` with `connection` set: append `context.user.id` to an in-memory
  `Map<docId, string[]>` (move to end if present; last = most recent editor).
- `store`: merge the in-memory list into `pendingContributorIds` in the same upsert (SQL
  dedupe keeping order), then clear the in-memory entry. Works across instances.
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

- If workspace has `auto_commit`: `commitIfDue({trigger: 'auto', force: clientsCount === 0})`.
- Last client leaving forces a commit, so the final state is always a revision.
- Errors are logged, never thrown (a throw would keep the document in memory).
- Hocuspocus `debounce` / `maxDebounce` from config (defaults 2 000 / 10 000). They are
  server-wide only; per-plan behaviour is the commit interval, not the debounce.

### Sweeper cron (every 10 min)

Documents with `modified = true` and `updatedAt < now − 10 min` → note `BatchGetNoteWorkspaces`
→ billing batch entitlements → for workspaces with `auto_commit`:
`commitIfDue({trigger: 'sweeper', force: true})`.

### Revision purge cron (daily 03:30)

1. billing `ListPlans` → `minDays = min(revision_history_days where ≥ 0)`; none → stop.
2. `SELECT DISTINCT document_id FROM revisions WHERE created_at < now − minDays`.
3. Batches of 500 → note `BatchGetNoteWorkspaces` → billing `BatchGetWorkspaceEntitlements`.
4. Per document: hard delete revisions older than its plan's days, **always keeping the
   latest**. Unlimited (`-1`) → skip. Documents missing from the note map → skip.

### Scaling toggle

```
NOTOPIA_DOCUMENT_SCALING_ENABLED=false   # default
NOTOPIA_DOCUMENT_REDIS_URL=              # required only when true (zod refine)
```

- `false`: no Redis; `LocksSetupModule` in-memory store.
- `true`: `@hocuspocus/extension-redis` (server `name` = hostname; Redis blocks concurrent
  stores across instances), `LocksSetupModule` Redis store.
- Compose: optional `document_redis` behind a profile.

### Crons and locks

`src/cron/*.cron.ts`: `@Cron(expr, { name, timeZone: 'Asia/Ho_Chi_Minh' })` +
`@OnOneInstance({ key })` + `@WithoutOverlapping()`; `app.enableShutdownHooks()`.
`// NOTE:` move to a document worker when split.

### API

- Revision list: optional `trigger` filter; response adds `authorId`, `trigger`,
  `contributorIds`.
- `api/document/` descriptions: auto-commit (plus/pro), revision retention per plan.

## Web (`apps/web`) — outline only, not implemented here

Billing settings tab (plan, status, renewal/access-until, payer, Upgrade, Manage, Cancel now
with refund preview, Transfer payer); "Claim billing" banner; pricing page from
`GET /billing/plans` in VND; global 402 handler → upgrade dialog using `details.requiredPlan`;
`<FeatureGate>` from `/entitlements`; downgrade warning; editor shows "Saved automatically" with
`auto_commit`, else manual Commit; revision list collapses `auto`; admin pages for plans,
prices, policy.

## Future AI service

Reads `ai_credits_monthly` via billing gRPC and keeps its own usage counter. Not designed here.

## Testing

| Item | Test |
| --- | --- |
| note gates | Handler tests with a fake entitlement port: under limit OK, at limit → `PlanLimitExceeded`, feature missing → `FeatureNotInPlan`, payer guard → `BillingPayerMustTransfer` |
| note cache | TTL expiry, flush on events |
| trash purge | Items older than retention deleted, newer kept, events emitted with service-account actor |
| `commitIfDue` | interval, force, hash-equal skip, contributor copy, author fallback |
| AutoCommitExtension | free plan → no commit; `clientsCount === 0` → forced |
| revision purge | keeps latest, respects per-plan days, skips unlimited |
| sweeper | commits only `auto_commit` workspaces |
| locks | two app instances on Redis run a cron once per tick |
| E2E (manual) | free workspace: 4th member → 402 with `requiredPlan: plus`; upgrade in sandbox → allowed within ≤ 60 s or immediately after `subscription.changed`; plus workspace: type, wait > 300 s → auto revision appears; close tab → revision forced |

## Acceptance

- All tests pass; `nx lint note document` clean; `nx gen api` no diff.
- `apps/document/AGENTS.md` and `internal/note/AGENTS.md` updated (auto-commit, gates, crons).
- Manual E2E scenarios pass against the sandbox from sub-project 03.
