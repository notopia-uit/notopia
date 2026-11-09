---
title: Getting Started
order: 1
---

# Getting Started

Local setup and the day-to-day development workflow. Everything below runs from the repo root.

## Prerequisites

Tools are managed with `mise`:

```bash
mise i            # install all tool versions
pnpm i            # install JS/TS dependencies
go mod download   # download Go dependencies
pnpm husky        # initialise hooks
```

## Infrastructure

```bash
just docker-up                    # core infrastructure (Postgres, Redis, Redpanda, Meilisearch, ...)
just docker-up-monitoring         # + Grafana, Prometheus, Loki, Tempo
compose --profile "dev-ui" up -d  # + DbGate, Redpanda Console
```

Everything is served under `*.notopia.localhost` through Traefik (`web`, `api`, `authentik`,
`traefik`, `rustfs`, `meilisearch`, `redpanda`, `dbgate`). Default credentials are
`notopiauit` / `notopiauit`.

## Run services

```bash
nx run note             # Go services run directly
nx run authorization
nx run document         # NestJS services build and watch
nx run search-worker
nx dev web              # Next.js
nx dev api-web          # Scalar API reference
```

## Code generation

API and contract changes start from the spec, never from generated code:

```bash
nx gen api     # OpenAPI: Go server, NestJS server, frontend client
nx gen proto   # Protobuf: Go, TypeScript, gRPC docs
```

## Seed vs fixtures

- **seed**: mandatory data the app needs in every environment, production included (reference
  data, default plans, the role to permission matrix). Must be idempotent. Run `just seed`.
- **fixtures**: fake data for local development only (demo workspaces, notes, memberships,
  documents, search index content). Run `just fixtures`, which runs seed first.

Per-service recipes exist as `just fixtures-note|fixtures-document|fixtures-search|...`. New
services follow the same naming: a `seed` target for mandatory data and a `fixtures` target
for dev fake data.

## Conventions

- Do not source `.env` files, Nx loads them.
- JS/TS: `oxlint`, `oxfmt`, `tsc`. Go: `golangci-lint` through `nx lint {project}`.
- Conventional commits; the pre-commit hook runs `nx affected` for typecheck, lint and format.
- Go service layout rules are in [Conventions](../conventions/).
