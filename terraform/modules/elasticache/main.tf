module "elasticache" {
  source  = "terraform-aws-modules/elasticache/aws"
  version = "~> 1.0"

  replication_group_id = "${var.project_name}-${var.environment}-redis"
  description          = "Redis cluster for ${var.project_name} ${var.environment}"

  engine             = "redis"
  engine_version     = "7.1"
  node_type          = var.node_type
  num_cache_clusters = var.num_cache_nodes
  port               = 6379
  subnet_ids         = var.private_subnet_ids

  at_rest_encryption_enabled = true
  transit_encryption_enabled = true

  automatic_failover_enabled = var.num_cache_nodes > 1
  multi_az_enabled           = var.num_cache_nodes > 1

  create_security_group = false
  security_group_ids    = [var.security_group_id]

  tags = {
    Name        = "${var.project_name}-${var.environment}-redis"
    Environment = var.environment
  }
}
