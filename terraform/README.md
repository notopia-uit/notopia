# Notopia — Terraform Infrastructure

AWS infrastructure for Notopia, managed with Terraform and organized by environment.

## Structure

```
terraform/
├── bootstrap/                  # One-time setup: S3 bucket + DynamoDB table for state
│   └── main.tf
├── modules/                    # Reusable infrastructure modules
│   ├── networking/             # VPC, subnets, NAT gateways, internet gateway
│   ├── security/               # IAM roles, security groups
│   ├── rds/                    # Aurora PostgreSQL 17 cluster
│   ├── elasticache/            # Redis replication group
│   ├── msk/                    # Amazon MSK (Kafka 3.7.1)
│   ├── s3/                     # S3 bucket for document attachments
│   ├── alb/                    # Application Load Balancer + target groups
│   ├── ecs/                    # ECS Fargate cluster, task definitions, services
│   └── monitoring/             # AMP, Grafana, CloudWatch dashboard, X-Ray/Prometheus IAM
└── environments/
    ├── dev/                    # Development environment
    │   ├── versions.tf         # Provider + S3 backend config
    │   ├── data.tf             # Data sources (account ID, ACM cert, AZs)
    │   ├── main.tf             # Module wiring
    │   ├── variables.tf        # Input variables
    │   ├── outputs.tf          # Output values
    │   └── terraform.tfvars    # Variable values (secrets go here)
    └── prod/                   # Production environment
        └── (same structure)
```

## Prerequisites

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.7
- AWS CLI configured with appropriate credentials
- An ACM certificate for `*.notopia.app` (or your domain) in `ap-southeast-1`

## Quick Start

### 1. Bootstrap state storage (one-time)

The Terraform state is stored in S3 with DynamoDB locking. Create these resources first:

```bash
cd terraform/bootstrap
terraform init
terraform apply
```

This creates:

- **S3 bucket** `notopia-tfstate` — encrypted, versioned, public access blocked
- **DynamoDB table** `notopia-tflock` — PAY_PER_REQUEST, point-in-time recovery

### 2. Deploy an environment

```bash
cd terraform/environments/dev   # or prod
```

Create/edit `terraform.tfvars` with your values:

```hcl
domain_name     = "dev.notopia.app"
certificate_arn = "arn:aws:acm:ap-southeast-1:123456789012:certificate/abc-123"

image_tags = {
  web           = "latest"
  document      = "latest"
  search_worker = "latest"
  note          = "latest"
  authorization = "latest"
  api_web       = "latest"
}

service_counts = {
  web           = 1
  document      = 1
  search_worker = 1
  note          = 1
  authorization = 1
  authentik     = 1
  meilisearch   = 1
  api_web       = 1
}

db_username        = "notopia"
db_password        = "CHANGE_ME"
authentik_client_id = "CHANGE_ME"
authentik_secret    = "CHANGE_ME"
authentik_token     = "CHANGE_ME"
authentik_secret_key = "CHANGE_ME"
admin_email         = "admin@notopia.app"
admin_password      = "CHANGE_ME"
meilisearch_api_key = "CHANGE_ME_IN_SSM"
```

Then:

```bash
terraform init
terraform plan
terraform apply
```

### 3. Access services

After apply, get the ALB DNS:

```bash
terraform output alb_dns_name
```

Create a CNAME record in your DNS:

- `dev.notopia.app` → `<alb_dns_name>`

Services will be available at:
| Path | Service |
|---|---|
| `/` | Web (Next.js) |
| `/note*` | Note API (Go) |
| `/document*` | Document API (NestJS) |
| `/authorization*` | Authorization API (Go) |
| `/-/authentik*` | Authentik (IdP) |
| `/search*`, `/indexes*` | Meilisearch |
| `/docs` | API documentation (Scalar) |

## Dev vs Prod Differences

| Setting                  | Dev               | Prod               |
| ------------------------ | ----------------- | ------------------ |
| NAT Gateways             | 1 (cost saving)   | 1 per AZ (3 total) |
| RDS instance class       | `db.r6g.large`    | `db.r6g.xlarge`    |
| RDS instances            | 1                 | 2 (Multi-AZ)       |
| Redis node type          | `cache.r6g.large` | `cache.r6g.xlarge` |
| Redis nodes              | 1                 | 2 (failover)       |
| MSK brokers              | 1                 | 3                  |
| MSK instance             | `kafka.m5.large`  | `kafka.m5.xlarge`  |
| MSK EBS                  | 50 GB             | 200 GB             |
| Service desired count    | 1 each            | 2 each             |
| ALB deletion protection  | disabled          | enabled            |
| CloudWatch log retention | 7 days            | 30 days            |

## Services Deployed

| Service           | Tech      | Port         | Health Check                |
| ----------------- | --------- | ------------ | --------------------------- |
| **web**           | Next.js   | 3000         | `GET /`                     |
| **note**          | Go        | 8081         | `GET /healthz` (port 28081) |
| **document**      | NestJS    | 8082         | `GET /healthz`              |
| **authorization** | Go        | 18089        | `GET /healthz` (port 28089) |
| **search-worker** | NestJS    | — (internal) | `GET /healthz` (port 28083) |
| **authentik**     | Python    | 9000         | `GET /-/health/ready`       |
| **meilisearch**   | Rust      | 7700         | `GET /health`               |
| **api-web**       | React SPA | 9080         | `GET /`                     |

All services run on **ECS Fargate** with **Cloud Map** service discovery (`*.notopia.local`).

## Infrastructure

| Resource              | Details                                                    |
| --------------------- | ---------------------------------------------------------- |
| **VPC**               | 10.0.0.0/16, 3 AZs, public + private subnets               |
| **Aurora PostgreSQL** | 17.4, encrypted, per-service databases                     |
| **Redis**             | Encryption at rest + in transit                            |
| **MSK (Kafka)**       | 3.7.1, TLS in transit, KMS encryption                      |
| **S3**                | Versioned, KMS encrypted, public access blocked            |
| **ALB**               | TLS 1.3, HTTP→HTTPS redirect                               |
| **AMP**               | Prometheus metrics from ECS tasks                          |
| **Grafana**           | Managed workspace with Prometheus + CloudWatch datasources |
| **X-Ray**             | Distributed tracing via OTel SDK                           |

## Observability

- **OpenTelemetry** is enabled on all application services (`OTEL_SDK_DISABLED=false`)
- Metrics export to **Amazon Managed Prometheus** via OTLP
- Traces export to **AWS X-Ray**
- Logs go to **CloudWatch** per service
- **Grafana** workspace provides dashboards with Prometheus + CloudWatch data sources
- **CloudWatch Dashboard** with ECS/ALB/RDS/ElastiCache widgets

## State Management

- State stored in S3 bucket `notopia-tfstate`
- State locking via DynamoDB table `notopia-tflock`
- Each environment has its own state key:
  - Dev: `dev/terraform.tfstate`
  - Prod: `prod/terraform.tfstate`
- State is encrypted at rest

## Common Commands

```bash
# Initialize
terraform init

# Preview changes
terraform plan

# Apply changes
terraform apply

# Destroy all resources
terraform destroy

# View outputs
terraform output

# Format HCL files
terraform fmt -recursive

# Validate configuration
terraform validate
```

## Secrets

Sensitive values (`db_password`, `authentik_*`, `admin_password`, `meilisearch_api_key`) should be set in `terraform.tfvars` which is gitignored. Never commit real secrets.

For production, consider using:

- AWS Secrets Manager for database credentials
- SSM Parameter Store for API keys (already used for Meilisearch key)
- Environment variable injection via CI/CD pipeline

## CI/CD Integration

When deploying via CI/CD:

1. Configure AWS credentials (OIDC or IAM role)
2. Run `terraform init -backend-config="key=$ENVIRONMENT/terraform.tfstate"`
3. Run `terraform plan -out=tfplan`
4. Run `terraform apply tfplan`

Set sensitive variables via environment variables:

```bash
export TF_VAR_db_password="$DB_PASSWORD"
export TF_VAR_authentik_client_id="$AUTHENTIK_CLIENT_ID"
# ... etc
```
