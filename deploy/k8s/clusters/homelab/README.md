# Homelab infrastructure provisioning

This is an operator runbook, not authorization to deploy. The homelab Flux
Kustomizations, Receiver, and every HelmRelease are checked in suspended; the
Receiver token is absent. No live verification or deployment has been performed
by this runbook. Obtain explicit authorization before cluster access, credential
creation, or deployment.

The Valkey and Meilisearch HelmReleases are suspended because no credentials have
been supplied. Do not resume a release until its encrypted Secret is committed,
included in `infra/secrets/kustomization.yaml`, and Flux has the matching age
private key. Application releases that depend on missing credentials must also
remain suspended. All application and edge releases are also suspended.
Do not expose the edge until the independent identity-header boundary checks in
`deploy/k8s/AGENTS.md` have passed.

## Application release prerequisites

The empty `image.tag` values are intentionally not deployable image selections:
the chart falls back to its default app version, which has not been verified.
Publish or independently verify an immutable image for each service and set its
tag before resuming that release. Verify image access and migration compatibility.
No chart-compatible application images or credentials have been verified.

Anonymous GHCR inspection found `2603.07.1311` for web, note, and document,
but those images were built from revision `5e62e887` before the current chart
contracts. In particular that revision's note executable does not implement the
chart's migration command, and its document image does not contain the current
migration entrypoint. Do not use those old images merely because their tags
exist. Authorization exposed only a build-cache tag; anonymous access to billing
and search-worker was denied. Publish current compatible images or verify
authorized registry access and compatibility before setting each tag.

The Kafka broker is `redpanda.redpanda.svc.cluster.local:9093`, derived by
rendering the homelab's pinned Redpanda chart 26.2.5 with its checked-in values.
The rendered headless Service and internal broker listener use that port. This
is not a live-cluster check: verify the deployed Service, advertised listeners,
connectivity, and configuration drift before resuming any Kafka client.
Also verify the Istio-generated `default-istio` and `notopia-edge-istio` service
account names, ambient enrollment and mTLS, Valkey and Meilisearch service names,
and service-to-service health checks against the actual deployment.

Set the Authentik client ID, service-account user IDs, RustFS endpoint, region
and bucket, and other production settings from the provisioned systems. Include
the note Authentik token in its Secret. Use the exact environment names from the
services, not local development credentials. Document's database fields are
provided from the CNPG Secret, and document must remain at one replica.

Authorization requires external policy seeding from the same image before every
install, upgrade, and rollback; follow `charts/authorization/README.md`.
Authorization upgrades are blocked until the controls in
[`Authorization upgrade and recovery`](#authorization-upgrade-and-recovery)
are verified; ordinary automatic remediation is incompatible with its policy seed.
Do not set `security.externalBoundaryAcknowledged` until the independent edge
boundary has been implemented and verified as described in the edge chart README.
Suspending a HelmRelease does not remove an already deployed release.

For search-key bootstrap, provision and resume Meilisearch and its credentials,
then configure and resume search-worker independently of note. Once its seed
has succeeded, provision note's search key and other prerequisites before
resuming note. Do not make the worker depend on note readiness. Flux app
reconciliation must not wait for all releases before the worker can run.

## Secrets

Create production credentials through the relevant provider or a secure local
process. Never put plaintext in Git, shell history, command-line arguments, or
unprotected temporary files. The Secret manifests in `infra/secrets/` are SOPS
encrypted with the root `.sops.yaml` rule, but contain only the placeholder
`REPLACE_BEFORE_USE`, not credentials. They are intentionally absent from the
Kustomization: encryption alone does not make a Secret ready to apply. CNPG
manages its own Secrets and has no manifest here.

### Provisioning with Neovim

For each Notopia Secret, open its existing file from the repository root
(example for Valkey note):

```sh
nvim deploy/k8s/clusters/homelab/infra/secrets/valkey-note.secret.sops.yaml
```

1. Enable and trust this repository's `.nvim.lua` in Neovim. It uses the age
   identity at `~/projects/homelab/secret/age.agekey`; verify that identity
   matches the recipient in `.sops.yaml` before editing.
   Use a trusted machine and terminal; disable terminal recording, clipboard
   managers and other capture, and protect against memory/swap exposure. Do not
   write, autosave, or copy a plaintext file at any point.
2. Use `\d` to decrypt in the buffer. Replace every placeholder in `stringData`
   with the actual credential as a YAML string, using proper quoting. The
   worker's key is the Meilisearch master
   key; note's search key is obtained only after the worker seed succeeds as
   described below. Provision note's Secret only when both values are ready.
3. Run `:write`. The repository's Neovim configuration encrypts the buffer before
   writing this matching filename; if encryption fails it must not write the file.
   Use `\e` to replace the plaintext buffer with the encrypted version, or `\d`
   to decrypt an existing encrypted file in the buffer. Check the on-disk file
   contains SOPS metadata and `ENC[...]` values under `stringData` and that its
   recipient matches `.sops.yaml`. Verify decryption with the actual Flux identity
   without printing plaintext; structural checks alone are insufficient. Never
   write the plaintext buffer to a different filename or expose it to untrusted
   editor plugins, clipboard managers, or language servers.
4. Confirm all placeholders are replaced without disclosing values, review and
   commit only the encrypted file, then add its filename under `resources` in
   `infra/secrets/kustomization.yaml`. Keep affected releases suspended until
   their credentials and all other prerequisites are ready.

For the Receiver, `receiver.secret.sops.yaml` is an encrypted placeholder
reference for `notopia-webhook-token` in `flux-system`, not a resource for this
repository's infra Kustomization. Provision it using a secure SOPS editing
workflow **in the homelab repository**, using its destination path, SOPS creation
rule and Flux recipient after verifying them there. The Notopia Neovim shortcuts
above do not apply to that repository. Add the resulting encrypted file to that
repository's webhook kustomization, not this one. Do not assume this repository's
SOPS rule or recipient applies to the other repository.

| Secret | Required contents / source |
| --- | --- |
| `valkey-note`, `valkey-billing`, `valkey-document` | Each has its own `password` key for the default Valkey user. |
| `meilisearch` | `MEILI_MASTER_KEY`, a production master key. |
| `search-worker` | `NOTOPIA_SEARCH_WORKER_MEILI_API_KEY`: the same Meilisearch master key. |
| `note` | `NOTOPIA_NOTE_MEILISEARCH_API_KEY`: worker-seeded search key; `NOTOPIA_NOTE_AUTHENTIK_TOKEN`: least-privilege Authentik API token. |
| `billing` | `NOTOPIA_BILLING_STRIPE_SECRET_KEY`: Stripe API secret; `NOTOPIA_BILLING_STRIPE_WEBHOOK_SECRET`: signing secret for the configured endpoint. |
| `document` | `NOTOPIA_DOCUMENT_S3_ACCESS_KEY_ID` and `NOTOPIA_DOCUMENT_S3_SECRET_ACCESS_KEY`: RustFS credentials scoped to the document bucket. |
| `web` | `BETTER_AUTH_SECRET`: independently generated secret; `AUTHENTIK_CLIENT_SECRET`: confidential provider secret. |

The Secret keys above are passed through `envFrom`, except the explicit Valkey
and Meilisearch chart references. CNPG supplies database credentials separately.
Set nonsecret `AUTHENTIK_CLIENT_ID` to the actual provider client ID in web
values; set `NOTOPIA_DOCUMENT_S3_ENDPOINT`, `NOTOPIA_DOCUMENT_S3_REGION`, and
`NOTOPIA_DOCUMENT_S3_BUCKET_NAME` in document values from the provisioned RustFS
service. Configure the actual Authentik user IDs as
`NOTOPIA_NOTE_IDENTITY_SERVICE_ACCOUNT_USER_ID`,
`NOTOPIA_DOCUMENT_IDENTITY_SERVICE_ACCOUNT_USER_ID`, and
`NOTOPIA_BILLING_IDENTITY_SERVICE_ACCOUNT_USER_ID` in their respective values.
Confirm current service configuration before provisioning. Sources:
`internal/note/config/config.go`, `pkg/common/config/authentik.go`,
`pkg/common/config/meilisearch.go`, `internal/billing/config/config.go`,
`apps/document/src/config/env.validation.ts`, `apps/search-worker/seed/seed.ts`,
and `apps/web/lib/auth.ts`.

Provision Authentik application/provider `app-web`
as confidential, with callback `https://notopia.${PUBLIC_DOMAIN}/api/auth/callback/authentik`,
the `notopia` scope mapping, `notopia-admins` group, and cron service accounts.
Provision the RustFS bucket and scoped key, and a Stripe webhook at
`https://notopia-api.${PUBLIC_DOMAIN}/billing/webhooks/stripe`.

After Meilisearch is provisioned and the search worker's seed has created the
configured search key UID, retrieve that key over authenticated HTTPS from
`/keys/<uid>` using the master key, store it in the `note` Secret, and only then
resume note. Never expose the master key to browsers. Tenant tokens authorize the
public search route; CORS is not an authorization control.

CNPG generates `cnpg-<service>-app` for each service. Wire the `uri` key to the
Go charts' `database.urlSecret`. For document, map `host`, `port`, `username`,
`password`, and `dbname` to its discrete database environment variables. These
single-instance clusters have no backup configured; provision backups before
relying on their data.

The route uses the homelab's Gateway API CORS filter for the deployed web origin
and the local gateway origin `https://notopia.${LOCAL_DOMAIN}`. Confirm the
installed Gateway API CRDs and Istio implementation support this filter and
enforce both origins before exposing search. CORS only controls browsers.

## Coordinated secret rotation

Updating a stable Secret reference does not restart pods or refresh environment
variables. Use a supervised maintenance window, inventory actual Secret consumers
(including Jobs and any external clients), and confirm the deployed chart versions,
workload names, replicas, health, backups, provider rotation and rollback behavior.
Keep releases suspended unless explicitly authorized otherwise. Never trigger an
unattended Helm upgrade or migration to refresh a Secret: changing values or resuming
Flux can invoke hooks. Authorization's database Secret is no exception; follow
[Authorization upgrade and recovery](#authorization-upgrade-and-recovery) for any
install, upgrade or rollback, and do not run migration or policy seeding merely to
restart pods. Coordinate CNPG database password changes with the database itself,
not just its Kubernetes Secret.

1. Where supported, create overlapping credentials and retain the previous usable
   credential securely. If overlap is impossible, plan downtime, block affected
   traffic and pause producers/consumers as needed; do not assume a rolling restart
   can keep mismatched credentials working. Prepare and verify an encrypted Git
   rollback and provider-side recovery without disclosing values. Account for
   Better Auth session invalidation, Stripe webhook delivery/retries, and pending
   work. For Valkey, verify the pinned chart's ACL Secret reload behavior and
   persistent data recovery; do not assume an updated Secret changes a running
   server's ACL. For Meilisearch, verify master-key and derived-key behavior with
   the deployed version and preserve an index backup.
2. Update provider and encrypted Secrets in the coordinated order supported by each
   provider, reconcile only the Secret resources, and verify the intended Secret
   versions reached the cluster without printing values. Suspending a HelmRelease
   does not stop existing pods; keep controllers from initiating unrelated Helm
   actions and verify none is in flight. Do not revoke old credentials yet.
3. Restart every affected server and client after its new credentials are available.
   The checked-in application charts render Deployments named `web`, `note`,
   `billing`, `document`, `search-worker`, and `authorization` in `notopia`:
   for each affected deployment, after confirming the live name and intended
   operation, run `kubectl -n notopia rollout restart deployment/<name>` and
   `kubectl -n notopia rollout status deployment/<name> --timeout=10m`.
   These commands restart pods, not Helm hooks; for authorization confirm no Helm
   action or seeding is in progress and preserve its existing image and policy.
   Restart only after database credentials are valid. Do not scale document above
   one replica. Completed Jobs retain old environments; ensure future Jobs are
   newly created with the current Secret, without rerunning migrations as a test.
4. For `meilisearch` and `valkey-note`, `valkey-billing`, `valkey-document`, inspect
   each installed release's `helm get manifest <release> -n notopia` and live
   workload to identify its exact Deployment or StatefulSet name, controller,
   replica count and Secret mounts; chart versions can differ. After confirming
   the supported reload/rotation procedure, use `kubectl -n notopia rollout
   restart <deployment-or-statefulset>/<actual-name>` followed by `kubectl -n
   notopia rollout status <deployment-or-statefulset>/<actual-name> --timeout=10m`.
   Do not guess selectors or restart all workloads. Coordinate Valkey server and
   its note or billing client; document currently has no Valkey URL configured,
   so inventory any separately configured consumers. For Meilisearch coordinate
   the server and search-worker master key, retrieve the new worker-seeded search
   key securely, update note's Secret and restart note; verify tenant tokens and
   browser search. Determine whether seeding is needed from the deployed version;
   do not blindly rerun a Helm release or its hooks.
5. Verify replacement pods, readiness, authenticated functionality and provider
   logs without revealing credentials; confirm every consumer and newly created
   Job has switched. Only then revoke old credentials. If any restart, health or
   functional check fails, stop the sequence and keep traffic blocked as needed.
   Restore the previous provider credentials and encrypted Secret where supported,
   reconcile the Secret and restart affected workloads again, checking compatibility
   and data recovery first. If reversal is unsupported, remain in maintenance and
   use the provider's documented recovery; Helm rollback does not restore external
   credentials or data. Escalate rather than revoking a still-needed credential.

## Ordered initial bootstrap

All paths below are relative to this repository unless identified as homelab.
Stop at any failed gate; do not expose an incomplete deployment.

1. **Plan and protect existing data.** Inventory existing databases, Valkey,
   Meilisearch, buckets, Kafka offsets and releases before changing them. CNPG
   `initdb` creates new databases; it does not migrate existing data. Arrange
   tested backups and an explicit restore/migration plan, migration compatibility
   and maintenance windows; do not overwrite existing volumes or assume a Helm
   rollback reverses schema changes. The single-instance CNPG clusters have no
   configured backups. Validate image pull access and immutable, chart-compatible
   images for every release; empty tags are not verified versions.
2. **Verify the platform.** Confirm the Flux source, `sops-age` decryption,
   `clustersetting` substitutions, Gateway API CORS support, DNS, TLS, storage,
   CNPG, Istio service accounts, ambient enrollment and mTLS. The namespace's
   ambient label is intentionally omitted: determine the actual enrollment
   mechanism before adding any label; verify capture and policy enforcement
   with an independently provisioned test workload before enabling pointers.
   Verify the checked-in Redpanda address,
   advertised listeners and network access against the live deployment. Confirm
   the webhook path routing and existing shared gateway applications. Do not
   assume local identities or service account names are production identities.
3. **Provision identities and secrets.** Configure Authentik's confidential
   `app-web` provider and callback, attach the `notopia` scope mapping from the
   design's referenced identity contract, create `notopia-admins`, and provision
   distinct cron service-account users. Configure their actual user IDs and a
   least-privilege Authentik API token for note; never substitute a human or local
   development identity. Provision the RustFS bucket and scoped credentials,
   Stripe secret and signing secret, and a unique Better Auth secret. Populate
   the nonsecret values and encrypted Secrets described above. For search-worker
   use `NOTOPIA_SEARCH_WORKER_MEILI_API_KEY` with the Meilisearch master key;
   note eventually needs `NOTOPIA_NOTE_MEILISEARCH_API_KEY` and its Authentik
   token. Check the current service configuration for every other exact key.
   Commit only encrypted Secrets and their kustomization entries; confirm Flux can
   decrypt them without printing values. Use provider-supported rotation and
   revoke superseded credentials only after all consumers have switched.
4. **Stage Flux in Git.** The homelab `notopia-infra`, `notopia-edge`, and
   `notopia-apps` Kustomizations are suspended. `notopia-apps` depends on both
   infra and edge; both infra and edge use `wait: true`. A suspended HelmRelease
   may prevent a waiting Kustomization from becoming Ready. Do not assume the
   dependency chain allows the worker to start. For bootstrap, make a reviewed,
   temporary homelab Git change removing only `notopia-edge` from the apps
   `dependsOn`; if suspended infra releases block readiness, temporarily set
   infra `wait: false`. Keep the edge Kustomization and HelmRelease suspended,
   and all app releases except the worker suspended. Resume infra and apps by
   changing their homelab `spec.suspend` to `false`, and reconcile the homelab
   source using the normal Flux process. This staging bypass is not approval to
   expose edge; restore both dependency and original `wait` after the edge has
   been independently secured and is Ready. Confirm each stage's status rather
   than assuming suspended resources are healthy.
5. **Start infrastructure and worker.** In this repository change the relevant
   infrastructure HelmRelease `spec.suspend` to `false` in Git after its Secret
   and prerequisites are ready. Reconcile source and infra and verify CNPG,
   Meilisearch, and required Valkey instances are healthy. Then set a verified
   worker image tag, change `apps/search-worker/helmrelease.yaml` to
   `spec.suspend: false`, commit and reconcile apps. The worker's seed hook must
   complete independently of note. Both worker and note currently use UID
   `00000000-0000-4000-8000-000000000001`; confirm they still match. The seed
   creates or reuses a `notes`-only search key and configures index settings; it
   does not backfill existing data. Plan and verify a separate reindex if needed.
6. **Transfer the search key securely.** After the seed Job succeeds, use an
   authenticated, verified TLS connection to Meilisearch's `/keys/<uid>` API
   with the master key and securely transfer the returned `key` into the note
   Secret's `NOTOPIA_NOTE_MEILISEARCH_API_KEY`. Use a secret manager or protected
   interactive tooling that does not expose credentials in command arguments,
   history, logs, terminal recording, or temporary files; do not print the
   response. Encrypt and commit the Secret and reconcile infra before note is
   resumed. The master key must never reach a browser. A master-key rotation can
   change derived keys: coordinate worker and Meilisearch credentials, retrieve
   the new search key, update note and verify tenant tokens before revocation.
7. **Seed authorization and start applications.** Before every authorization
   install, upgrade **and rollback**, follow
   [`authorization/README.md`](../../charts/authorization/README.md): run
   exactly one external `./authorization seed` using the same verified image,
   database and secrets as the intended deployment. For upgrades and rollbacks,
   block traffic, stop all authorization pods and prevent recreation and
   concurrent seeders; restore traffic only after replacement pods are healthy.
   The chart also runs a pre-install/pre-upgrade seed hook, but it does not replace
   the required external seed or protect against concurrent policy writers. Then
   resume authorization and each other
   app in Git only after its own secrets, image, dependencies and migrations are
   ready; keep document at one replica. Explicitly inspect migrate and seed Jobs.
8. **Secure edge, then restore the graph.** Independently implement and verify
   the non-bypassable identity-header boundary and adversarial tests in
   [`notopia-edge/README.md`](../../charts/notopia-edge/README.md). Only then set
   its acknowledgment value, resume the edge HelmRelease and its homelab
   Kustomization in Git, and reconcile. An acknowledgment is not a control.
   Restore apps' edge dependency and infra's original `wait: true` if changed;
   verify the full graph is Ready. Suspending or a failed upgrade does not remove
   or secure an existing release; do not rely on suspension as traffic control.

## GitHub Receiver

In the homelab repository securely create and SOPS-encrypt the missing
`notopia-webhook-token` Secret in `flux-system` with key `token`, include it in
the webhook kustomization, confirm
Flux can decrypt it, then set Receiver `spec.suspend: false` and reconcile.
Configure a GitHub JSON webhook for `ping` and `push` using the same secret.
After the Receiver becomes Ready, obtain its actual `.status.webhookPath` and
append it directly to `https://flux-webhook.${PUBLIC_DOMAIN}`. The status path
already starts with `/hook/`; **do not** prepend another `/hook/`. Verify a ping
and a push reach the Receiver and trigger the Notopia GitRepository reconcile.
Rotate the shared token on both sides in a coordinated window without exposing it.

## Authorization upgrade and recovery

The general plan calls for automatic rollback on a failed upgrade, but that can
restart the old image with the new image's policy before an operator can seed
the old policy. Authorization is therefore an explicit exception: its checked-in
HelmRelease has `upgrade.remediation.retries: 0` and
`upgrade.remediation.remediateLastFailure: false`. The `rollback` strategy
remains configured but no final-failure remediation is requested. Other releases
retain the planned automatic rollback. See the
[Flux HelmRelease failure handling documentation](https://fluxcd.io/flux/components/helm/helmreleases/#configuring-failure-handling).

**Do not permit unattended authorization upgrades.** Before initial install or
any later upgrade, verify the deployed helm-controller version honors these
settings, the live HelmRelease has observed them, and no other automation can
rollback, reinstall, change the chart or values, or restart authorization pods
during seeding and failure recovery. This has not been tested in the homelab.
If those controls cannot be verified, keep authorization suspended and do not
upgrade it. A false remediation setting alone does not enforce policy seeding,
stop running pods, or prevent new source revisions from triggering upgrades.

Use a supervised maintenance window for every desired change, including chart,
image, and values changes. First suspend the HelmRelease in Git and confirm the
live suspension has been observed and no Helm action is in flight; block traffic,
stop pods and prevent their recreation and concurrent seeders as required by the
authorization chart. Seed once with the target immutable image and database.
Only after success apply the reviewed desired revision and resume reconciliation
for that operation. Monitor the result; once stable, suspend the release again to
prevent unattended subsequent upgrades. Control other reconciler and autoscaler
activity throughout. Suspension is not a substitute for blocking traffic or
stopping workloads, and cannot cancel an already-running Helm action.

On failed upgrade, keep traffic blocked and stop any remaining target pods;
ensure no controller can recreate them. Do not force a Flux retry or run a Helm
rollback before seeding. With reconciliation suspended and no action in flight,
assess database migration compatibility and restore data if necessary; seed
exactly once using the chosen rollback image and matching database and secrets.
Only then use a reviewed, operator-controlled rollback/reconciliation to the
matching chart, image and values, verify the actual pods and policy, and restore
traffic. If Flux or Helm state prevents that sequence, remain stopped and resolve
it before proceeding. This is a proposed operational alternative, not a
live-validated rollback guarantee.

## Live verification (pending authorization)

Record actual timestamps and results, without tokens or sensitive response bodies:

```sh
flux get kustomizations -n flux-system
flux get helmreleases -n notopia
kubectl -n notopia get jobs
```

Confirm all intended Notopia Kustomizations and HelmReleases are Ready and the
latest migrate and seed Jobs completed. Check public `/note/health/ready` is 200;
protected requests without a token and forged-group requests are 403. Verify a
valid non-admin receives 403 on a real billing admin endpoint and a valid admin
is allowed. Run the full edge adversarial suite, including all five spoofed
headers and bypass paths, before exposure. Verify web login, `wss` editing,
tenant-token search, webhook-triggered reconciliation, Tempo edge → note →
authorization and migrate Job spans, and unchanged Kite and Authentik operation.
All of these cluster and external checks remain pending until authorized and run.

## Failure and rollback

Stop reconciliation or traffic as appropriate and investigate failed Jobs,
health, and logs without publishing secrets. Git-revert the affected desired
configuration and reconcile only after assessing the data impact; Helm rollback
cannot undo database or index migrations. Restore from tested backups if needed.
Authorization rollback requires the exclusive external seed procedure above using
the rollback image before pods restart. Restore staging dependencies and any
temporarily changed wait settings once safe; never remove the edge security gate.

## Local checks

Run `node --test deploy/k8s/clusters/homelab/*.test.ts`
from the repository root. The tests require installed dependencies, `helm`, `kubectl`, Git, and SOPS;
synthetic credentials are encrypted in memory for the validation tests. The
repository check examines Git-tracked and non-ignored untracked YAML and JSON
Kubernetes Secret documents, including `List` resources. Structural SOPS
checks do not decrypt or authenticate the MAC; operators must verify decryption
with the actual Flux identity before resuming releases.
