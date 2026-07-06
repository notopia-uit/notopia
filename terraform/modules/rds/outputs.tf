output "endpoint" {
  value = module.rds.db_instance_address
}

output "port" {
  value = module.rds.db_instance_port
}

output "database_name" {
  value = module.rds.db_instance_name
}

output "instance_id" {
  value = module.rds.db_instance_identifier
}
