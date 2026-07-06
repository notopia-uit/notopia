# Terraform Changelog

### 2026-07-06 — Remove DynamoDB state locking, switch to S3-native lockfiles

DynamoDB-based state locking is deprecated by HashiCorp. Switched to `use_lockfile = true` on the S3 backend, which stores the lock as `<key>.tflock` in the same bucket — no extra AWS resource needed.

**Affected files:**
- `terraform/bootstrap/main.tf` — Removed `aws_dynamodb_table.tflock` resource and `dynamodb_table_name` output
- `terraform/environments/dev/versions.tf` — Replaced `dynamodb_table` with `use_lockfile = true`
- `terraform/environments/prod/versions.tf` — Replaced `dynamodb_table` with `use_lockfile = true`

**Action required:** Run `terraform destroy` on the bootstrap stack to remove the orphaned DynamoDB table, or manually delete it from AWS.
