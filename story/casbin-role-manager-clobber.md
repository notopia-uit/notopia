# The Helper Enforcer That Erased Everyone's Membership

_How a "harmless" temporary casbin enforcer, built only to parse a CSV, silently replaced the real enforcer's role manager, and why every workspace member got `403` after a restart._

Written with AI during billing-03, October 2026. Fixed in `8e55c60d`, tests added in `f85c26e5`.

---

## The Symptom

During the billing end-to-end run, the authorization service was restarted, and suddenly:

```
GET /note/workspaces/{id}/tree        → 403
GET /billing/workspaces/{id}/entitlements → 403 user 112 lacks billing permission
```

The user was the workspace **owner**. The membership row was sitting right there in the authorization database:

```
g, user:112, owner, workspace:01a13526-…
```

Restarting with `APP_ENV=production` instead of `development` seemed to "fix" it, so at first it looked like a dev-only fixtures problem. It wasn't. It was worse.

---

## A 60-second Casbin Primer

Our model (`internal/authorization/app/model.conf`):

```ini
[request_definition]
r = sub, dom, obj, act          # user, workspace, object type, action

[policy_definition]
p = sub, obj, act               # role, object class, action

[role_definition]
g = _, _, _                     # user → role, inside a workspace (domain)
g2 = _, _                       # object type → object class

[matchers]
m = g(r.sub, p.sub, r.dom) && g2(r.obj, p.obj) && r.act == p.act
```

There are two kinds of data:

| Rows                    | Example                           | Where it lives                   | Changes                                          |
| ----------------------- | --------------------------------- | -------------------------------- | ------------------------------------------------ |
| `p` (permission matrix) | `p, owner, workspace, read`       | `policy.csv` (static)            | when we change what a role may do                |
| `g2` (object classes)   | `g2, note, workspace_item`        | `policy.csv` (static)            | rarely                                           |
| `g` (memberships)       | `g, user:112, owner, workspace:W` | **database**, written at runtime | every time someone joins, leaves or changes role |

The important bit for this story: casbin does **not** answer `g(...)` by scanning `g` rows. When policies are loaded, it builds an in-memory graph called the **role manager** (`RM`) out of the `g` rows, and `g(user, role, domain)` asks that graph.

```
model["g"]["g"]  ──►  Assertion { Policy: [...g rows...], RM: *RoleManager }
                                                      ▲
                                   Enforce() → g() asks this
```

The `g` rows are the data; the role manager is the index built from them. If you lose the index, the data is still in the database, but every check says "no".

---

## The Code That Did It

The service has the static matrix in an embedded `policy.csv` and needs to push it into the database on startup. Casbin has no "load policies from a string into my existing enforcer" helper. Its file adapter reads from disk paths only, and the string adapter is an adapter, not a loader. So a helper was written that looked perfectly reasonable:

```go
// policyloader.go (old)
func LoadPoliciesFromString(enforcer *casbin.TransactionalEnforcer, csv string) error {
    tempAdapter := stringadapter.NewAdapter(csv)
    tempModel := enforcer.GetModel()                                   // ← same model object
    tempEnforcer, err := casbin.NewTransactionalEnforcer(tempModel, tempAdapter)
    …
    policies, _ := tempEnforcer.GetPolicy()
    enforcer.AddPoliciesEx(policies)                                   // copy rows across
    …
}
```

The idea: let a throwaway enforcer parse the CSV (it understands `p`, `g`, `g2` lines), then copy the rows into the real one. The throwaway enforcer is never used again, so what could go wrong?

---

## What Casbin Does When You Create an Enforcer

`casbin.NewTransactionalEnforcer(m, adapter)` → `InitWithModelAndAdapter` (casbin v3.11.0, `enforcer.go`):

```go
e.model = m          // keeps the *same* model you passed
…
e.initialize()       // → initRmMap()
…
e.LoadPolicy()
```

and `initRmMap` (`enforcer.go:585`):

```go
for ptype, assertion := range e.model["g"] {
    …
    assertion.RM = <new empty role manager>   // writes onto the assertion object
    e.rmMap[ptype] = assertion.RM
}
```

`enforcer.GetModel()` returns the real enforcer's model, so `assertion` here is the **real enforcer's `g` assertion**. Creating the helper enforcer replaced the real enforcer's role manager with a brand-new, empty one. Then the helper loaded `policy.csv`, which has no `g` rows, so that new role manager stayed empty.

```
before:  real enforcer ─► model["g"]["g"].RM = {user:112 → owner @ W, …}   (built from DB)
helper:  NewTransactionalEnforcer(sameModel, csvAdapter)
after:   real enforcer ─► model["g"]["g"].RM = {}                            (empty)
```

Every `g(user, role, workspace)` now returned false. Every permission check failed. The rows were still in Postgres, and the policy list in memory still contained them; only the index was gone.

### "But casbin's LoadPolicy copies the model"

That was the first theory, and it's half right. `LoadPolicy` does work on `model.Copy()`, so loading doesn't mutate the shared model's policy rows. But `Assertion.copy()` copies the `RM` field **by pointer** (`model/assertion.go`), and `initRmMap` runs before `LoadPolicy` anyway, directly on the shared assertions. Copying the model doesn't protect the role manager.

---

## Why It Looked Like a Dev-Only Bug

Startup in `internal/authorization/server.go`:

```go
a.BootStrapPolicies(ctx)            // always
if cfg.AppEnv == development {
    a.LoadFixtures(ctx)             // dev memberships from policy_fixtures.csv
}
```

- **Before the billing work:** `BootStrapPolicies` used `AddPolicies` directly (no helper), and only `LoadFixtures` went through `LoadPoliciesFromString`. Production never touched the helper; development did, through fixtures. That matches "dev broken, prod fine". We couldn't reproduce the exact development failure afterwards: in a repro, the fixtures' `g` adds rebuilt the role links and masked the problem. Either way, only the path that went through the helper misbehaved.
- **After `24c6f124`:** to stop the hardcoded list drifting from `policy.csv`, `BootStrapPolicies` was switched to the same helper. Now **production** clobbered the role manager on every start. In development the fixtures' `g` adds happened to rebuild the links and mask it.

In both versions the root cause was the same: building a second enforcer on the live model.

The bootstrap test passed the whole time, because it only checked `HasPolicy(...)`, which reads the policy rows that were still there. It never called `Enforce`, which is the only thing that uses the role manager.

---

## The Fix

Don't build a second enforcer. Parse the CSV into a **standalone** model and apply the rules to the real enforcer through its normal API (`internal/authorization/app/policysync.go`):

```go
func parsePolicyCSV(csv string) (model.Model, error) {
    m, _ := model.NewModelFromString(ModelConf)       // fresh model, owns its own assertions
    for line := range strings.SplitSeq(csv, "\n") {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        if err := persist.LoadPolicyLine(line, m); err != nil {   // casbin's own CSV line parser
            return nil, err
        }
    }
    return m, nil
}

func syncPolicies(e *casbin.TransactionalEnforcer, csv string, prune bool) error {
    desired, _ := parsePolicyCSV(csv)
    p, _ := desired.GetPolicy("p", "p")
    if prune { /* RemovePolicies(current p rows not in csv) */ }
    e.AddPoliciesEx(p)                                  // adds only missing rows
    for _, pt := range []string{"g", "g2"} {
        rules, _ := desired.GetPolicy("g", pt)
        e.AddNamedGroupingPoliciesEx(pt, rules)         // never removes g rows
    }
    return nil
}

func (a *App) BootStrapPolicies(ctx) error { return syncPolicies(a.Enforcer, PolicyCSV, true) }
func (a *App) LoadFixtures(ctx) error      { return syncPolicies(a.Enforcer, PolicyFixturesCSV, false) }
```

`persist.LoadPolicyLine` is the same function casbin's adapters use to read CSV lines, so we get casbin's own parsing without casbin's enforcer setup. Nothing touches the real enforcer's assertions except its own API.

What `policy.csv` means now:

- It is the **single source of truth** for `p` and `g2`.
- Every start adds missing rows and **removes `p` rows no longer in the CSV**, so deleting a permission from the file deletes it from every database.
- `g` rows (memberships) are never removed: they're runtime data.

---

## The Tests That Would Have Caught It

1. **Enforce after restart.** In a file-backed sqlite database, persist a runtime membership, create a _new_ enforcer on the same database (a restart), run only `BootStrapPolicies`, then assert `Enforce(user, workspace, "workspace", "read") == true`. This failed on the old code with "persisted membership must survive bootstrap".
2. **Prune.** A stale `p` row in the database that isn't in `policy.csv` disappears after bootstrap, and running bootstrap twice is harmless.
3. **Same matrix, both paths.** `TestModels` (36 permission cases) now runs twice: once with casbin's string adapter reading the CSV directly, once with `BootStrapPolicies` + `LoadFixtures` syncing into sqlite. Both must give identical answers.
4. **Parser.** Comments, blank lines, whitespace, `p`/`g`/`g2` rows, the embedded files (16 `p`, 2 `g2`), and an unknown rule type (`g3`) must error.

---

## Takeaways

- In casbin, **`g` rows are data and the role manager is the index**. A check can fail with the data intact. Test with `Enforce`, not `HasPolicy`.
- **Never pass `enforcer.GetModel()` to another enforcer constructor.** Enforcer setup writes role managers onto the model's assertions, so the two enforcers end up sharing, and overwriting, them.
- To load policy text, parse with `persist.LoadPolicyLine` into a model from `model.NewModelFromString`, then use `AddPoliciesEx` / `AddNamedGroupingPoliciesEx` / `RemovePolicies` on the real enforcer.
- A restart is a test case: write a test that persists data, builds a fresh process-like instance, and checks behaviour, not rows.
