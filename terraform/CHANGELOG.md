# Terraform Changelog

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
