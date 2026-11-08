# billing

![Version: 0.1.1](https://img.shields.io/badge/Version-0.1.1-informational?style=flat-square)
## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../notopia-common-chart | notopia-common-chart | 0.1.0 |

## Migration and seeding

On install and upgrade, Helm creates the hook service account, runs the migration,
then runs the seed job before applying workload resources. Provide the database,
Redis, Stripe and all other referenced external secrets and image pull secrets
before starting Helm; the chart does not create or verify their readiness.

Seeding can create or change Stripe products, prices and portal configuration.
Coordinate releases, manual seeds and retries: hook ordering does not guarantee
safe concurrent or repeated executions, and failed releases do not undo Stripe
side effects. A failed migration or seed may have partially changed database or
Stripe state, even if the workload upgrade never occurs.

## Rollback compatibility

Rollback does not run these pre-upgrade hooks and can restore old code against
the new migration and seed state. Helm does not automatically roll back database
or Stripe state. Before upgrading or rolling back, verify the old code is
compatible with the resulting schema, data and Stripe configuration; plan and
perform any recovery explicitly before resuming service. Automatic rollback is
not a substitute for this compatibility check. Helm replaces the retained hook
service account before the next hook run; remove it manually after uninstall when
no longer needed.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Pod affinity |
| autoscaling.enabled | bool | `false` | Enable horizontal scaling |
| autoscaling.maxReplicas | int | `3` | Maximum replicas |
| autoscaling.minReplicas | int | `1` | Minimum replicas |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | CPU utilization target |
| containerSecurityContext.readOnlyRootFilesystem | bool | `true` | Mount the container root filesystem read-only |
| database.urlSecret | object | `{"key":"uri","name":""}` | Existing database URI secret |
| env | object | `{}` | Plain environment |
| envFrom | list | `[]` | External environment sources |
| extraEnv | list | `[]` | Additional environment entries |
| extraVolumeMounts | list | `[]` | Additional volume mounts |
| extraVolumes | list | `[]` | Additional volumes |
| fullnameOverride | string | `""` | Override release name |
| hooks.migrate.activeDeadlineSeconds | int | `300` | Migration deadline |
| hooks.migrate.resources | object | `{}` | Migration resources, defaults to pod resources; memory limit must be an integer quantity |
| hooks.seed.activeDeadlineSeconds | int | `300` | Seed deadline |
| hooks.seed.resources | object | `{}` | Seed resources, defaults to pod resources; memory limit must be an integer quantity |
| image.digest | string | `""` | Immutable image digest |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.repository | string | `"ghcr.io/notopia-uit/billing"` | Image repository |
| image.tag | string | `"latest"` | Image version |
| imagePullSecrets | list | `[]` | Image pull secrets |
| istio.allowedPrincipals | list | `[]` | Allowed SPIFFE principals |
| istio.enabled | bool | `true` | Enable Istio workload policy |
| istio.ports | list | `["18084","8084","28084"]` | Protected workload ports |
| nameOverride | string | `""` | Override chart name |
| nodeSelector | object | `{}` | Node selector |
| notopia-common-chart | object | `{"global":{}}` | Common library configuration |
| otel.enabled | bool | `true` | Enable OpenTelemetry environment |
| otel.endpoint | string | `""` | Collector endpoint |
| otel.protocol | string | `"grpc"` | Collector protocol |
| otel.resourceAttributes | object | `{}` | Additional resource attributes |
| podAnnotations | object | `{}` | Additional pod annotations |
| podDisruptionBudget.enabled | bool | `false` | Enable disruption budget |
| podDisruptionBudget.minAvailable | int | `1` | Minimum available replicas |
| podLabels | object | `{}` | Additional pod labels |
| podSecurityContext.fsGroup | int | `1000` | Group owning supported writable volumes |
| podSecurityContext.runAsGroup | int | `1000` | Numeric primary group ID |
| podSecurityContext.runAsNonRoot | bool | `true` | Require the container to run as a non-root user |
| podSecurityContext.runAsUser | int | `1000` | Numeric user ID; image files must be readable and executable by this user |
| redis.addr | string | `""` | Redis address |
| redis.db | int | `0` | Redis database |
| redis.passwordSecret | object | `{"key":"password","name":""}` | Existing Redis password secret |
| replicaCount | int | `1` | Number of replicas |
| resources.limits | object | `{"memory":"256Mi"}` | Container resource limits; memory must be an integer quantity (e.g. 256Mi or 2Gi), not fractional |
| resources.requests | object | `{"cpu":"50m","memory":"64Mi"}` | Container resource requests |
| route.enabled | bool | `true` | Create HTTPRoute |
| route.exposeHealth | bool | `true` | Expose health endpoints |
| route.hostnames | list | `[]` | Route hostnames |
| route.parentRefs | list | `[]` | Gateway parent references |
| runtime | string | `"go"` | Application runtime |
| serviceAccount.annotations | object | `{}` | Service account annotations |
| serviceAccount.create | bool | `true` | Create service account |
| serviceAccount.name | string | `""` | Existing or created account name |
| tolerations | list | `[]` | Pod tolerations |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
