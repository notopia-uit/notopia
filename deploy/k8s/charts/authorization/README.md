# authorization

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square)
## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../notopia-common-chart | notopia-common-chart | 0.1.0 |

## Required external policy seeding

**Before every install and upgrade**, an operator must seed the database using
`./authorization seed` from the same immutable image that will serve traffic,
with the same database connection and external secrets. Wait for that command
to succeed before starting the release. The chart's post-install/post-upgrade
seed hook is not early enough for `helm --wait`, and policy is loaded only at
process startup.

For an upgrade, schedule a maintenance window: stop all authorization pods and
block traffic, suspend autoscaling and any controller that could recreate them,
and wait for termination before running exactly one external seed. Ensure no
other policy seeder runs concurrently. After seeding succeeds, deploy the same
image, restore the intended replica count and autoscaling, and verify every
replacement pod is healthy before restoring traffic. On rollback, repeat this
procedure using the rollback image's policy; do not run the old and new policy
seeders concurrently.

The chart does not verify that the external seed ran or that the database policy
matches the selected image. This manual procedure causes downtime and is not an
automated upgrade guarantee. Failed or omitted seeding can leave Helm waiting or
leave pods with stale authorization rules. Do not proceed without coordinating
the database, deployment controller, and traffic controls.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Pod affinity |
| autoscaling.enabled | bool | `false` | Enable horizontal scaling |
| autoscaling.maxReplicas | int | `3` | Maximum replicas |
| autoscaling.minReplicas | int | `1` | Minimum replicas |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | CPU utilization target |
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
| image.repository | string | `"ghcr.io/notopia-uit/authorization"` | Image repository |
| image.tag | string | `"latest"` | Image version |
| imagePullSecrets | list | `[]` | Image pull secrets |
| istio.allowedPrincipals | list | `[]` | Allowed SPIFFE principals |
| istio.enabled | bool | `true` | Enable Istio workload policy |
| istio.ports | list | `["18089","28089"]` | Protected workload ports |
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
| replicaCount | int | `1` | Number of replicas |
| resources.limits | object | `{"memory":"256Mi"}` | Container resource limits; memory must be an integer quantity (e.g. 256Mi or 2Gi), not fractional |
| resources.requests | object | `{"cpu":"50m","memory":"64Mi"}` | Container resource requests |
| route.enabled | bool | `false` | Create HTTPRoute |
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
