# Kubernetes charts

Run commands from repository root. Install pinned `mise.toml` versions with
`mise install`, dependencies with `pnpm install`, and pinned helm-unittest plugin with `just helm-setup`.
See each chart's `project.json` for exact targets.

- Edit `charts/<name>/values.yaml`, templates, and `README.md.gotmpl`; never hand-edit
  generated `values.schema.json` or chart READMEs. Run `nx run <name>-chart:gen`
  (use `notopia-common-chart` for library) to regenerate schema and docs.
- Run `nx lint <name>-chart` and `nx test <name>-chart` for each affected chart, including
  common library dependents. Targets build locked dependencies, run Helm lint,
  kubeconform, helm-unittest, and Node tests. Kubeconform CRD schemas need network
  access. Render and unit tests do not validate live-cluster security or upgrades.
- Before release, bump SemVer `Chart.yaml` version for every changed chart. Common
  library changes also require bumping every dependent chart and updating its dependency version
  and lockfile. Flux ChartVersion reconciliation depends on version changes.

## Safety

- Follow `charts/authorization/README.md` for mandatory external seeding before every install,
  upgrade, and rollback; post hook is too late for `--wait`, and policy loads at startup.
- Do not deploy or expose edge until its external identity-header boundary is independently
  implemented and verified; acknowledgement values are assertions, not security controls.
  Read `charts/notopia-edge/README.md` for required controls and adversarial checks.
- Keep document at one replica; Redis alone does not make event delivery safe across replicas.
- See [deployment limitations](../../docs/src/limitation/kubernetes-deployment.md) for remaining
  validation requirements and [document replicas](../../docs/src/limitation/document-replica-events.md)
  for event-delivery limitation.