# Authorization Service

Casbin-backed workspace roles and permission checks. gRPC only — no HTTP API.
Entrypoint: `cmd/authorization/`. Contract: `proto/authorization/authorization.proto`.

## Layout

- `app/` — flat, one file per RPC (`createworkspace.go`, `hasworkspacepermission.go`, …),
  each with a `_test.go`. No `cmd_`/`q_` prefixes here, unlike `internal/note/`.
- `app/casbin.go`, `policysync.go` — enforcer setup and policy syncing
- `infra/` — `db.go` (Casbin postgres adapter), `casbin.go`, `integrationevent.go`
- `controller/grpc/` — RPC handlers; `errs/` → gRPC status in `controller/grpc/err.go`

## Casbin model (`app/model.conf`)

Request is `sub, dom, obj, act` = user, workspace, object type, action.

```
m = g(r.sub, p.sub, r.dom)   # user has role in this workspace (domain-scoped)
 && g2(r.obj, p.obj)         # object type maps to a policy object class
 && r.act == p.act
```

Two grouping relations, easy to confuse:

- `g` — **user → role, scoped by workspace**. Stored in the DB; this is the membership data.
- `g2` — **concrete object → object class**, static in `policy.csv`:
  `note → workspace_item`, `folder → workspace_item`.

`policy.csv` (`p` lines) holds only the role→permission matrix for `owner`/`editor`/`viewer`
over `workspace` and `workspace_item`. It is static config: change it to change what a role
can do; it is not per-tenant data.

`policy.csv` is the single source of truth for the static matrix (`p` + `g2`). `BootStrapPolicies`
syncs it on every start: adds missing rows and prunes stale `p` rows from the DB.
`policy_fixtures.csv` holds dev memberships (`g`), add-only, loaded by `LoadFixtures` only when
`AppEnv` is development; tests reuse it. DB `g` rows are runtime memberships and are never pruned.
Sync parses CSV into a standalone model and applies rules to the real enforcer; never build a second
enforcer on the shared model, it replaces the role manager and drops loaded `g` links.

Note `editor` may `delete` a `workspace_item` but only `read` the `workspace` itself.

Billing permissions on `workspace`: `billing_read` (owner, editor, viewer) and
`billing_manage` (owner only).

## Adding a permission or role

1. Add the `p` lines to `app/policy.csv` (and a `g2` line if it's a new object type)
2. Extend the enum in `proto/authorization/authorization.proto`, run `nx gen proto`
3. Map the enum in `controller/grpc/service.go` and in each caller's `svc_authorization.go`
4. `app/model_test.go` and `policy_fixtures.csv` cover the matrix — extend them

## Consumers

`internal/note` (`infra/service/authorization.go`) and `apps/document`
(`src/authorization/`) are the only callers, both over gRPC. `integrationevent.go` publishes
role-change and member-removal events that `apps/document` consumes to drop live
Hocuspocus connections.
