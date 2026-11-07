# Kubernetes Homelab Wiring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deploy Notopia through Flux and reconcile main-branch pushes through a GitHub Receiver.

**Architecture:** Notopia owns cluster-specific infrastructure, values and releases; the homelab repository owns the namespace, source, Receiver and three dependency-ordered Flux Kustomizations. Existing platform services remain shared and unchanged.

**Tech Stack:** Flux, Kustomize, SOPS/age, CNPG, Valkey, Meilisearch, Istio ambient, Gateway API.

**Spec:** `docs/superpowers/specs/2026-10-24-k8s-03-homelab-wiring-design.md` and `docs/superpowers/specs/2026-10-24-k8s-00-overview-design.md`. Requires plans `2026-11-03-k8s-01-app-readiness.md` and `2026-11-03-k8s-02-helm-charts.md`.

## Global Constraints

- Four independent CNPG clusters, three Valkey instances, no CNPG backups; no api-web deployment.
- Use existing Authentik, RustFS, Redpanda, CNPG, Istio and Alloy; do not alter the shared gateway.
- Secrets are encrypted with SOPS age recipient `age1ytym5g6mxf0974syp2g2ecrh0y74tnl0u9ee4xm8caawhyv25pls3f2sy4`; never commit plaintext credentials.
- Homelab changes are in `~/projects/homelab`; inspect its instructions and worktree before edits. Do not apply cluster changes or create external credentials without user authorization.
- Flux intervals 12h, Receiver events `ping,push`; runtime public domain is `${PUBLIC_DOMAIN}`.

## Review Focus

- First install: migration hooks need existing secrets, CNPG readiness and service account (Tasks 1–3).
- Bootstrap search key: missing note key must not prevent worker seed from completing (Tasks 1, 3, 4).
- Paths and forwarded headers: unauthenticated or non-admin admin requests remain denied (Tasks 2, 4).
- Shared gateway, existing applications and local domain routes continue working (Tasks 2, 4).
- Secret rotation or failed rollout: no plaintext leak; failure visible without unsafe migration retry (Tasks 1–4).

## File Structure

`deploy/k8s/clusters/homelab/infra/` contains CNPG, Valkey, Meilisearch and encrypted secrets; `edge/` and `apps/` contain HelmReleases and generated values ConfigMaps. `~/projects/homelab/kubernetes/apps/notopia/` contains Flux pointers and Receiver; namespace is defined in its existing namespaces file.

### Task 1: Infrastructure and encrypted configuration

**Files:** Create `deploy/k8s/clusters/homelab/infra/{kustomization.yaml,cnpg/,valkey/,meilisearch/,secrets/}`; modify repository `.sops.yaml`; add `deploy/k8s/clusters/homelab/README.md`.

**Interfaces:** CNPG clusters `cnpg-{note,billing,authorization,document}`, each producing `cnpg-<service>-app` secret; Valkey secrets `valkey-{note,billing,document}`; Meilisearch master key Secret `meilisearch`; service Secrets named as spec §This repo.

- [ ] Add structural tests for four independent CNPG 18 clusters, instances 1, storage 1Gi, disabled `ro,r`, three persistent noeviction/appendonly Valkeys, production Meilisearch with persistent storage, tenant-token public HTTPRoute and origin-limited CORS. Test that repository secrets contain SOPS metadata and no plaintext secret values.
- [ ] Run `kustomize build deploy/k8s/clusters/homelab/infra`; expect failure before implementation.
- [ ] Verify exact official chart versions/values against upstream and existing homelab conventions. Create resources, SOPS rule, encrypted secrets using actual user-provided credentials only; for absent credentials document provisioning and keep dependent releases suspended until available, never commit fake production secrets. Map CNPG `uri`, and document discrete connection keys.
- [ ] Run `kustomize build deploy/k8s/clusters/homelab/infra` and strict kubeconform with Flux/CNPG/Gateway schemas; run secret structural tests. Expect PASS. Validate manifests without contacting production.
- [ ] Commit Notopia repo files with `feat(deploy): provision homelab infrastructure`.

### Task 2: Edge and app HelmReleases

**Files:** Create `deploy/k8s/clusters/homelab/{edge,apps}/` with kustomizations, nameReference configurations, HelmReleases and six per-service values files.

**Interfaces:** Consume plan 02 charts, Task 1 Secret names and CNPG names. Export HelmReleases named `notopia-edge`, `web`, `note`, `authorization`, `billing`, `document`, `search-worker`; `apps/<service>/values.yaml` carries `image.tag`.

- [ ] Test Kustomize output includes generated values ConfigMap references that change name when values change; each release uses the GitRepository chart path, 12h interval, timeout 10m, no install retries and rollback upgrade remediation. Test endpoint, gRPC caller principals, ports, secrets, and public route rules against overview.
- [ ] Run `kustomize build` for edge and apps; expect failure.
- [ ] Implement releases and values per spec §§138–213, including per-service secrets, Alloy endpoint, broker addresses verified from actual homelab, service principals and web `wss://` URL. Expand public health rules into valid explicit Istio path patterns. Keep admin JWT rules disjoint from general JWT rules. Set initial image tags to actual available images, not invented versions.
- [ ] Render each chart using generated values and validate with `helm template` plus kubeconform and chart tests. Run `kustomize build deploy/k8s/clusters/homelab/edge` and `kustomize build deploy/k8s/clusters/homelab/apps`; expect PASS.
- [ ] Commit with `feat(deploy): configure homelab releases`.

### Task 3: Homelab Flux source and Receiver

**Files:** In `~/projects/homelab`, modify `kubernetes/cluster/namespaces.yaml` and existing app inclusion; create `kubernetes/apps/notopia/{kustomization.yaml,repos/,webhook/,notopia-infra.yaml,notopia-edge.yaml,notopia-apps.yaml}`.

**Interfaces:** `GitRepository notopia` in `flux-system`; Receiver `notopia` references it; Flux Kustomizations `notopia-infra`, `notopia-edge`, `notopia-apps` use the exact paths in spec §Homelab repo changes and dependency chain infra → edge → apps.

- [ ] Test rendered source uses branch main, 12h and the exact deploy/k8s include ignore; Receiver listens for ping/push; infra waits for CNPG, edge waits for infra, apps waits for both; wait is false only for apps; decryption and substitutions mirror existing apps.
- [ ] Run `kustomize build kubernetes/apps/notopia` from homelab; expect failure before creation.
- [ ] Inspect ambient enrollment in actual cluster configuration; add the namespace label only if compatible. Add encrypted Receiver token and Flux pointers following homelab conventions. Ensure app inclusion is wired into its existing parent kustomization; never overwrite unrelated work.
- [ ] Run the kustomize build and schema validation; inspect `git -C ~/projects/homelab diff` and confirm no unrelated changes. Expect PASS. Commit only the listed homelab files with `feat(notopia): connect Flux source and receiver`.

### Task 4: Bootstrap runbook and integration verification

**Files:** Complete `deploy/k8s/clusters/homelab/README.md`; update the Notopia and homelab configs if integration exposes defects.

**Interfaces:** Consume Tasks 1–3. Bootstrap worker Meilisearch key using the same UID as note, then inject its retrieved value into note's SOPS Secret; configure Authentik scope, cron users, RustFS, Stripe and GitHub Receiver as spec §Manual runbook.

- [ ] Write ordered runbook with explicit initial reconciliation sequence that lets worker seed run before note's key is populated, secret creation/rotation, existing-data migration precautions, Receiver path and webhook, rollback behavior, and safe credential handling. Specify how to resume any suspended release.
- [ ] After explicit user authorization and credentials, deploy; verify `flux get kustomizations`, `flux get helmreleases -n notopia`, and `kubectl -n notopia get jobs`: expected Ready and completed hook Jobs.
- [ ] Verify public health 200, unauthenticated protected requests 403, forged groups 403, valid non-admin admin request 403 and valid admin allowed; validate web login, wss editing and tenant-token search. Confirm webhook push triggers reconciliation, Tempo trace spans edge→note→authorization and migrate job, and Kite/Authentik still function.
- [ ] Record actual command results and any unavailable external checks as pending; commit runbook with `docs(deploy): document homelab bootstrap`.
