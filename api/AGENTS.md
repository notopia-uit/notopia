# api

Contract: change yaml, run codegen, implement. Never hand-edit generated servers or clients.

## Surfaces

| Dir | Generates | Consumed by |
| --- | --- | --- |
| `note/` | Go gin strict server + models (oapi-codegen) → `pkg/api/note` | `internal/note/controller/http` |
| `document/` | NestJS abstract API classes (openapi-generator) → `packages/api-document-nestjs-server` | `apps/document` |
| `share/` | TS types (+ Go) for cross-service **Kafka event payloads** → `packages/api-share-gen`, `pkg/api/share` | `apps/document`, `apps/search-worker` |
| `common/` | — | `$ref` target: shared error responses and `securitySchemes` (OIDC, Oauth2) |
| `bundled/` | Redocly output, **generated — do not edit** | `api-web`, all codegen |

`share/` is not HTTP API. Defines event payload once across services. New integration event → add schema there first.

## File layout per surface

`openapi.yaml` indexes `paths:` entries via `$ref` to `paths/*.yaml`, one file per path, named after URL with `/` → `_` (e.g. `workspaces_{workspaceId}_members.yaml`). Reusable pieces in `components/{schemas,parameters,examples}/`. Each path file lists every method with its own `security`, `tags`, `operationId`, and error `$ref`s — copy neighboring file.

`operationId` determines generated method names; effectively public API.

## Pipeline

```
paths/*.yaml -> redocly bundle -> bundled/{note,document,share}.json
                       -> redocly join -> bundled/openapi.json  (the unified spec)
                             |-> oapi-codegen          -> pkg/api/*        (Go server)
                             |-> openapi-ts (hey-api)  -> packages/api-gen (React Query client)
                             `-> openapi-generator     -> packages/api-document-nestjs-server
```

- `nx gen api` — run everything after spec change
- `nx dev api` — Scalar on :9080 with watch; `nx preview api` — one-shot
- `nx mock api` — mock server from spec
- `nx lint api` — Redocly lint. Known exceptions in `.redocly.lint-ignore.yaml`; prefer fixing spec over adding entry

## Adding an endpoint

1. `paths/{name}.yaml` + register under `paths:` in surface's `openapi.yaml`
2. New schemas under `components/schemas/`; reuse `../../common/components/responses/*` for 400/401/404/500 for uniform error shapes
3. `nx gen api`
4. Implement generated method — Go: `internal/note/controller/http/` (`StrictServerInterface` won't compile until done); NestJS: override abstract method in `apps/document/src/*/\*.api.ts`
5. Frontend calls via `@notopia-uit/api-gen` — no manual fetch code

## Adding a new API surface (e.g. billing)

Create `api/{name}/` with own `openapi.yaml`, `project.json`, codegen config (`oapi-codegen.yaml` for Go, mirroring `api/note/`); add to `implicitDependencies` in `api/project.json` and `sources`/`dependsOn` in `apps/api-web/project.json`.