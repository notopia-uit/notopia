# Homelab infrastructure provisioning

The Valkey and Meilisearch HelmReleases are suspended because no credentials have
been supplied. Do not resume a release until its encrypted Secret is committed,
included in `infra/secrets/kustomization.yaml`, and Flux has the matching age
private key. Application releases that depend on missing credentials must also
remain suspended; this directory does not yet contain the application releases.
Do not expose the edge until the independent identity-header boundary checks in
`deploy/k8s/AGENTS.md` have passed.

## Secrets

Create production credentials through the relevant provider or a secure local
process. Never put plaintext in Git, shell history, command-line arguments, or
unprotected temporary files. Create each Kubernetes Secret locally with the
appropriate name and keys, encrypt it with SOPS using the root `.sops.yaml` rule,
review that all `data` or `stringData` values are encrypted, and add the
`.secret.sops.yaml` file to `infra/secrets/kustomization.yaml`. The recipient is
the homelab SOPS age key; the Flux decryption key must already be provisioned.

| Secret | Required contents / source |
| --- | --- |
| `valkey-note`, `valkey-billing`, `valkey-document` | Each has its own `password` key for the default Valkey user. |
| `meilisearch` | `MEILI_MASTER_KEY`, a production master key. |
| `search-worker` | The same Meilisearch master key under the worker's configured environment key. |
| `note` | Meilisearch search API key, obtained after worker seed as described below. |
| `billing` | Stripe secret key and webhook signing secret. |
| `document` | RustFS access key and secret for the document bucket. |
| `web` | `BETTER_AUTH_SECRET` and `AUTHENTIK_CLIENT_SECRET`; configure the client ID as well. |

Use each service's current configuration definitions to select its exact
environment variable names. Provision Authentik application/provider `app-web`
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

The route uses the homelab's Gateway API CORS filter. Confirm the installed
Gateway API CRDs and Istio implementation support this filter and enforce the
specified origin before exposing search. CORS only controls browsers.

Finally configure the GitHub push webhook to the Notopia Flux Receiver as
described in the homelab wiring design; this is a separate manual operation.

## Local checks

Run `python3 -m unittest discover -s deploy/k8s/clusters/homelab -p 'test_*.py' -v`
from the repository root. The tests require PyYAML, `kubectl`, Git, and SOPS;
synthetic credentials are encrypted in memory for the validation tests. The
repository check examines Git-tracked and non-ignored untracked YAML and JSON
Kubernetes Secret documents, including `List` resources. Structural SOPS
checks do not decrypt or authenticate the MAC; operators must verify decryption
with the actual Flux identity before resuming releases.
