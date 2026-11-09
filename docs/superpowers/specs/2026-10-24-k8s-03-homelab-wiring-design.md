# K8s 03 — Homelab cluster wiring (Flux)

> Sub-project 3 of 4. See `2026-10-24-k8s-00-overview-design.md`.
> Depends on 01 and 02. Touches two repos: this one (`deploy/k8s/clusters/homelab/`) and
> `~/projects/homelab`.

## Goal

Flux deploys Notopia from this repo. A push to `main` reconciles within seconds through the
existing webhook Receiver. All Notopia-specific state lives in this repo. Homelab only gets a
pointer.

## Homelab conventions mirrored (verified)

- An app is a Flux `Kustomization` in `flux-system` → `path` + `decryption: sops/sops-age` +
  `postBuild.substituteFrom` (`clustersetting` Secret and ConfigMap, providing `${PUBLIC_DOMAIN}`,
  `${LOCAL_DOMAIN}`, `${TIMEZONE}`).
- HelmRelease values: `values.yaml` → `configMapGenerator` → `valuesFrom`, with a
  `kustomizeconfig.yaml` `nameReference` so a values change renames the ConfigMap and upgrades the
  release.
- CNPG: `ClusterImageCatalog postgresql-minimal-trixie`, major 18; `disabledDefaultServices: [ro, r]`.
- Front gateway: `Gateway default` in `istio-system`, listeners `https-wildcard`
  (`*.${PUBLIC_DOMAIN}`) and `https-wildcard-local` (`*.${LOCAL_DOMAIN}`).
- Webhook: `Receiver webhook` (type `github`) behind `flux-webhook.${PUBLIC_DOMAIN}/hook/` →
  `webhook-receiver`. A second Receiver needs no new route.
- Alerts: the Telegram `Alert` already watches `HelmRelease: "*"` and `Kustomization: "*"`.
- SOPS: age recipient `age1ytym5g6mxf0974syp2g2ecrh0y74tnl0u9ee4xm8caawhyv25pls3f2sy4`.

## Homelab repo changes

```
kubernetes/cluster/namespaces.yaml        + Namespace notopia (label istio.io/dataplane-mode: ambient)
kubernetes/apps/notopia/
  kustomization.yaml
  repos/
    kustomization.yaml
    notopia.gitrepository.yaml           # https://github.com/notopia-uit/notopia, branch main, interval 12h
                                         # spec.ignore: "/*\n!/deploy/k8s/"
  webhook/
    kustomization.yaml
    receiver.yaml                         # Receiver notopia, type github, events [ping, push] → GitRepository notopia
    token.secret.sops.yaml
  notopia-infra.yaml                      # Flux Kustomization → deploy/k8s/clusters/homelab/infra
                                          #   sourceRef GitRepository notopia; dependsOn cnpg-system-cnpg
  notopia-edge.yaml                       # → …/edge;  dependsOn notopia-infra
  notopia-apps.yaml                       # → …/apps;  dependsOn notopia-infra, notopia-edge
```

- The three notopia Kustomizations use the same `decryption` and `postBuild` blocks as other apps.
  `notopia-infra` and `notopia-edge` use `wait: true`. `notopia-apps` uses `wait: false` (each
  HelmRelease reports its own status, so one failing service doesn't block the others' readiness).
- Interval `12h` (like the rest of homelab). The Receiver makes deploys immediate. Pushes to other
  branches or outside `deploy/k8s/` produce no change, so the reconcile is a no-op.
- The GitHub webhook on `notopia-uit/notopia` (push events, JSON, the Receiver token) is added by hand.
- Whether the ambient namespace label is how this cluster enrolls namespaces is verified first;
  the existing namespaces carry no label, so enrollment may come from elsewhere.

## This repo: `deploy/k8s/clusters/homelab/`

```
.sops.yaml (repo root)                    # rule: deploy/k8s/clusters/.*\.sops\.ya?ml → data|stringData, homelab age recipient
deploy/k8s/clusters/homelab/
  infra/
    kustomization.yaml                    # namespace: notopia
    cnpg/
      note.cluster.yaml  billing.cluster.yaml  authorization.cluster.yaml  document.cluster.yaml
    valkey/
      repos/valkey.helmrepository.yaml
      note/       helmrelease.yaml values.yaml kustomization.yaml kustomizeconfig.yaml
      billing/    …
      document/   …
    meilisearch/
      repos/meilisearch.helmrepository.yaml
      helmrelease.yaml values.yaml httproute.yaml kustomization.yaml kustomizeconfig.yaml
    secrets/
      kustomization.yaml
      note.secret.sops.yaml               # meilisearch search key (from runbook step), …
      billing.secret.sops.yaml            # stripe secret key, webhook secret
      document.secret.sops.yaml           # rustfs access key / secret
      search-worker.secret.sops.yaml      # meilisearch master key
      web.secret.sops.yaml                # BETTER_AUTH_SECRET, AUTHENTIK_CLIENT_SECRET
      meilisearch.secret.sops.yaml        # MEILI_MASTER_KEY
      valkey-{note,billing,document}.secret.sops.yaml
  edge/
    kustomization.yaml kustomizeconfig.yaml helmrelease.yaml values.yaml
  apps/
    kustomization.yaml kustomizeconfig.yaml
    web/ note/ authorization/ billing/ document/ search-worker/
      helmrelease.yaml values.yaml
```

### CNPG (one cluster per service, no backups)

```yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: cnpg-note
spec:
  instances: 1
  imageCatalogRef:
    {
      apiGroup: postgresql.cnpg.io,
      kind: ClusterImageCatalog,
      name: postgresql-minimal-trixie,
      major: 18,
    }
  storage: { size: 1Gi }
  postgresql:
    parameters: { timezone: $, { TIMEZONE }, log_timezone: $, { TIMEZONE } }
  bootstrap:
    initdb: { database: note, owner: note }
  managed:
    services:
      disabledDefaultServices: [ro, r]
```

The `cnpg-<svc>-app` Secret (generated by CNPG, key `uri`) feeds each chart's
`database.urlSecret`. Document reads discrete fields (`NOTOPIA_DOCUMENT_DB_HOST`, …), mapped from
the same Secret's `host`, `port`, `username`, `password` and `dbname` keys.

### Valkey (one per service: note, billing, document)

- The official Valkey chart (`valkey-io/valkey-helm`): standalone, persistence on, `maxmemory-policy
  noeviction`, `appendonly yes` (matching the compose services), and a password from
  `valkey-<svc>` Secret. The chart and its value keys are verified when this is implemented.
- Document's instance exists so that `scaling.enabled` can be switched on by values alone. The
  default `replicaCount: 1` and `scaling.enabled: false` mean it is unused at first.

### Meilisearch

- `meilisearch/meilisearch-kubernetes` chart, `environment: production`, master key from the
  `meilisearch` Secret, persistence on.
- `httproute.yaml`: host `notopia-meilisearch.${PUBLIC_DOMAIN}`, parent `default` gateway
  (`https-wildcard`), with **no** notopia edge auth (tenant tokens are the auth). CORS allows
  `https://notopia.${PUBLIC_DOMAIN}`.

### Edge values (homelab)

```yaml
gateway:
  className: istio
  annotations: { networking.istio.io/service-type: ClusterIP }
  hostname: notopia-api.${PUBLIC_DOMAIN}
frontRoute:
  parentRefs:
    - { name: default, namespace: istio-system, kind: Gateway, sectionName: https-wildcard }
  hostnames: [notopia-api.${PUBLIC_DOMAIN}]
istio:
  jwt:
    issuer: https://authentik.${PUBLIC_DOMAIN}/
    jwksUri: http://authentik-server.authentik/application/o/app-web/jwks/
  publicRules:
    - paths: ["/*/health/*"]
    - paths: [/billing/webhooks/stripe]
    - paths: [/billing/plans], methods: [GET]
    - paths: ["/document/ws/document*"]
  groupRules:
    - paths: ["/billing/admin/*"], groups: [notopia-admins]
  telemetry:
    enabled: true
    tracingProviders: [otel-trace]
    accessLogProviders: [otel-log]
```

The issuer form (domain + `/`, without the app slug) follows egolia's verified fix
(`fix(istio): issuer from authentik is only domain`).

### App HelmRelease (example: note)

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: note
spec:
  interval: 12h
  timeout: 10m
  chart:
    spec:
      chart: ./deploy/k8s/charts/note
      sourceRef: { kind: GitRepository, name: notopia, namespace: flux-system }
  install:
    remediation: { retries: 0 }
  upgrade:
    remediation: { retries: 0, remediateLastFailure: true, strategy: rollback }
  valuesFrom:
    - kind: ConfigMap
      name: note-helm-values
```

The `remediation` settings mean a failed migrate hook or rollout rolls back instead of retrying a
half-applied migration.

The cluster `values.yaml` sets: `image.tag` (written by the release workflow),
`route.parentRefs` → `notopia-edge` Gateway, `otel.endpoint: http://alloy-alloy.monitoring.svc:4317`,
`database.urlSecret.name: cnpg-note-app`, `redis.*`, `envFrom: [{ secretRef: { name: note } }]`, plain
env (`NOTOPIA_NOTE_SERVER_URL`, service gRPC addresses, Kafka brokers
taken from the homelab Redpanda Service when the values are written, Authentik), and
`istio.allowedPrincipals` (edge SA + gRPC callers).

gRPC callers (for `allowedPrincipals`). The exact callers are taken from the `Services` config of
each service when the values are written:

| Target        | Callers                 |
| ------------- | ----------------------- |
| authorization | note, document, billing |
| billing       | note, document          |
| note          | document                |

Web's route attaches directly to the `default` gateway (`notopia.${PUBLIC_DOMAIN}` and
`notopia.${LOCAL_DOMAIN}`), with no edge auth. Its env sets `API_URL`, `DOCUMENT_WS_URL`
(`wss://`), `MEILISEARCH_HOST`, the Authentik discovery URL and the redirect URI.

## Manual runbook (`deploy/k8s/clusters/homelab/README.md`)

Run once, in order:

1. **Authentik** (production, by hand):
   - OAuth2 provider + application `app-web` (confidential), redirect
     `https://notopia.${PUBLIC_DOMAIN}/api/auth/callback/authentik`
   - scope mapping `notopia` (expression from 01 §6), attached to the provider
   - group `notopia-admins`
   - service-account users for the cron actors (`Identity.ServiceAccountUserID`)
   - put the client ID and secret into `web.secret.sops.yaml`
2. **RustFS**: bucket for document storage + an access key scoped to it →
   `document.secret.sops.yaml`.
3. **Stripe**: secret key + webhook endpoint `https://notopia-api.${PUBLIC_DOMAIN}/billing/webhooks/stripe` →
   `billing.secret.sops.yaml`.
4. **Meilisearch search key bootstrap**: the first `search-worker` install runs its seed, which
   creates the key with the configured UID. Read the value with
   `curl -H "Authorization: Bearer $MASTER" https://notopia-meilisearch…/keys/<uid>`, put it into
   `note.secret.sops.yaml`, and commit. note becomes ready after that reconcile. The key is
   derived from the master key + UID, so it stays stable unless the master key changes.
5. GitHub webhook on the notopia repo → `https://flux-webhook.${PUBLIC_DOMAIN}/hook/<receiver path>`.

## Verification

- `flux get kustomizations | grep notopia`, then `flux get helmreleases -n notopia`: all Ready.
- `kubectl -n notopia get jobs`: the latest migrate and seed Jobs Completed.
- `curl https://notopia-api…/note/health/ready`: 200, no token needed.
- `curl https://notopia-api…/note/...` without a token: 403. With a forged `X-Forwarded-Groups`
  and no token, on `/billing/admin/...`: 403.
- Log in on `https://notopia.${PUBLIC_DOMAIN}`, open a note (websocket over `wss`), and search
  (Meilisearch through tenant token).
- Tempo: a trace crosses `notopia-edge` → note → authorization. A migrate Job span is visible
  under `service.name=note`.
- Kite and Authentik still work (shared gateway untouched).
