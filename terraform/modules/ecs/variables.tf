variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "aws_region" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "private_subnet_ids" {
  type = list(string)
}

variable "ecs_task_execution_role_arn" {
  type = string
}

variable "ecs_task_role_arn" {
  type = string
}

variable "ecs_security_group_id" {
  type = string
}

variable "ecr_repository_url" {
  type    = string
  default = "ghcr.io/notopia-uit"
}

variable "image_tags" {
  type = object({
    web           = string
    document      = string
    search_worker = string
    note          = string
    authorization = string
    api_web       = string
  })
}

variable "service_counts" {
  type = object({
    web           = number
    document      = number
    search_worker = number
    note          = number
    authorization = number
    authentik     = number
    meilisearch   = number
    api_web       = number
  })
}

variable "service_names" {
  type    = list(string)
  default = ["web", "document", "search-worker", "note", "authorization", "authentik", "meilisearch", "api-web"]
}

variable "domain_name" {
  type = string
}

# ALB
variable "https_listener_arn" {
  type = string
}

variable "web_target_group_arn" {
  type = string
}

variable "note_target_group_arn" {
  type = string
}

variable "document_target_group_arn" {
  type = string
}

variable "authz_target_group_arn" {
  type = string
}

variable "authentik_target_group_arn" {
  type = string
}

variable "meilisearch_target_group_arn" {
  type = string
}

variable "api_web_target_group_arn" {
  type = string
}

# Infrastructure
variable "redis_endpoint" {
  type = string
}

variable "redis_port" {
  type    = number
  default = 6379
}

variable "msk_bootstrap_brokers" {
  type = string
}

variable "s3_bucket_name" {
  type = string
}

# Per-service database configuration
variable "note_database" {
  type = object({
    host     = string
    port     = number
    name     = string
    username = string
    password = string
  })
  sensitive = true
}

variable "document_database" {
  type = object({
    host     = string
    port     = number
    name     = string
    username = string
    password = string
  })
  sensitive = true
}

variable "authorization_database" {
  type = object({
    host     = string
    port     = number
    name     = string
    username = string
    password = string
  })
  sensitive = true
}

variable "authentik_database" {
  type = object({
    host     = string
    port     = number
    name     = string
    username = string
    password = string
  })
  sensitive = true
}

variable "authentik_client_id" {
  type      = string
  sensitive = true
}

variable "authentik_secret" {
  type      = string
  sensitive = true
}

variable "authentik_token" {
  type      = string
  sensitive = true
}

variable "authentik_secret_key" {
  type      = string
  sensitive = true
}

variable "admin_email" {
  type = string
}

variable "admin_password" {
  type      = string
  sensitive = true
}

variable "meilisearch_api_key_ssm_arn" {
  type = string
}

# Monitoring
variable "xray_role_arn" {
  type    = string
  default = ""
}

variable "prometheus_remote_write_role_arn" {
  type    = string
  default = ""
}

variable "prometheus_workspace_endpoint" {
  type    = string
  default = ""
}
