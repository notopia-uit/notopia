# Kubernetes Helm Charts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide portable, tested Helm charts for every deployable service and the API edge.

**Architecture:** A library chart supplies shared naming, environment, security and Jobs; six service charts and an edge chart own service-specific resources. Values and external Secrets carry configuration; no chart depends on Flux or SOPS.

**Tech Stack:** Helm, helm-schema, helm-docs, helm-unittest, kubeconform, Nx, Gateway API, Istio.

**Spec:** `docs/superpowers/specs/2026-10-24-k8s-02-helm-charts-design.md` and `docs/superpowers/specs/2026-10-24-k8s-00-overview-design.md`. Requires completed `2026-11-03-k8s-01-app-readiness.md`.

## Global Constraints

- One chart each for web, note, authorization, billing, document, search-worker and notopia-edge, plus `notopia-common-chart` library.
- Default one replica; each chart version changes whenever it or the library changes; generated schema and README are committed.
- Secrets precede Helm hooks; do not use a release-owned ConfigMap for hook env. Chart resources work without Istio or Flux.
- Endpoint, header and authorization contract is overview §§58–90. Admin group protection cannot be bypassed by another ALLOW rule.

## Review Focus

- Non-admin authenticated request to admin path must be denied despite default JWT rule (Task 4).
- Invalid wildcard paths must not expose unintended endpoints (Task 4).
- Multi-replica or autoscaled document/search-worker without required shared state must fail rendering (Task 3).
- Missing chart dependency or invalid generated schema must fail CI rather than skip validation (Tasks 1–3).
- Hook before ServiceAccount creation and secret availability must not deadlock installation (Tasks 2–3).
- Document scaling must not be presented as production-safe until deletion and permission events reach every replica (Task 3).

## File Structure

`deploy/k8s/charts/notopia-common-chart/` provides helpers; peer chart directories own `Chart.yaml`, values, generated artifacts, templates, tests and Nx `project.json`. Edge owns Gateway API and Istio policies. Tool versions belong in both mise configs.

### Task 1: Chart toolchain and library

**Files:** Modify `mise.toml`, `.mise.ci.toml`, `.gitignore`; create `deploy/k8s/charts/notopia-common-chart/{Chart.yaml,values.yaml,project.json}` and `templates/*.tpl`, plus a test fixture chart under `deploy/k8s/charts/notopia-common-chart/tests/`.

**Interfaces:** Named templates `notopia-common.{name,fullname,chart,labels,selectorLabels,serviceAccountName,image,env,otelEnv,memoryEnv,hookJob,podSecurityContext,containerSecurityContext,istioWorkloadPolicy}` as specified in §notopia-common.

- [ ] Pin matching versions in both mise files: helm 3 (verify compatible available 3.x release before pinning; Helm 4 is not assumed compatible), `github:dadav/helm-schema` 0.24.0, `github:norwoodj/helm-docs` 1.14.2, `github:yannh/kubeconform` 0.8.0 and `yq` 4.54.1; pin helm-unittest plugin 1.2.1 in setup. Verify binaries install with mise before choosing the Helm 3 exact patch and record that patch in both files.
- [ ] Write fixture tests: `256Mi` Go limit yields `GOMEMLIMIT` at 90%, Node uses 75%; digest overrides tag; env and secrets render identically in Jobs and deployment; restrictive security and downward API attrs render. Run `helm unittest` on fixture; expect failure.
- [ ] Implement helpers and Nx gen/lint/test targets; ignore generated `charts/` dependencies, not source charts. Derive memory quantities safely or fail clearly for unsupported quantities; do not emit duplicate environment variable names.
- [ ] Run fixture tests, `helm lint --strict` and `nx run notopia-common-chart:gen`; expect PASS and deterministic generated files.
- [ ] Commit with `feat(helm): add shared chart library`.

### Task 2: Go service charts and hooks

**Files:** Create complete chart trees for `deploy/k8s/charts/{note,authorization,billing}/`, including Chart, values, templates, schema, README, tests and Nx project.

**Interfaces:** Consume Task 1 helpers and Task 1 plan's commands. Produce chart names `note`, `authorization`, `billing`; use port/probe/command table in spec §Service charts and hook annotations/weights in §Hook order.

- [ ] Write helm-unittest cases for each command, port and health route; migrate pre-install/pre-upgrade weight -5 for note/billing; seed post-install/post-upgrade weight 5 for authorization/billing; no extra hook; missing required Redis config fails; Istio disabled emits no security objects.
- [ ] Run `helm unittest` for each chart; expect failures.
- [ ] Implement charts with inline env, external secret refs, security contexts, optional routes/PDB/HPA and principal policy. Pre-install migration must use an already existing service account or a Helm pre-install hook ServiceAccount with earlier weight; do not reference an ordinary release ServiceAccount that does not yet exist. Specify and test cleanup behavior for any hook account.
- [ ] Run `helm dependency build`, `nx run <chart>:gen`, `nx lint <chart>` and `nx test <chart>` for all three; kubeconform strict must validate rendered resources including configured CRD schemas. Expect PASS and clean regeneration.
- [ ] Commit with `feat(helm): add Go service charts`.

### Task 3: Node service charts and scaling

**Files:** Create complete chart trees for `deploy/k8s/charts/{document,search-worker,web}/` with corresponding generated files, tests and Nx projects.

**Interfaces:** Consume Task 1 helpers and plan 01 Node entries; use port/probe/command table in spec §Service charts.

- [ ] Test document migration hook, worker seed hook, no web hook, health route toggles, Istio off, and TCP web probe. Test document replicas >1 or HPA without `scaling.enabled` plus Redis secret fails; search-worker replicas >1 **or HPA** without shared Redis fails. Test valid shared-state configurations render.
- [ ] Run `helm unittest` for each chart; expect failure.
- [ ] Implement service charts and guards. Ensure HPA cannot scale worker above one without Redis; this closes the spec's replica-only guard gap. Apply the pre-install service-account ordering from Task 2. Document that Redis configuration alone does not establish safe document scale-out: `apps/document/AGENTS.md` identifies per-replica deletion/permission delivery as unresolved. Keep homelab at one replica until a separately approved fix and multi-replica integration test exist.
- [ ] Run dependency build, Nx gen/lint/test for all three; expect PASS and strict schema validation. Confirm generated docs/schema have no diff on second gen.
- [ ] Commit with `feat(helm): add Node service charts`.

### Task 4: API edge and authorization

**Files:** Create complete `deploy/k8s/charts/notopia-edge/` chart, templates including `istio/`, values, schemas, README, tests and Nx project.

**Interfaces:** Consume Task 1 naming helpers. Produce Gateway, front HTTPRoute header removal, RequestAuthentication claim mapping, AuthorizationPolicy and optional Telemetry according to spec §notopia-edge.

- [ ] Write tests for all five stripped headers, mapping `groups_json`, Istio off, public routes, exact health paths for each service, admin group enforcement, and default authenticated rule. Assert the ordinary JWT allow excludes both `/billing/admin` and `/billing/admin/*` so ALLOW union cannot bypass group rule. Test anonymous, ordinary user and admin decisions in a real Istio integration environment.
- [ ] Run `helm unittest deploy/k8s/charts/notopia-edge`; expect failure.
- [ ] Implement policies with disjoint public, admin and general JWT matches. Do not use the spec's `/*/health/*` as an Istio simple wildcard: expand to supported per-service prefixes. Ensure public rules cannot overlap admin; reject invalid/conflicting rule values during rendering. Document external JWT/header layer when Istio is disabled.
- [ ] Run dependency build, Nx gen/lint/test for edge; deploy to a disposable Istio test cluster and exercise all three identities and spoofed headers. Expect PASS and 403 for non-admin admin requests. If a cluster is unavailable, record integration verification as pending.
- [ ] Commit with `feat(helm): add authenticated API edge`.
