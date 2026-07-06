module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 7.2"

  identifier = "${var.project_name}-${var.environment}-${var.service_name}"

  engine               = "postgres"
  engine_version       = "17.4"
  family               = "postgres17"
  major_engine_version = "17"
  instance_class       = var.instance_class

  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name     = var.database_name
  username    = var.db_username
  password_wo = var.db_password

  create_db_subnet_group = true
  subnet_ids             = var.private_subnet_ids

  vpc_security_group_ids = [var.security_group_id]
  publicly_accessible    = false

  multi_az = var.environment == "prod"

  parameter_group_name = null
  parameters = [
    {
      name  = "log_connections"
      value = "1"
    },
    {
      name  = "log_disconnections"
      value = "1"
    }
  ]

  skip_final_snapshot              = var.environment != "prod"
  deletion_protection              = var.environment == "prod"
  final_snapshot_identifier_prefix = var.environment == "prod" ? "${var.project_name}-${var.environment}-${var.service_name}-final" : null

  backup_retention_period = var.environment == "prod" ? 7 : 1

  tags = {
    Name        = "${var.project_name}-${var.environment}-${var.service_name}"
    Environment = var.environment
  }
}
