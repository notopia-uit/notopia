output "cluster_id" {
  value = module.ecs.cluster_id
}

output "cluster_arn" {
  value = module.ecs.cluster_arn
}

output "service_discovery_namespace_id" {
  value = aws_service_discovery_private_dns_namespace.main.id
}
