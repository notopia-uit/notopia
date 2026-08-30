# api/

The contract. Every HTTP surface starts here — change the yaml, run codegen, then implement.
Never hand-edit generated servers or clients.

## Surfaces

| Dir | Generates | Consumed by |
| --- | --- | --- |
| `note/` | Go gin strict server + models (oapi-codegen) → `pkg/api/note` | `internal/note/controller/http` |
| `document/` | NestJS abstract API classes (openapi-generator) → `packages/api-document-nestjs-server` | `apps/document` |
| `share/` | TS types (+ Go) for cross-service **Kafka event payloads** → `packages/api-share-gen`, `pkg/api/share` | `apps/document`, `apps/search-worker` |
| `common/` | — | `$ref` target: shared error responses and `securitySchemes` (OIDC, Oauth2) |
| `bundled/` | Redocly output, **generated — do not edit** | `api-web`, all codegen |

`share/` is not an HTTP API. It exists so an event payload is defined once instead of being
retyped in each service. New integration event → add a schema there first.

## File layout per surface

`openapi.yaml` is an index: `paths:` entries `$ref` into `paths/*.yaml`, one file per path,
named after the URL with `/` → `_` (e.g. `workspaces_{workspaceId}_members.yaml`). Reusable
pieces live in `components/{schemas,parameters,examples}/`. Each path file lists every
method with its own `security`, `tags`, `operationId`, and error `$ref`s — copy a
neighbouring file rather than starting fresh.

`operationId` is what generated method names come from, so it is effectively public API.

## Pipeline

```
paths/*.yaml -> redocly bundle -> bundled/{note,document,share}.json
                       -> redocly join -> bundled/openapi.json  (the unified spec)
                             |-> oapi-codegen          -> pkg/api/*        (Go server)
                             |-> openapi-ts (hey-api)  -> packages/api-gen (React Query client)
                             `-> openapi-generator     -> packages/api-document-nestjs-server
```

- `nx gen api` — run everything after a spec change
- `nx dev api` — Scalar on :9080 with watch; `nx preview api` — one-shot
- `nx mock api` — mock server from the spec
- `nx lint api` — Redocly lint. Known exceptions are in `.redocly.lint-ignore.yaml`;
  prefer fixing the spec over adding an entry

## Adding an endpoint

1. `paths/{name}.yaml` + register it under `paths:` in that surface's `openapi.yaml`
2. New schemas under `components/schemas/`; reuse `../../common/components/responses/*`
   for 400/401/404/500 so error shapes stay uniform
3. `nx gen api`
4. Implement the newly generated method — Go: `internal/note/controller/http/` (the
   `StrictServerInterface` will not compile until you do); NestJS: override the abstract
   method in `apps/document/src/*/\*.api.ts`
5. Frontend calls it via `@notopia-uit/api-gen` — no manual fetch code

## Adding a new API surface (e.g. billing)

Create `api/{name}/` with its own `openapi.yaml`, `project.json`, and codegen config
(`oapi-codegen.yaml` for Go, mirroring `api/note/`), then add it to `implicitDependencies`
in `api/project.json` and to the `sources`/`dependsOn` in `apps/api-web/project.json`.
