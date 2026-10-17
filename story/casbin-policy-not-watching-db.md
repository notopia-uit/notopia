# Casbin Remembers, the Database Changes

_Why a casbin enforcer doesn't see policy rows written by anyone else, and what that means for seeding and for running more than one authorization instance._

Written with AI during billing-03, October 2026. Not fixed yet; this is the problem statement for a later phase.

---

## How the Enforcer Holds Policy

The authorization service checks permissions with a casbin `TransactionalEnforcer` backed by Postgres (gorm adapter). At startup it calls `LoadPolicy()` once: every row of `casbin_rule` is read into memory, and the `g` rows are turned into the in-memory role graph (see [the role manager story](./casbin-role-manager-clobber.md)).

After that, **every `Enforce` call reads memory only**. The database is touched only when _this_ enforcer writes (`AddPolicy`, `RemoveGroupingPolicy`, …): it updates its memory and saves the row, through autosave.

```
process A: memory ◄── LoadPolicy() at start ── casbin_rule ──► LoadPolicy() at start ──► memory :process B
           A writes → A memory + DB                              B never hears about it
```

Casbin doesn't poll or listen to the database. A row written by someone else stays invisible until the process reloads.

---

## Where It Bites

### 1. A separate seed process

If `policy.csv` sync moves out of the service into `cmd/authorizationseed` (planned), the seed writes `p` rows straight into the database. A running authorization instance keeps answering from its old matrix until it restarts. A new permission (say `billing_manage`) is **denied** by already-running pods even though the row exists.

Workable rule for now: **seed first, then start or restart authorization.** In Kubernetes that's a seed Job before the rollout, and a pod that finds the matrix missing reports "not ready" through its startup probe so it gets restarted, instead of serving wrong answers.

### 2. More than one authorization instance

This one already exists in principle. Instance A handles `UpdateWorkspaceMembers` and writes the new `g` row; instance B still has the old role graph. A user just added to a workspace gets `403` whenever their request lands on B, until B restarts. Removing a member is worse: B keeps **allowing** them.

Today we run one instance, so nobody has seen it. Scaling authorization horizontally without fixing this would be a security bug.

---

## Ways Out

| Option              | How                                                                                                                                                                                                                         | Trade-off                                                                  |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| Single instance     | Keep `replicas: 1`                                                                                                                                                                                                          | No HA; the restart rule still applies to seeding                           |
| Periodic reload     | `e.LoadPolicy()` every N seconds                                                                                                                                                                                            | Simple, but stale for up to N seconds, and a full reload each time         |
| **Watcher (Redis)** | [`casbin/redis-watcher/v2`](https://pkg.go.dev/github.com/casbin/redis-watcher/v2): `e.SetWatcher(w)`; every write publishes on a Redis channel, and other instances run their update callback (classically `LoadPolicy()`) | Needs Redis (authorization has none today); near-instant                   |
| Watcher (Postgres)  | [`IguteChung/casbin-psql-watcher`](https://beta.pkg.go.dev/github.com/IguteChung/casbin-psql-watcher) over `LISTEN/NOTIFY`, also implements incremental `WatcherEx`                                                         | No new infrastructure; community-maintained, check casbin v3 compatibility |

A watcher only fires for writes that go **through an enforcer with the watcher set**. A seed binary would have to set the same watcher, or the running instances would still need a restart after seeding.

---

## Takeaways

- A casbin enforcer is a cache of the policy table with no invalidation by default.
- "Row exists in the database" ≠ "permission granted": check `Enforce` on the instance that serves traffic.
- Before running more than one authorization instance, add a watcher (Redis or Postgres `LISTEN/NOTIFY`), and make the seed binary notify it too.
