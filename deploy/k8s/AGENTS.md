# Kubernetes charts

Run commands from the repository root. Install the versions pinned in `mise.toml` with
`mise install`, install workspace dependencies with `pnpm install`, and run `just helm-setup`
for the pinned helm-unittest plugin. See each chart's `project.json` for its exact targets.

- Edit `charts/<name>/values.yaml`, templates, and any `README.md.gotmpl`; do not hand-edit
  generated `values.schema.json` or chart README files. Run `nx run <name>-chart:gen`
  (use `notopia-common-chart` for the library) to regenerate schema and documentation.
- Run `nx lint <name>-chart` and `nx test <name>-chart` for each affected chart, including
  dependents of the common library. Targets build locked dependencies, run Helm lint and
  kubeconform, and run helm-unittest plus Node tests. Kubeconform's CRD schemas require network
  access. Render and unit tests are not live-cluster security or upgrade validation.
- Before release, bump the SemVer `Chart.yaml` version for every changed chart. A common
  library change also requires bumping every dependent chart and updating its dependency version
  and lockfile. Flux's ChartVersion reconciliation depends on version changes.

## Safety

- Follow `charts/authorization/README.md` for mandatory external seeding before every install,
  upgrade, and rollback; the post hook is too late for `--wait` and policy loads at startup.
- Do not deploy or expose the edge until its external identity-header boundary is independently
  implemented and verified; acknowledgement values are assertions, not security controls.
  Read `charts/notopia-edge/README.md` for required controls and adversarial checks.
- Keep document at one replica; Redis alone does not make event delivery safe across replicas.
- See [deployment limitations](../../docs/src/limitation/kubernetes-deployment.md) for remaining
  validation requirements and [document replicas](../../docs/src/limitation/document-replica-events.md)
  for the event-delivery limitation.
