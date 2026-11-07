# Kubernetes deployment limitations

Chart rendering and unit tests do not establish that deployments, upgrades, or identity boundaries
are safe. Complete the following checks on the intended cluster and built images before deployment.

- **Authorization seeding:** the post-install/upgrade hook runs too late for `helm --wait`, while
  policy loads at startup. Follow the downtime-requiring external seed procedure in the
  `deploy/k8s/charts/authorization/README.md` for installs,
  upgrades, and rollbacks. The chart cannot verify it occurred. A future design must address
  rollout, rollback, and concurrency.
- **Edge trust boundary:** implement and independently verify a non-bypassable identity-header
  boundary and prevent direct downstream access. NetworkPolicy and acknowledgement flags alone
  are not proof. Follow the blockers in the
  `deploy/k8s/charts/notopia-edge/README.md`.
- **Live identity tests:** test direct edge access, alternate routes, spoofed-label pods, all five
  forged identity headers with anonymous and missing-claim tokens, and ordinary versus admin access.
  Test non-Istio mode separately if used; repeat after configuration or infrastructure changes.
- **Live upgrades and gateway:** exercise document and search-worker upgrades, hooks, rollout,
  gateway routing, and health exposure against running workloads on the intended implementation.
- **Image and filesystem:** test all six built service images and their migration/seed hooks with
  the configured UID, GID, filesystem mode, and actual startup dependencies and requests. The
  charts default to UID/GID 1000 and a read-only root, but these have not been validated against
  all images. Use writable volumes for required paths and verify ownership. Hook jobs do not expose
  extra volume mounts; hooks requiring them need a chart change. Root and a writable root filesystem
  are available for trusted debugging, but reduce isolation and do not persist container writes.
- **Web cache:** `.next/cache` is an ephemeral writable volume that masks image contents at that
  path. No explicit ISR configuration was found in the current web source. Verify the built image's
  runtime writes; ISR output can require writable paths outside that mount, including `.next/server`.
- **Document scaling:** see [document replica event delivery](./document-replica-events.md).

For trusted debugging, each service chart accepts:

```yaml
podSecurityContext:
  runAsNonRoot: false
  runAsUser: 0
  runAsGroup: 0
containerSecurityContext:
  readOnlyRootFilesystem: false
```

These settings also apply to chart hooks. Cluster policy may reject root containers. The filesystem
and process identity are independent; disabling either restriction does not disable the other.
