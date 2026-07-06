output "ecs_task_execution_role_arn" {
  value = module.ecs_task_execution_role.arn
}

output "ecs_task_role_arn" {
  value = module.ecs_task_role.arn
}

output "alb_security_group_id" {
  value = aws_security_group.alb.id
}

output "ecs_security_group_id" {
  value = aws_security_group.ecs.id
}

output "rds_security_group_id" {
  value = aws_security_group.rds.id
}

output "elasticache_security_group_id" {
  value = aws_security_group.elasticache.id
}

output "msk_security_group_id" {
  value = aws_security_group.msk.id
}

output "meilisearch_security_group_id" {
  value = aws_security_group.meilisearch.id
}
