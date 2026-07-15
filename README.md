# Notopia

<div align="center">

**Markdown notes with real-time collaboration and graph view.**

[![Docs](https://img.shields.io/badge/docs-blue?logo=gitbook)](https://notopia-uit.github.io/notopia/)
[![API Reference](https://img.shields.io/badge/api-6BA539?logo=openapiinitiative&logoColor=white)](https://notopia-uit.github.io/notopia/api/index.html)
[![Github Landing](https://img.shields.io/badge/github_landing-yellow?logo=github)](https://github.com/notopia-uit/)
[![Codecov](https://img.shields.io/codecov/c/github/notopia-uit/notopia)](https://codecov.io/gh/notopia-uit/notopia)
[![Wakatime](https://wakatime.com/badge/github/notopia-uit/notopia.svg)](https://wakatime.com/badge/github/notopia-uit/notopia)

| Service           |                                                                                          Quality Gate                                                                                          |                                                                                      Bugs                                                                                      |                                                                                         Code Smells                                                                                          |                                                                                          Maintainability                                                                                          |
| :---------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------: | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------: | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------: | :-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------: |
| **Note**          |          [![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=notopia-uit_note)          |          [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=bugs)](https://sonarcloud.io/summary/new_code?id=notopia-uit_note)          |          [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=notopia-uit_note)          |          [![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_note&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=notopia-uit_note)          |
| **Web**           |           [![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=notopia-uit_web)           |           [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=bugs)](https://sonarcloud.io/summary/new_code?id=notopia-uit_web)           |           [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=notopia-uit_web)           |           [![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_web&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=notopia-uit_web)           |
| **Document**      |      [![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=notopia-uit_document)      |      [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=bugs)](https://sonarcloud.io/summary/new_code?id=notopia-uit_document)      |      [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=notopia-uit_document)      |      [![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_document&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=notopia-uit_document)      |
| **Authorization** | [![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization) | [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=bugs)](https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization) | [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization) | [![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_authorization&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=notopia-uit_authorization) |
| **Search Worker** | [![Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker) | [![Bugs](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=bugs)](https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker) | [![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker) | [![Maintainability](https://sonarcloud.io/api/project_badges/measure?project=notopia-uit_search-worker&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=notopia-uit_search-worker) |

</div>

## 🏢 Domain

Notopia is a markdown note-taking app with real-time collaboration, graph-based navigation, and multi-workspace support. It is **not** intended to replace existing solutions — it is a learning sandbox and university project.

**Core concepts:**

- **Workspaces** — isolated spaces with role-based access control (owner, admin, member)
- **Notes** — markdown documents organized in folders, with revision history
- **Graph** — navigational view showing note relationships (backlinks, references)
- **Collaboration** — real-time editing via CRDT (Yjs + Hocuspocus), presence indicators
- **Search** — full-text search powered by Meilisearch, synced via Kafka events

**Service responsibilities:**

| Service | Role |
| --- | --- |
| **Note** | Core business logic — workspaces, folders, notes, members |
| **Document** | Editor backend — real-time collab, revision history, file storage |
| **Authorization** | RBAC policy enforcement via Casbin |
| **Search Worker** | Indexes note content into Meilisearch via Kafka events |
| **Web** | Next.js frontend |

---

## 👨‍💻 Dev Flow

### Prerequisites

All tools are managed via `mise`:

```bash
mise i            # Install all tool versions
pnpm i            # Install JS/TS dependencies
go mod download   # Download Go dependencies
pnpm husky        # Initialize git hooks
```

### Infrastructure

```bash
just docker-up                    # Start core infrastructure (Postgres, Redis, Kafka, Meilisearch, etc.)
just docker-up-monitoring         # + Grafana, Prometheus, Loki, Tempo
compose --profile "dev-ui" up -d  # + DbGate, Redpanda Console
```

Local domains via Traefik (add to `/etc/hosts`):

| Domain | Service |
| --- | --- |
| `web.notopia.localhost` | Web frontend |
| `api.notopia.localhost` | API gateway |
| `authentik.notopia.localhost` | Identity provider |
| `meilisearch.notopia.localhost` | Search |
| `rustfs-api.notopia.localhost` | S3 storage |
| `redpanda.notopia.localhost` | Kafka console |
| `dbgate.notopia.localhost` | DB admin |

Default credentials: `notopiauit` / `notopiauit`

### Run Services

```bash
# Go services (direct run, no build step)
nx run note                        # Note service (:18081, debug: 23451)
nx run authorization               # Authorization service (:18089, debug: 23459)

# NestJS services (build + watch)
nx run document                    # Document service (debug: 9222)
nx run search-worker               # Search worker (debug: 9223)

# Frontend
nx dev web                         # Next.js (:3000)
nx dev api-web                     # Scalar API docs (:3002)
```

### Code Generation

API changes must start from `api/` (OpenAPI) or `proto/` (Protobuf):

```bash
nx gen api     # Generate API code (NestJS server, Go server, frontend client)
nx gen proto   # Generate protobuf code (Go, TypeScript)
```

### Database Migrations

```bash
nx run note:goose:up               # Apply note service migrations
nx run note:goose:down             # Rollback
nx run note:goose:create           # Create new migration
```

### Seed Data

```bash
just seed-init                     # Seed all services
just seed-init-note                # Seed note service only
just seed-init-document            # Seed document service only
just seed-init-search              # Seed search index
```

### Conventions

- **Do not source `.env`** — Nx auto-loads them; sourced env vars won't be overridden
- **JS/TS**: `oxlint` lint, `oxfmt` format, `tsgo` typecheck (not `tsc`)
- **Go**: `golangci-lint` lint, `go fmt` format
- **Commits**: conventional commit format
- **API contracts**: change specs in `api/` or `proto/`, never modify generated code directly
- **UI components**: install in `packages/ui` via `shadcn add <component> -c @notopia-uit/ui`
- **Temp files**: write to `./tmp/{projectName}`, never `/tmp/`
- **Pre-commit hook**: runs `nx affected` for typecheck + lint + format on changed files

### Useful Commands

```bash
just all                           # CI: lint + typecheck + build + gen in parallel
nx lint <project> --fix            # Auto-fix lint issues
nx affected                        # Run targets only on changed projects
find . -type l ! -exec test -e {} \; -print -delete   # Remove broken symlinks
```

### OTEL

<https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/>

---

## ☑️ TO-DO

### Backend

- [ ] Note service
  - [ ] Routing kafka message based on metadata workspace id if partitioning or listening
  - [ ] Some handlers should rename the UserID to actor ID, and... so do the domain event?
  - [ ] Some query handler inject domain repo for getting workspace id
- [ ] Document service
  - [ ] Currently only hocuspocus guard the document, other like create/update/delete revision not check, bcs I'm lazy
    > If do, should create a guard outside of it
  - [ ] Health check
  - [ ] Consider merging blocknote & hocuspocus into editor module
- [ ] Authorization service
  - [ ] Health check
- [ ] Cannot deal with `domain.com:8080/api/v1` base path
- [ ] no env validation for document, search-worker
- [ ] Health check to other services (api service, meili, postgres...)
- [ ] Connection pool max connections, idle, timeout for database, meili
- [ ] gin should be protected with `SetTrustedProxies`
- [ ] Event is tracked by either otel or correlation id.
      But, currently use wotel + kafka tracer, and partially correlation id but not really connected.
- [ ] Revision endpoints aren't protected with authorization

### Frontend

- [ ] use suspend with skeleton, use suspend query for streamming
- [ ] Manage server state by tanstack, not always useState
- [ ] Clean architecture is considerable? `https://www.freecodecamp.org/news/reusable-architecture-for-large-nextjs-applications/`
- [ ] Add custom theme (or not)
- [ ] Set up logger
- [ ] Refactor
  - [ ] Alert
  - [ ] Use tanstack for update cache instead invalidate it

### Both

- [ ] yjs isn't typesafety, like getting Ymap, and set value.
      May try to see other libs, how do they do
- [ ] Those NestJS logging, we need to find a better way to wrap those controller log. NestJS Pino only http? not microservice.
      And guess that we should either using middleware or interceptor
- [ ] Mutating `update, add, delete` workspace member doesn't send event to user client
