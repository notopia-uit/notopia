output "vpc_id" {
  value = module.networking.vpc_id
}

output "alb_dns_name" {
  value = module.alb.alb_dns_name
}

output "ecs_cluster_id" {
  value = module.ecs.cluster_id
}

output "rds_note_endpoint" {
  value = module.rds_note.endpoint
}

output "rds_document_endpoint" {
  value = module.rds_document.endpoint
}

output "rds_authorization_endpoint" {
  value = module.rds_authorization.endpoint
}

output "rds_authentik_endpoint" {
  value = module.rds_authentik.endpoint
}

output "redis_endpoint" {
  value = module.elasticache.primary_endpoint
}

output "msk_bootstrap_brokers" {
  value     = module.msk.bootstrap_brokers_tls
  sensitive = true
}

output "s3_bucket_name" {
  value = module.s3.bucket_id
}

output "prometheus_workspace_endpoint" {
  value = module.monitoring.prometheus_workspace_endpoint
}

output "grafana_workspace_endpoint" {
  value = module.monitoring.grafana_workspace_endpoint
}

output "cloudwatch_log_group" {
  value = "/aws/ecs/notopia-prod"
}
