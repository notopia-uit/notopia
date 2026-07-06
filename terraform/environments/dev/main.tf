locals {
  name_prefix     = "notopia-dev"
  certificate_arn = var.certificate_arn != "" ? var.certificate_arn : data.aws_acm_certificate.main[0].arn
}

module "networking" {
  source = "../../modules/networking"

  project_name       = "notopia"
  environment        = "dev"
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  single_nat_gateway = true

  public_subnet_cidrs  = [for i, az in var.availability_zones : cidrsubnet(var.vpc_cidr, 8, i + 1)]
  private_subnet_cidrs = [for i, az in var.availability_zones : cidrsubnet(var.vpc_cidr, 8, i + 11)]
}

module "s3" {
  source = "../../modules/s3"

  project_name = "notopia"
  environment  = "dev"
}

module "security" {
  source = "../../modules/security"

  project_name  = "notopia"
  environment   = "dev"
  aws_region    = var.aws_region
  account_id    = data.aws_caller_identity.current.account_id
  vpc_id        = module.networking.vpc_id
  s3_bucket_arn = module.s3.bucket_arn
}

module "rds_note" {
  source = "../../modules/rds"

  project_name       = "notopia"
  environment        = "dev"
  service_name       = "note"
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_id  = module.security.rds_security_group_id
  database_name      = "note"
  db_username        = var.db_username
  db_password        = var.db_password
  instance_class     = "db.r6g.large"
}

module "rds_document" {
  source = "../../modules/rds"

  project_name       = "notopia"
  environment        = "dev"
  service_name       = "document"
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_id  = module.security.rds_security_group_id
  database_name      = "document"
  db_username        = var.db_username
  db_password        = var.db_password
  instance_class     = "db.r6g.large"
}

module "rds_authorization" {
  source = "../../modules/rds"

  project_name       = "notopia"
  environment        = "dev"
  service_name       = "authorization"
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_id  = module.security.rds_security_group_id
  database_name      = "authorization"
  db_username        = var.db_username
  db_password        = var.db_password
  instance_class     = "db.r6g.large"
}

module "rds_authentik" {
  source = "../../modules/rds"

  project_name       = "notopia"
  environment        = "dev"
  service_name       = "authentik"
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_id  = module.security.rds_security_group_id
  database_name      = "authentik"
  db_username        = var.db_username
  db_password        = var.db_password
  instance_class     = "db.r6g.large"
}

module "elasticache" {
  source = "../../modules/elasticache"

  project_name       = "notopia"
  environment        = "dev"
  private_subnet_ids = module.networking.private_subnet_ids
  security_group_id  = module.security.elasticache_security_group_id
  node_type          = "cache.r6g.large"
  num_cache_nodes    = 1
}

module "msk" {
  source = "../../modules/msk"

  project_name         = "notopia"
  environment          = "dev"
  private_subnet_ids   = module.networking.private_subnet_ids
  security_group_id    = module.security.msk_security_group_id
  broker_instance_type = "kafka.m5.large"
  broker_count         = 1
  ebs_volume_size      = 50
}

module "alb" {
  source = "../../modules/alb"

  project_name      = "notopia"
  environment       = "dev"
  vpc_id            = module.networking.vpc_id
  public_subnet_ids = module.networking.public_subnet_ids
  certificate_arn   = local.certificate_arn
}

resource "aws_ssm_parameter" "meilisearch_key" {
  name  = "/notopia/dev/meilisearch-api-key"
  type  = "SecureString"
  value = var.meilisearch_api_key

  tags = {
    Name        = "notopia-dev-meilisearch-key"
    Environment = "dev"
  }
}

module "ecs" {
  source = "../../modules/ecs"

  project_name                = "notopia"
  environment                 = "dev"
  aws_region                  = var.aws_region
  vpc_id                      = module.networking.vpc_id
  private_subnet_ids          = module.networking.private_subnet_ids
  ecs_task_execution_role_arn = module.security.ecs_task_execution_role_arn
  ecs_task_role_arn           = module.security.ecs_task_role_arn
  ecs_security_group_id       = module.security.ecs_security_group_id
  ecr_repository_url          = var.ecr_repository_url
  image_tags                  = var.image_tags
  service_counts              = var.service_counts
  domain_name                 = var.domain_name

  https_listener_arn             = module.alb.https_listener_arn
  web_target_group_arn           = module.alb.web_target_group_arn
  note_target_group_arn          = module.alb.note_target_group_arn
  document_target_group_arn      = module.alb.document_target_group_arn
  authorization_target_group_arn = module.alb.authorization_target_group_arn
  authentik_target_group_arn     = module.alb.authentik_target_group_arn
  meilisearch_target_group_arn   = module.alb.meilisearch_target_group_arn
  api_web_target_group_arn       = module.alb.api_web_target_group_arn

  redis_endpoint        = module.elasticache.primary_endpoint
  redis_port            = module.elasticache.port
  msk_bootstrap_brokers = module.msk.bootstrap_brokers_tls
  s3_bucket_name        = module.s3.bucket_id

  note_database = {
    host     = module.rds_note.endpoint
    port     = module.rds_note.port
    name     = module.rds_note.database_name
    username = var.db_username
    password = var.db_password
  }

  document_database = {
    host     = module.rds_document.endpoint
    port     = module.rds_document.port
    name     = module.rds_document.database_name
    username = var.db_username
    password = var.db_password
  }

  authorization_database = {
    host     = module.rds_authorization.endpoint
    port     = module.rds_authorization.port
    name     = module.rds_authorization.database_name
    username = var.db_username
    password = var.db_password
  }

  authentik_database = {
    host     = module.rds_authentik.endpoint
    port     = module.rds_authentik.port
    name     = module.rds_authentik.database_name
    username = var.db_username
    password = var.db_password
  }

  authentik_client_id         = var.authentik_client_id
  authentik_secret            = var.authentik_secret
  authentik_token             = var.authentik_token
  authentik_secret_key        = var.authentik_secret_key
  admin_email                 = var.admin_email
  admin_password              = var.admin_password
  meilisearch_api_key_ssm_arn = aws_ssm_parameter.meilisearch_key.arn

  xray_role_arn                    = module.monitoring.xray_role_arn
  prometheus_remote_write_role_arn = module.monitoring.prometheus_remote_write_role_arn
  prometheus_workspace_endpoint    = module.monitoring.prometheus_workspace_endpoint
}

module "monitoring" {
  source = "../../modules/monitoring"

  project_name       = "notopia"
  environment        = "dev"
  aws_region         = var.aws_region
  alb_arn_suffix     = module.alb.alb_arn_suffix
  log_retention_days = 7
}
