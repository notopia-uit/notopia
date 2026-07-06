output "primary_endpoint" {
  value = module.elasticache.replication_group_primary_endpoint_address
}

output "reader_endpoint" {
  value = module.elasticache.replication_group_reader_endpoint_address
}

output "port" {
  value = module.elasticache.replication_group_port
}
