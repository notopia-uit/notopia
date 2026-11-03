# K8s 02 — Helm charts

> Sub-project 2 of 4. See `2026-10-24-k8s-00-overview-design.md`.
> Depends on 01 (commands, health endpoints, env names). Blocks 03.

## Goal

One Helm chart per service and one for the API edge, sharing helpers through a library chart.
The charts follow Helm conventions, have generated schemas and docs, and are unit-tested. They
work without Istio or Flux.

## Layout

```
deploy/k8s/charts/
  notopia-common/          # type: library — named templates only, no resources
  web/
  note/
  authorization/
  billing/
  document/
  search-worker/
  notopia-edge/            # API edge: Gateway, front route, JWT authn, edge authz, telemetry
```

Each chart has this shape:

```
<chart>/
  Chart.yaml               # apiVersion v2, version (semver), appVersion unused, dependency on notopia-common
  values.yaml              # every key documented with "# --" and "# @schema"
  values.schema.json       # generated (helm-schema)
  README.md                # generated (helm-docs)
  templates/
    _helpers.tpl           # chart-local only; shared helpers come from notopia-common
    NOTES.txt
    …
  tests/                   # helm-unittest
```

- The dependency is `notopia-common` with `repository: file://../notopia-common`. Flux builds
  charts from the `GitRepository` and resolves relative `file://` dependencies.
  `deploy/k8s/charts/*/charts/` is gitignored.
- **Version rule**: any change to a chart (or to `notopia-common`) bumps that chart's `version`.
  A `notopia-common` change bumps every dependent chart. CI enforces this (sub-project 04),
  because Flux only re-packages a Git chart when `version` changes (default
  `reconcileStrategy: ChartVersion`).

## `notopia-common` (library)

Named templates:

| Template                                                         | Purpose                                                                                                                                                                                                                           |
| ---------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `notopia-common.name`, `.fullname`, `.chart`                     | standard naming, with `nameOverride` / `fullnameOverride`                                                                                                                                                                         |
| `notopia-common.labels`, `.selectorLabels`                       | `app.kubernetes.io/{name,instance,version,component,part-of=notopia,managed-by}`, `helm.sh/chart`                                                                                                                                 |
| `notopia-common.serviceAccountName`                              | honors `serviceAccount.create` and `serviceAccount.name`                                                                                                                                                                          |
| `notopia-common.image`                                           | `repository:tag`, or `@digest` when `image.digest` is set                                                                                                                                                                         |
| `notopia-common.env`                                             | renders `env` + `extraEnv` + OTel env + memory env, shared by the Deployment and hook Jobs                                                                                                                                        |
| `notopia-common.otelEnv`                                         | `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `OTEL_RESOURCE_ATTRIBUTES` (`k8s.namespace.name`, `k8s.pod.name`, `k8s.node.name` via the downward API, merged with `otel.resourceAttributes`) |
| `notopia-common.memoryEnv`                                       | `GOMEMLIMIT` (`runtime: go`, 90% of the memory limit) or `NODE_OPTIONS=--max-old-space-size` (`runtime: node`, 75%)                                                                                                               |
| `notopia-common.hookJob`                                         | Job for `migrate` / `seed`: same image, SA, env and resources as the Deployment; `backoffLimit: 0`; `activeDeadlineSeconds`; hook annotations                                                                                     |
| `notopia-common.podSecurityContext`, `.containerSecurityContext` | `runAsNonRoot`, `readOnlyRootFilesystem`, `drop: [ALL]`, `seccompProfile: RuntimeDefault`                                                                                                                                         |
| `notopia-common.istioWorkloadPolicy`                             | L4 `AuthorizationPolicy` (ztunnel): allow only the listed principals to the listed ports                                                                                                                                          |

**Environment comes from values, not from a chart-owned ConfigMap.** Pre-install hook Jobs run
before the release's own resources exist, so a release ConfigMap wouldn't be there yet. Plain env
is rendered inline, and secrets are referenced through `envFrom` / `valueFrom` from Secrets that
exist outside the release (SOPS and CNPG). A `checksum/env` pod annotation restarts pods when the
env values change.

## Service charts

Templates each service chart renders:

| Template                         | Rendered when                                                                   |
| -------------------------------- | ------------------------------------------------------------------------------- |
| `deployment.yaml`                | always                                                                          |
| `service.yaml`                   | always. Ports have `appProtocol` (`http`, `kubernetes.io/h2c` for gRPC)         |
| `serviceaccount.yaml`            | `serviceAccount.create` (default `true`, `automountServiceAccountToken: false`) |
| `httproute.yaml`                 | `route.enabled`                                                                 |
| `pdb.yaml`                       | `podDisruptionBudget.enabled`                                                   |
| `hpa.yaml`                       | `autoscaling.enabled`                                                           |
| `job-migrate.yaml`               | the chart has a migrate command (note, billing, document)                       |
| `job-seed.yaml`                  | the chart has a seed command (authorization, billing, search-worker)            |
| `istio/authorizationpolicy.yaml` | `istio.enabled`                                                                 |

Per-chart specifics live in templates, not values. Commands, ports and probe paths are fixed
by the service code from 01:

| Chart         | Container command         | Ports                               | Probes                                        | Hooks                       |
| ------------- | ------------------------- | ----------------------------------- | --------------------------------------------- | --------------------------- |
| note          | `serve`                   | http 8081, grpc 18081, health 28081 | `/note/health/{startup,ready,live}` on health | migrate: `migrate up`       |
| authorization | `serve`                   | grpc 18089, health 28089            | `/authorization/health/…`                     | seed: `seed`                |
| billing       | `serve`                   | http 8084, grpc 18084, health 28084 | `/billing/health/…`                           | migrate + seed              |
| document      | `node main.cjs`           | http 8082                           | `/document/health/{startup,ready,live}`       | migrate: `node migrate.cjs` |
| search-worker | `node main.cjs`           | health 28083                        | `/search-worker/health/{ready,live}`          | seed: `node seed.cjs`       |
| web           | `node apps/web/server.js` | http 3000                           | TCP socket on 3000 (no health route in web)   | —                           |

Ports and health paths are read from each service's config defaults. 01 verifies the exact values
before templating.

### Hook order

| Hook        | Annotation                  | Weight | Delete policy                                              |
| ----------- | --------------------------- | ------ | ---------------------------------------------------------- |
| migrate Job | `pre-install,pre-upgrade`   | `-5`   | `before-hook-creation` (the last run stays for inspection) |
| seed Job    | `post-install,post-upgrade` | `5`    | `before-hook-creation`                                     |

A failed hook fails the release. Migrations must be **expand/contract** (old pods keep serving
during the rollout). Seeds are idempotent and run on every upgrade.

### Scaling

Every chart supports scaling, and every default is `replicaCount: 1`, `autoscaling.enabled: false`.

- `document`: `scaling.enabled` sets `NOTOPIA_DOCUMENT_SCALING_ENABLED=true` and requires
  `scaling.redisUrlSecret` (`name`, `key`). The template `fail`s if `replicaCount > 1` or
  autoscaling is on without it.
- `search-worker`: `replicaCount > 1` requires `redis.urlSecret`. Without it, tombstones are in
  memory (single instance only), so the template `fail`s.
- `note`, `billing`: Redis is always required (`redis.addr`, `redis.passwordSecret`, `redis.db`),
  and they are scale-safe already (Redis cron lock, outbox).

### Values shape (example: note)

```yaml
# -- Override the chart name
nameOverride: ''
fullnameOverride: ''

image:
  # -- Image repository
  repository: ghcr.io/notopia-uit/note
  # -- Image tag (set by the release workflow)
  tag: ''
  # -- Image digest; takes precedence over tag
  digest: ''
  pullPolicy: IfNotPresent
imagePullSecrets: []

replicaCount: 1
autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 3
  targetCPUUtilizationPercentage: 80

serviceAccount:
  create: true
  name: ''
  annotations: {}

resources:
  requests: { cpu: 50m, memory: 64Mi }
  limits: { memory: 256Mi } # no CPU limit by convention

# -- Plain env (rendered inline; shared with hook Jobs)
env: {}
# -- Extra env entries (valueFrom etc.)
extraEnv: []
# -- Secrets / ConfigMaps that exist outside this release
envFrom: []

database:
  # -- Secret holding the Postgres URI (CNPG "<cluster>-app", key "uri")
  urlSecret: { name: '', key: uri }
redis:
  addr: ''
  db: 0
  passwordSecret: { name: '', key: password }

route:
  enabled: true
  parentRefs: [] # e.g. the notopia-edge Gateway
  hostnames: []
  # -- Health is routed publicly under /<svc>/health
  exposeHealth: true

otel:
  enabled: true
  endpoint: ''
  protocol: grpc
  resourceAttributes: {}

hooks:
  migrate:
    activeDeadlineSeconds: 300
    resources: {} # defaults to .Values.resources

istio:
  enabled: true
  # -- SPIFFE principals allowed to reach this workload
  allowedPrincipals: [] # edge gateway SA + gRPC callers

podDisruptionBudget: { enabled: false, minAvailable: 1 }
podAnnotations: {}
podLabels: {}
nodeSelector: {}
tolerations: []
affinity: {}
extraVolumes: []
extraVolumeMounts: []
```

## `notopia-edge`

```yaml
gateway:
  enabled: true # false → routes attach to an existing gateway
  className: istio
  annotations: {} # homelab: networking.istio.io/service-type: ClusterIP
  hostname: '' # notopia-api.<domain>
  listenerPort: 80

frontRoute: # chain from a shared front gateway (homelab: default Gateway)
  enabled: true
  parentRefs: []
  hostnames: []
  stripHeaders:
    - X-Forwarded-ID
    - X-Forwarded-Email
    - X-Forwarded-Name
    - X-Forwarded-Groups
    - X-Forwarded-Roles

istio:
  enabled: true
  jwt:
    issuer: '' # https://authentik.<domain>/
    jwksUri: '' # in-cluster Authentik app-web JWKS
    claimHeaders:
      X-Forwarded-ID: sub
      X-Forwarded-Email: email
      X-Forwarded-Name: name
      X-Forwarded-Groups: groups_json
  publicRules: [] # { paths: [...], methods: [...] }
  groupRules: [] # { paths: [...], groups: [...] }
  telemetry:
    enabled: false
    tracingProviders: [] # homelab: otel-trace
    accessLogProviders: [] # homelab: otel-log
    randomSamplingPercentage: 100
```

Templates:

| Template                           | Rendered when                              | Notes                                                                                                                                                  |
| ---------------------------------- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `gateway.yaml`                     | `gateway.enabled`                          | one HTTP listener, `allowedRoutes` from the release namespace                                                                                          |
| `front-httproute.yaml`             | `frontRoute.enabled`                       | backend `{{fullname}}-istio:80` (the Istio-generated Service); `RequestHeaderModifier.remove`                                                          |
| `istio/requestauthentication.yaml` | `istio.enabled`                            | `targetRefs: Gateway {{fullname}}`; `outputClaimToHeaders` from `claimHeaders`                                                                         |
| `istio/authorizationpolicy.yaml`   | `istio.enabled`                            | ALLOW: public rules; group rules with `requestPrincipals: ["*"]` + `when request.auth.claims[groups]`; everything else with `requestPrincipals: ["*"]` |
| `istio/telemetry.yaml`             | `istio.enabled && istio.telemetry.enabled` |                                                                                                                                                        |

The Gateway's service account (Istio names it `{{fullname}}-istio`) is the principal that the
service charts' `istio.allowedPrincipals` admits.

Without Istio (`istio.enabled: false`), the chart renders only the Gateway API objects. The
cluster must then supply its own JWT → header layer that honors the header contract in 00.
`NOTES.txt` says this.

## Tooling (mise: `mise.toml` and `.mise.ci.toml`)

| Tool                          | Use                                                                                                                           |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `helm`                        | lint, template, dependency build                                                                                              |
| `github:dadav/helm-schema`    | `values.schema.json` from `# @schema` blocks; `-p` helm-docs compatibility; `--check` in CI                                   |
| `github:norwoodj/helm-docs`   | chart `README.md` from `# --` comments                                                                                        |
| `github:yannh/kubeconform`    | validates `helm template` output (`-strict`, with CRD schemas for Gateway API, Istio and CNPG from the datreeio CRDs-catalog) |
| `helm-unittest` (helm plugin) | `tests/*_test.yaml`                                                                                                           |

Exact versions are pinned in the plan's first task.

Each chart gets an nx project (`deploy/k8s/charts/<chart>/project.json`) with these targets:

- `gen`: helm-schema + helm-docs (it is part of the existing CI `gen` + "uncommitted changes" check)
- `lint`: `helm dependency build` → `helm lint --strict` → `helm template | kubeconform`
- `test`: `helm unittest`

## Testing (helm-unittest, conditional paths only)

- Hooks: migrate and seed Jobs render with the right annotations, weights and command, and are
  absent for charts without them.
- `istio.enabled: false` renders no `security.istio.io` objects in any chart.
- `route.exposeHealth`: a health rule is present or absent, pointing at the health port.
- Scaling guards: document / search-worker `fail` with `replicaCount: 2` and no Redis.
- `notopia-common.memoryEnv`: `GOMEMLIMIT` for a `256Mi` limit.
- Edge: `frontRoute` strips every contract header; `groupRules` render a `when` on
  `request.auth.claims[groups]`.
