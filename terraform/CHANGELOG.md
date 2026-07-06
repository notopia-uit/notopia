# Terraform Changelog

### 2026-07-06 — Migrate all modules to terraform-aws-modules

Replaced raw AWS resources with community modules across all 9 modules. This improves reliability (maintained by community), reduces boilerplate, and follows AWS best practices baked into each module.

**Modules migrated:**

| Module | Source | Version |
|---|---|---|
| networking | `terraform-aws-modules/vpc/aws` | `~> 6.0` |
| rds | `terraform-aws-modules/rds/aws` | `~> 7.2` |
| alb | `terraform-aws-modules/alb/aws` | latest |
| s3 | `terraform-aws-modules/s3-bucket/aws` | `~> 5.0` |
| elasticache | `terraform-aws-modules/elasticache/aws` | `~> 1.0` |
| msk | `terraform-aws-modules/msk-kafka-cluster/aws` | `~> 3.0` |
| security (IAM) | `terraform-aws-modules/iam/aws//modules/iam-role` | `~> 6.0` |
| ecs | `terraform-aws-modules/ecs/aws` | `~> 6.0` |
| monitoring | `terraform-aws-modules/cloudwatch/aws//modules/log-group` + `iam/aws` | `~> 5.0`, `~> 6.0` |

**Monitoring module specifics:**
- `aws_cloudwatch_log_group.ecs` → `module "ecs_log_group"` using cloudwatch log-group submodule
- `aws_iam_role.grafana` / `aws_iam_role.xray` / `aws_iam_role.prometheus_remote_write` → `module "grafana_role"` / `module "xray_role"` / `module "prometheus_remote_write_role"` using IAM role submodule
- `aws_cloudwatch_dashboard.main` and `aws_prometheus_workspace.main` remain as raw resources (no community module available)
- `aws_grafana_workspace.main` remains as raw resource

**Key interface changes:**
- ALB module creates its own security group (removed `security_group_id` variable)
- RDS uses `password_wo` (write-only) instead of `password`
- S3 uses `lifecycle_rule` instead of `lifecycle_configuration`
- IAM module outputs `arn` instead of `iam_role_arn`
- ECS cluster uses `module.ecs.cluster_id` instead of `aws_ecs_cluster.main.id`

---

### 2026-07-06 — Separate RDS instances per service

Refactored from a single Aurora PostgreSQL cluster to 4 independent RDS instances (one per service), matching the Docker Compose architecture. This provides full resource isolation — a compromised credential for one service cannot access another service's database.

**What changed:**

- **RDS module** (`modules/rds/`): Converted from Aurora cluster (`aws_rds_cluster` + `aws_rds_cluster_instance`) to standalone PostgreSQL instance (`aws_db_instance`). New `service_name` variable for resource naming. Parameter group family changed from `aurora-postgresql17` to `postgres17`. Added `allocated_storage`, `max_allocated_storage` variables. Added `multi_az` (prod only), `deletion_protection` (prod only), `backup_retention_period`.
- **ECS module** (`modules/ecs/`): Replaced shared `rds_endpoint`, `rds_port`, `db_username`, `db_password` variables with per-service database objects: `note_database`, `document_database`, `authorization_database`, `authentik_database`. Each is an `object({host, port, name, username, password})` marked `sensitive`.
- **Environment configs** (`environments/dev/main.tf`, `environments/prod/main.tf`): Single `module "rds"` replaced with 4 separate module calls: `rds_note`, `rds_document`, `rds_authorization`, `rds_authentik`. Each passes its own `service_name`, `database_name`, and connects to the ECS module via the corresponding `*_database` object.
- **Outputs** (`environments/*/outputs.tf`): Replaced `rds_endpoint` with per-service endpoints: `rds_note_endpoint`, `rds_document_endpoint`, `rds_authorization_endpoint`, `rds_authentik_endpoint`.

**Database mapping (matches Docker Compose):**

| Service | Module | Database Name | Docker Compose Equivalent |
|---|---|---|---|
| note | `rds_note` | `note` | `note_postgresql` (port 5433) |
| document | `rds_document` | `document` | `document_postgresql` (port 5434) |
| authorization | `rds_authorization` | `authorization` | `authorization_postgresql` (port 5439) |
| authentik | `rds_authentik` | `authentik` | `authentik_postgresql` |

**Cost impact:** 4x single instances cost roughly the same as 1 Aurora cluster + 1 instance. Dev uses `db.r6g.large` per instance; prod uses `db.r6g.xlarge`.

**Action required:** This is a destructive change for existing deployments. The old Aurora cluster will be destroyed and 4 new RDS instances created. Database data must be migrated manually.

---

### 2026-07-06 — Remove DynamoDB state locking, switch to S3-native lockfiles

DynamoDB-based state locking is deprecated by HashiCorp. Switched to `use_lockfile = true` on the S3 backend, which stores the lock as `<key>.tflock` in the same bucket — no extra AWS resource needed.

**Affected files:**
- `terraform/bootstrap/main.tf` — Removed `aws_dynamodb_table.tflock` resource and `dynamodb_table_name` output
- `terraform/environments/dev/versions.tf` — Replaced `dynamodb_table` with `use_lockfile = true`
- `terraform/environments/prod/versions.tf` — Replaced `dynamodb_table` with `use_lockfile = true`

**Action required:** Run `terraform destroy` on the bootstrap stack to remove the orphaned DynamoDB table, or manually delete it from AWS.
