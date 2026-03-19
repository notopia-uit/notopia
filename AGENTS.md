# Monorepo

Manged with Nx. Package manager: pnpm (use ./pnpm-workspace.yaml), go (single go.mod for all packages), mise, buf

## Structure

```
api/                            # OpenAPI spec
proto/                          # Protobuf definitions
apps/                           # JS/TS applications
  web/                          # NextJS
  document/                     # Document service (NestJS)
packages/                       # JS packages
  api-document-nestjs-server    # NestJS server codegen from OpenAPI spec, using openapi-generator
  api-gen                       # Frontend API client codegen from OpenAPI spec, using heyapi/openapi-ts
  lib                           # Share code, blocknote model,...
  pb                            # Protobuf generated code, use connectrpc
  ui                            # React UI components, using shadcn/ui
docs/                           # Vitepress Documentation (class, sequence, architecture diagrams, database, etc.)
cmd/                            # Go services
  note/                         # Note service
internal/                       # Internal Go packages
  api                           # Code gen from oapi-codegen
  common
  healthmanager
  helper
  logging
  metadata
  otel
  pb
pkg/                            # Go packages
```

## Technologies

### API

- Redocly to bundle and join yaml spec to json
- Scalar render joined spec, mock server
- openapi-generator generate NestJS server code
- heyapi/openapi-ts generate frontend API client code (typescript, react-query, nextjs)
- oapi-codegen generate Go server code (gin, strict server)

### Frontend

- Use shadcn component as much as possible, install in `packages/ui`, the package name is `@notopia-uit/ui`
  - Example: `shadcn add button -c @notopia-uit/ui`
- Use Tailwind CSS, try to avoid writing custom css and color

## Nx

- Targets:
  - build
  - dev: run in development mode, continuous
  - start for nextjs style
  - serve: run the built
- Run `nx lint {projectName} --fix` to apply eslint fix for those typescript projects

## General Rules

- Temp file must be go into `./tmp/{projectName}`
- While writing code, try to not write unnecessary comment into code
