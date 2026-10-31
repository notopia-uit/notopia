# Monorepo

Manged with Nx. Package manager: pnpm (use ./pnpm-workspace.yaml), go (single go.mod for all packages), mise, buf

## Structure

```
api/                            # OpenAPI spec (one dir per API surface)
  common/                       # Shared responses, securitySchemes (OIDC, Oauth2)
  note/                         # note service spec    -> Go server (oapi-codegen)
  billing/                      # billing service spec -> Go server (oapi-codegen)
  document/                     # document service spec -> NestJS server (openapi-generator)
  share/                        # Cross-service event/DTO schemas -> @notopia-uit/api-share-gen
  bundled/                      # Redocly output (generated, do not edit)
proto/                          # Protobuf definitions
  note/, authorization/, billing/
apps/                           # JS/TS applications
  document/                     # Document service (NestJS) (nx: document)
  search-worker/                # Search worker (NestJS) sync to meilisearch (nx: search-worker)
  web/                          # NextJS (nx: web)
  api-web/                      # OpenAPI Spec rendered with Scalar React (nx: api-web)
  test-editor/                  # Blocknote editor sandbox (nx: test-editor)
packages/                       # JS packages
  api-document-nestjs-server/   # NestJS server codegen from OpenAPI spec, using openapi-generator
  api-gen/                      # Frontend API client codegen from OpenAPI spec, using heyapi/openapi-ts
  api-share-gen/                # Types for cross-service events, generated from api/share
  lib/                          # Shared code, blocknote model, yjs helpers,...
  lib-server/                   # Shared server-side (node-only) code
  pb/                           # Protobuf generated code, use grpc, protovalidate
  ui/                           # React UI components, using shadcn/ui (@notopia-uit/ui)
docs/                           # Vitepress Documentation (class, sequence, architecture diagrams, database, etc.), and AI (superpowers,...)
story/                          # The stories/blogs while doing this project (AI/human written)
cmd/                            # Go services (entrypoint only: main.go + wire)
  note/                         # Note service (nx: note)
  authorization/                # Authorization service (nx: authorization)
  authorizationseed/            # Authorization seed: Casbin policy matrix (nx: authorizationseed)
  authorizationfixtures/        # Authorization dev fixtures: memberships (nx: authorizationfixtures)
  billing/                      # Billing service, Stripe (nx: billing)
  notefixturegen/               # Dev fixture generator (fake notes from Obsidian vault)
  billingseed/                  # Billing seed: Stripe prices + portal config (nx: billingseed)
  billingfixtures/              # Billing dev fixtures (nx: billingfixtures)
internal/                       # Internal Go packages (the actual service code)
  note/
  authorization/
  authorizationseed/
  authorizationfixtures/
  billing/
  notefixturegen/
  billingseed/
  billingfixtures/
pkg/                            # Go packages
  api                           # Code gen from oapi-codegen
  common                        # Including commonhttp, commonconfig, commonhandler
  helper
  logging
  metadata
  otel
  pb
```

## Technologies

### API

- Redocly to bundle and join yaml spec to json
- Scalar render joined spec, mock server
- openapi-generator generate NestJS server code
- heyapi/openapi-ts generate frontend API client code (typescript, react-query, nextjs)
- oapi-codegen generate Go server code (gin, strict server)
- API/Contract change should come from API spec in `api` and `proto`, run `nx gen api` and `nx gen proto` to generate code

### Frontend

- Use shadcn component as much as possible, install in `packages/ui`, the package name is `@notopia-uit/ui`
  - Example: `shadcn add button -c @notopia-uit/ui`
- Use Tailwind CSS, try to avoid writing custom css and color
- If you wish to read nextjs documentation (new nextjs version, not the old you have been trained),
  read in `node_modules/next/dist/docs/`, vercel have bundle all documentation in next package, you can read it offline

### Go

- Do not run `go generate`, because we don't mange tool via go tool, but via mise. Stick with nx targets
- Wiring is google/wire: edit `wire.go`, run `nx wire {service}`, never hand-edit `wire_gen.go`.
  Each package exposes `var ProvideX = NewX`
- Services are DDD + CQRS: layering, command/query rules, domain vs app models, sqlc split are in
  [`docs/ai/GO_CLEAN_ARCH.md`](./docs/ai/GO_CLEAN_ARCH.md) (read before writing Go service code);
  concrete layout in [`internal/note/AGENTS.md`](./internal/note/AGENTS.md)
- Shared building blocks (config, `Cmd`/`Query`, gin, otel) are in [`pkg/`](./pkg/AGENTS.md)
- Go 1.26: `new(value)` / `new(fn(...))` returns a pointer; Go 1.27 supports generic methods on structs
- Prefer testify for testing, mockery if need interface mocking (list only necessary interfaces)
- SQLC custom plugin for dynamic query at runtime <https://raw.githubusercontent.com/vtuanjs/sqlc-gen-go/refs/heads/main/README.md>
- If a (pointer?) struct implement an interface, add e.g. `var _ Interface = (*Struct)(nil)`

### Messaging

- Redpanda (Kafka API). Go services publish through a Watermill **transactional outbox**, so
  domain writes and event publishes commit together; NestJS services publish/consume
  directly with `@nestjs/microservices`
- Topic naming: `events.integration.{service}.{aggregate}.{event}`
- Cross-service payload types are defined **once** in `api/share/`, never retyped per service

### Service-to-service

- Public traffic enters through Traefik, which validates the Authentik JWT and injects user
  headers. Go reads the user via `commonhttp.GatewayUserAuth()`; NestJS uses `UserGuard` +
  `@User()`. The `apps/document` websocket is the exception — it validates a JWT itself
- Internal calls are gRPC (`proto/`). Never call another service's public REST API internally

### Infrastructure

- Including Identity Provider (Authentik), API Gateway (Traefik), Object Storage (RustFS), and other services.
  Read [AGENTS.md](./deploy/compose/AGENTS.md) for details

## Nx

- Targets:
  - build
  - dev: run in development mode, continuous
  - start for nextjs style
  - serve: run the built
- Run `nx lint {projectName} --fix` to apply oxlint fix for those typescript projects
- Run `nx lint {projectsName}` for golangcilint for those go projects (especially in `cmd/` dir)
- Should run lint whenever changing code
- Run `nx` directly, not need to via `pnpm exec nx`

## Seed vs Fixtures

- **seed**: mandatory data the app needs in every environment, prod included (reference data,
  default plans, role→permission matrix). Must be idempotent. `just seed`
- **fixtures**: fake data for local development only (demo workspaces, notes, memberships,
  documents, search index content). `just fixtures` (runs seed first)
- Per-service recipes: `just fixtures-note|fixtures-document|fixtures-search`. New services
  follow the same naming: `seed` target for mandatory data, `fixtures` target for dev fake data

## General Rules

- Temp file must be go into `./tmp/{projectName}`, avoid writing to `/tmp/` when things need to be persisted
- While writing code, try **not to write unnecessary comments** into code
- There are many `*.env*` file, which contains safe local development environment variables. Only `*.env.local*` can contain sensitive environment variables, and should be gitignored
- Do not run `npx`, or `npx tsc`. Prefer using `pnpm dlx`, `pnpm exec`, `tsc`
- If something is not worth to code/fix, but might have issue in the future, then comments
- Prefer breakline for human readable, over one-line; code, naming convention over saving token because human still maintain the source code
- Do not swallow errors/exceptions in code, always log, or throw/return to caller
- Constant/enum Some over hard string (e.g. concurrency `VND`, `USD`, or others), easier for expansion, testing, type safety
- Maintainable over fast, technical debt

### Superpowers

- Skip unnecessary, obvious cases testing for faster, saving tokens
- Not need to write full code, can use `…`, skip imports, mapper code, or write comments to note expand the code more while writing;
  agents later can fill in the details, but should not skip important logic

### Git

Before commit, make sure to:

- Use conventional commit
- Should run `oxfmt {pathChanged}` after writing JS code
- Run `nx lint {projectName} --fix` to apply eslint/golangcilint fix
- If introduce deps, should run `go mod tidy` for go, and `pnpm install` for JS/TS
