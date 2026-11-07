# K8s 00 — Kubernetes deployment overview

> Index for the Kubernetes deployment sub-projects. Shared decisions and contracts live here;
> each sub-project spec references this file instead of repeating it.
> Order: 01 → 02 → 03. 04 can run in parallel with 02.
> Progress is tracked by the task checkboxes in each plan under `docs/superpowers/plans/`.

| #   | Sub-project                                                        | Spec                                         |
| --- | ------------------------------------------------------------------ | -------------------------------------------- |
| 01  | App readiness: CLI, migrations, seeds, config, health, auth claims | `2026-10-24-k8s-01-app-readiness-design.md`  |
| 02  | Helm charts                                                        | `2026-10-24-k8s-02-helm-charts-design.md`    |
| 03  | Homelab cluster wiring (Flux)                                      | `2026-10-24-k8s-03-homelab-wiring-design.md` |
| 04  | CI and release                                                     | `2026-10-24-k8s-04-ci-release-design.md`     |

## Goal

Deploy every Notopia service end to end on the homelab Kubernetes cluster (Flux, Istio ambient,
Gateway API, CNPG, Redpanda, RustFS, Authentik). Deploys are triggered immediately by a push,
not by interval polling. The charts stay usable on clusters without Istio or Flux.

## Deployed workloads

| Workload        | Kind                  | Public route                 | Own CNPG                      | Own Valkey                           |
| --------------- | --------------------- | ---------------------------- | ----------------------------- | ------------------------------------ |
| `web`           | Next.js               | `notopia.${PUBLIC_DOMAIN}`   | no (better-auth is stateless) | no                                   |
| `note`          | Go                    | `notopia-api…/note`          | yes                           | yes                                  |
| `authorization` | Go (gRPC + health)    | `notopia-api…/authorization` | yes                           | no                                   |
| `billing`       | Go                    | `notopia-api…/billing`       | yes                           | yes                                  |
| `document`      | NestJS + Hocuspocus   | `notopia-api…/document`      | yes                           | yes, used only when scaling is on    |
| `search-worker` | NestJS Kafka consumer | health only                  | no                            | no (in-memory tombstones, 1 replica) |

`api-web` is not deployed. Every data store is **per service**: there are no shared Postgres or
Valkey instances.

Platform services already in homelab, used as is: Authentik, Redpanda
(`auto_create_topics_enabled: true`, so no topic management), RustFS, CNPG operator, Istio,
Alloy (OTLP at `alloy-alloy.monitoring.svc:4317`).

## Hostnames

| Host                                   | Target                                                                     |
| -------------------------------------- | -------------------------------------------------------------------------- |
| `notopia.${PUBLIC_DOMAIN}`             | web                                                                        |
| `notopia-api.${PUBLIC_DOMAIN}`         | API edge (path prefix per service)                                         |
| `notopia-meilisearch.${PUBLIC_DOMAIN}` | Meilisearch. The browser searches directly using note-issued tenant tokens |
| `authentik.${PUBLIC_DOMAIN}`           | existing Authentik                                                         |

Single-level names only, because the wildcard certificate covers `*.${PUBLIC_DOMAIN}`.

## Repository ownership

- **notopia repo, `deploy/k8s/`**: charts and everything Notopia-specific for a cluster
  (`clusters/homelab/`: CNPG clusters, Meilisearch, Valkey, SOPS secrets, HelmReleases, values).
- **homelab repo**: `notopia` namespace, a `GitRepository` pointing at this repo, the webhook
  `Receiver`, and three Flux `Kustomization`s pointing into `deploy/k8s/clusters/homelab/`.
- Charts never reference Flux, SOPS or homelab. Another cluster adds its own `clusters/<name>/`.

## Gateway header contract

Services do not verify JWTs (except the document websocket). They trust these headers, set only
by the gateway:

| Header               | Source claim  | Format                                            |
| -------------------- | ------------- | ------------------------------------------------- |
| `X-Forwarded-ID`     | `sub`         | string                                            |
| `X-Forwarded-Email`  | `email`       | string                                            |
| `X-Forwarded-Name`   | `name`        | string                                            |
| `X-Forwarded-Groups` | `groups_json` | JSON array string, e.g. `["admin"]`               |
| `X-Forwarded-Roles`  | —             | reserved, not set (no `roles` claim exists today) |

`groups_json` is a string claim emitted by a custom Authentik scope mapping (sub-project 01).
That way Istio `outputClaimToHeaders` (string, int and bool only) and Traefik `headerMap` produce
byte-identical headers. The list claim `groups` stays in the token for gateway policy checks
(Istio `request.auth.claims[groups]`, Traefik `require`).

Every gateway strips client-supplied `X-Forwarded-*` headers before authentication: Traefik with
`removeMissingHeaders: true`, Istio with a `RequestHeaderModifier` on the front route.

## Auth at the edge

| Path                               | Rule                                                      |
| ---------------------------------- | --------------------------------------------------------- |
| `/{svc}/health/*`                  | public                                                    |
| `/billing/webhooks/stripe`         | public                                                    |
| `GET /billing/plans`               | public                                                    |
| `/document/ws/document`            | public at the gateway (the service validates its own JWT) |
| `/billing/admin/*`                 | valid JWT **and** `groups` contains the admin group       |
| everything else under the API host | valid JWT                                                 |

The admin rule is defense in depth; billing keeps `HasGroup(AdminGroup)`.

## Deploy flow

```
release workflow → build + push ghcr image(s)
                 → yq set image.tag in deploy/k8s/clusters/homelab/apps/<svc>/values.yaml
                 → commit to main
GitHub push webhook → Flux Receiver → reconcile GitRepository notopia → HelmRelease upgrade
HelmRelease upgrade: pre-upgrade migrate Job → rollout (wait) → post-upgrade seed Job
```

Release stays manual (workflow_dispatch) for now. The workflow is shaped so that adding
`push: main` with Nx affected detection switches it to automatic.

## Out of scope

- CNPG backups (no barman / R2 for notopia clusters)
- Automated Authentik and RustFS provisioning (manual checklist in 03)
- A `gatewayauth` ext_authz service, Wasm, and a VirtualService dynamic-metadata approach
- Deploying `api-web`
- Flux image automation (tags are bumped by the release workflow; Renovate stays for third-party
  versions)
