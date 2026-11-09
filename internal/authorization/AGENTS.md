# Authorization Service

Casbin-backed workspace roles and permission checks. gRPC only; no HTTP API.
Entrypoint: `cmd/authorization/`. Contract: `proto/authorization/authorization.proto`.

## Layout

- `app/` — flat, one file per RPC (`createworkspace.go`, `hasworkspacepermission.go`, …),
  each with `_test.go`. No `cmd_`/`q_` prefixes, unlike `internal/note/`.
- `app/casbin.go` — enforcer setup
- `policy/` — embedded `model.conf`, `policy.csv`, `policy_fixtures.csv`; `Seed`, `LoadFixtures`, `DetectDrift`
- `infra/` — `db.go` (Casbin postgres adapter), `casbin.go`, `integrationevent.go`
- `controller/grpc/` — RPC handlers; `errs/` → gRPC status in `controller/grpc/err.go`

## Casbin model (`policy/model.conf`)

Request: `sub, dom, obj, act` = user, workspace, object type, action.

```
m = g(r.sub, p.sub, r.dom)   # user has role in this workspace (domain-scoped)
 && g2(r.obj, p.obj)         # object type maps to a policy object class
 && r.act == p.act
```

Two easily confused grouping relations:

- `g` — **user → role, scoped by workspace**. DB-stored membership data.
- `g2` — **concrete object → object class**, static in `policy.csv`:
  `note → workspace_item`, `folder → workspace_item`.

`policy.csv` (`p` lines) holds only role→permission matrix for `owner`/`editor`/`viewer`
over `workspace` and `workspace_item`. Static config, not per-tenant data; change it to change role permissions.

`policy.csv` is source of truth for static matrix (`p` + `g2`). `policy.Seed` syncs it:
adds missing rows, prunes stale DB `p` rows. `policy_fixtures.csv` holds dev memberships
(`g`), add-only, loaded by `policy.LoadFixtures`; tests reuse it. DB `g` rows are runtime memberships,
never pruned. Sync parses CSV into standalone model, applies rules to real enforcer;
never build second enforcer on shared model: it replaces role manager, drops loaded `g` links.

Service never seeds or loads fixtures. Seeding: `cmd/authorization seed` (`internal/authorization/seed`, runs
`policy.Seed`, `just seed-authorization`). Dev fixtures: `cmd/authorizationfixtures`
(`internal/authorizationfixtures`, runs `policy.LoadFixtures`, `just fixtures-authorization`). Nothing
auto-loads in development.

At startup, service calls `policy.DetectDrift` once (any `p`/`g2` rule in `policy.csv` absent from
in-memory policy, or any `p` rule in memory absent from `policy.csv`). On drift, logs error and keeps running.
`policy` startup health check (`controller/health`) fails on same condition, leaving pod not ready.
Check intentionally in-memory: enforcer does not watch DB, so only restart picks up seed.

`editor` may `delete` a `workspace_item` but only `read` the `workspace` itself.

Billing permissions on `workspace`: `billing_read` (owner, editor, viewer) and
`billing_manage` (owner only).

## Adding a permission or role

1. Add `p` lines to `policy/policy.csv` (and `g2` line for new object type)
2. Extend enum in `proto/authorization/authorization.proto`, run `nx gen proto`
3. Map enum in `controller/grpc/service.go` and each caller's `svc_authorization.go`
4. `app/model_test.go` and `policy/policy_fixtures.csv` cover matrix — extend them

## Consumers

`internal/note` (`infra/service/authorization.go`) and `apps/document`
(`src/authorization/`) are the only callers, both via gRPC. `integrationevent.go` publishes
role-change and member-removal events; `apps/document` consumes them to drop live
Hocuspocus connections.