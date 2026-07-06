locals {
  name_prefix = "${var.project_name}-${var.environment}"
}

module "ecs_task_execution_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role"
  version = "~> 6.0"

  name = "${local.name_prefix}-ecs-task-execution"

  trust_policy_permissions = {
    ECSTasks = {
      actions = ["sts:AssumeRole"]
      principals = [{
        type        = "Service"
        identifiers = ["ecs-tasks.amazonaws.com"]
      }]
    }
  }

  policies = {
    AmazonECSTaskExecutionRolePolicy = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
  }

  tags = {
    Name        = "${local.name_prefix}-ecs-task-execution"
    Environment = var.environment
  }
}

module "ecs_task_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role"
  version = "~> 6.0"

  name = "${local.name_prefix}-ecs-task"

  trust_policy_permissions = {
    ECSTasks = {
      actions = ["sts:AssumeRole"]
      principals = [{
        type        = "Service"
        identifiers = ["ecs-tasks.amazonaws.com"]
      }]
    }
  }

  tags = {
    Name        = "${local.name_prefix}-ecs-task"
    Environment = var.environment
  }
}

resource "aws_iam_role_policy" "ecs_task_s3" {
  name = "${local.name_prefix}-ecs-task-s3"
  role = module.ecs_task_role.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          "${var.s3_bucket_arn}/*",
          var.s3_bucket_arn
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy" "ecs_task_ssm" {
  name = "${local.name_prefix}-ecs-task-ssm"
  role = module.ecs_task_role.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameters",
          "ssm:GetParameter"
        ]
        Resource = "arn:aws:ssm:${var.aws_region}:${var.account_id}:parameter/${local.name_prefix}/*"
      }
    ]
  })
}

resource "aws_security_group" "alb" {
  name_prefix = "${local.name_prefix}-alb-"
  vpc_id      = var.vpc_id
  description = "ALB security group"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name        = "${local.name_prefix}-alb"
    Environment = var.environment
  }
}

resource "aws_security_group" "ecs" {
  name_prefix = "${local.name_prefix}-ecs-"
  vpc_id      = var.vpc_id
  description = "ECS tasks security group"

  ingress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    security_groups = [aws_security_group.alb.id]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    self        = true
    description = "Allow traffic between ECS tasks"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name        = "${local.name_prefix}-ecs"
    Environment = var.environment
  }
}

resource "aws_security_group" "rds" {
  name_prefix = "${local.name_prefix}-rds-"
  vpc_id      = var.vpc_id
  description = "RDS security group"

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name        = "${local.name_prefix}-rds"
    Environment = var.environment
  }
}

resource "aws_security_group" "elasticache" {
  name_prefix = "${local.name_prefix}-redis-"
  vpc_id      = var.vpc_id
  description = "ElastiCache security group"

  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name        = "${local.name_prefix}-redis"
    Environment = var.environment
  }
}

resource "aws_security_group" "msk" {
  name_prefix = "${local.name_prefix}-msk-"
  vpc_id      = var.vpc_id
  description = "MSK security group"

  ingress {
    from_port       = 9092
    to_port         = 9094
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs.id]
  }

  ingress {
    from_port       = 9098
    to_port         = 9098
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name        = "${local.name_prefix}-msk"
    Environment = var.environment
  }
}

resource "aws_security_group" "meilisearch" {
  name_prefix = "${local.name_prefix}-meili-"
  vpc_id      = var.vpc_id
  description = "Meilisearch security group"

  ingress {
    from_port       = 7700
    to_port         = 7700
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name        = "${local.name_prefix}-meili"
    Environment = var.environment
  }
}
