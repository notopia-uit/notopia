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
- **Non-root image and filesystem validation:** all six service Dockerfiles default to root; charts
  supply numeric UID/GID 1000 and fsGroup 1000, with configurable IDs and read-only root filesystems.
  Image contents, dependency permissions, hook commands, and framework runtime writes have not been
  verified against the actual built images or a cluster. Before deployment, run each image and every
  migration/seed hook with the configured IDs and a read-only root, exercising real requests and
  startup dependencies. Override the IDs where image or volume ownership requires it. Web mounts
  an ephemeral writable `/app/apps/web/.next/cache`, which hides any image contents at that path;
  current web source has no explicit revalidation or ISR configuration, but Next.js may write ISR
  output under `.next/server`, outside that mount. Runtime behavior and writable paths remain pending
  user review and verification against the built image; do not assume this chart is runtime-ready.
  If other workloads need
  writable paths such as `/tmp`, mount appropriate volumes and verify their ownership; hook jobs do
  not currently expose extra volume mounts, so such a hook requires a chart change. No claim is made
  that read-only root or arbitrary IDs work with every image.
- **Document replicas:** deletion and permission events may not reach every replica; Redis alone
  is insufficient. Keep homelab at one replica until a separately approved fix and multi-replica
  integration test exist.
