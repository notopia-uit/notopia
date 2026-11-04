# Deferred chart follow-up

These issues are deferred for user review. Existing chart rendering and tests do not establish
that deployment, upgrades, or the identity boundary are safe.

- **Authorization seed timing:** the post-install/post-upgrade seed hook runs too late for
  `helm --wait`, while policy is loaded at process startup. Review a reliable seed and rollout
  design, including rollback and concurrency. Until then, follow the downtime-requiring external
  seed procedure in `charts/authorization/README.md`; the chart cannot verify it occurred.
- **Edge trust boundary:** determine and independently verify a non-bypassable way to strip all
  client identity headers before trusted claims are set, and prevent direct downstream access.
  NetworkPolicy and acknowledgement flags are not proof. Follow the deployment blockers and
  controls in `charts/notopia-edge/README.md`.
- **Live Istio identity and spoof validation:** in a representative cluster, test direct edge
  access, alternate routes, spoofed-label pods, all five forged identity headers with anonymous
  and missing-claim tokens, and ordinary versus admin access. Validate non-Istio mode separately
  if used. Repeat after configuration or infrastructure changes.
- **Node live upgrade and gateway tests:** validate document and search-worker upgrades against
  running workloads, including hooks and rollout behavior, and exercise gateway routing and health
  exposure on the intended implementation. Rendered manifests and Node tests are not substitutes.
- **Document replicas:** deletion and permission events may not reach every replica; Redis alone
  is insufficient. Keep homelab at one replica until a separately approved fix and multi-replica
  integration test exist.
