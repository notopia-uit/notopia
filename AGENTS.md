# Monorepo

Manged with Nx. Package manager: pnpm (use ./pnpm-workspace.yaml), go (single go.mod for all packages), mise, buf

## Structure

```
api/                            # OpenAPI spec
proto/                          # Protobuf definitions
apps/                           # JS/TS applications
  document/                     # Document service (NestJS) (nx: document)
  searchworker/                 # Search worker (NestJS) sync to meilisearch (nx: searchworker)
  web/                          # NextJS (nx: web)
  api-web/                      # OpenAPI Spec rendered with Scalar React (nx: api-web)
packages/                       # JS packages
  api-document-nestjs-server/   # NestJS server codegen from OpenAPI spec, using openapi-generator
  api-gen/                      # Frontend API client codegen from OpenAPI spec, using heyapi/openapi-ts
  lib/                          # Share code, blocknote model,...
  pb/                           # Protobuf generated code, use grpc, protovalidate
  ui/                           # React UI components, using shadcn/ui
docs/                           # Vitepress Documentation (class, sequence, architecture diagrams, database, etc.)
cmd/                            # Go services
  note/                         # Note service (nx: note)
  authorization/                # Authorization service (nx: authorization)
internal/                       # Internal Go packages
  note/
  authorization/
pkg/                            # Go packages
  api                           # Code gen from oapi-codegen
  common                        # Including commonhttp, commonconfig
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

## General Rules

- Temp file must be go into `./tmp/{projectName}`, avoid writing to `/tmp/` when things need to be persisted
- While writing code, try to not write unnecessary comment into code
- There are many `*.env*` file, which contains safe local development environment variables. Only `*.env.local*` can contain sensitive environment variables, and should be gitignored
- Do not run `npx`, or `npx tsc`. Prefer using `pnpm dlx`, `pnpm exec`, `tsgo` (replace for `tsc`)

### Git

Before commit, make sure to:

- Use conventional commit
- Should run `oxfmt {pathChanged}` after writing JS code
- Run `nx lint {projectName} --fix` to apply eslint/golangcilint fix
- If introduce deps, should run `go mod tidy` for go, and `pnpm install` for JS/TS
