module "ecs_log_group" {
  source  = "terraform-aws-modules/cloudwatch/aws//modules/log-group"
  version = "~> 5.0"

  name              = "/aws/ecs/${var.project_name}-${var.environment}"
  retention_in_days = var.log_retention_days

  tags = {
    Name        = "${var.project_name}-${var.environment}-ecs-logs"
    Environment = var.environment
  }
}

resource "aws_prometheus_workspace" "main" {
  alias = "${var.project_name}-${var.environment}"

  tags = {
    Name        = "${var.project_name}-${var.environment}-amp"
    Environment = var.environment
  }
}

module "grafana_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role"
  version = "~> 6.0"

  name = "${var.project_name}-${var.environment}-grafana"

  trust_policy_permissions = {
    Grafana = {
      actions = ["sts:AssumeRole"]
      principals = [{
        type        = "Service"
        identifiers = ["grafana.amazonaws.com"]
      }]
    }
  }

  policies = {
    AdministratorAccess = "arn:aws:iam::aws:policy/AdministratorAccess"
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-grafana"
    Environment = var.environment
  }
}

resource "aws_grafana_workspace" "main" {
  name                     = "${var.project_name}-${var.environment}"
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["AWS_SSO"]
  permission_type          = "SERVICE_MANAGED"
  role_arn                 = module.grafana_role.arn
  grafana_version          = "11.4.0"

  data_sources              = ["PROMETHEUS", "CLOUDWATCH"]
  notification_destinations = ["SNS"]

  tags = {
    Name        = "${var.project_name}-${var.environment}-grafana"
    Environment = var.environment
  }
}

module "xray_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role"
  version = "~> 6.0"

  name = "${var.project_name}-${var.environment}-xray"

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
    XRay = aws_iam_policy.xray.arn
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-xray"
    Environment = var.environment
  }
}

resource "aws_iam_policy" "xray" {
  name        = "${var.project_name}-${var.environment}-xray"
  path        = "/"
  description = "X-Ray write access"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "xray:PutTraceSegments",
        "xray:PutTelemetryRecords",
        "xray:GetSamplingRules",
        "xray:GetSamplingTargets",
        "xray:GetSamplingStatisticSummaries"
      ]
      Resource = "*"
    }]
  })
}

module "prometheus_remote_write_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role"
  version = "~> 6.0"

  name = "${var.project_name}-${var.environment}-prom-rw"

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
    PrometheusRemoteWrite = aws_iam_policy.prometheus_remote_write.arn
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}-prom-rw"
    Environment = var.environment
  }
}

resource "aws_iam_policy" "prometheus_remote_write" {
  name        = "${var.project_name}-${var.environment}-prom-rw"
  path        = "/"
  description = "Prometheus remote write to AMP"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "aps:RemoteWrite",
        "aps:GetSeries",
        "aps:GetLabels",
        "aps:GetMetricMetadata"
      ]
      Resource = aws_prometheus_workspace.main.arn
    }]
  })
}

resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "${var.project_name}-${var.environment}"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "text"
        x      = 0
        y      = 0
        width  = 24
        height = 1
        properties = {
          markdown = "# Notopia - ${title(var.environment)} Environment"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 1
        width  = 12
        height = 6
        properties = {
          title  = "ECS CPU Utilization"
          region = var.aws_region
          metrics = [
            ["AWS/ECS", "CPUUtilization", "ClusterName", "${var.project_name}-${var.environment}"]
          ]
          period = 300
          stat   = "Average"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 1
        width  = 12
        height = 6
        properties = {
          title  = "ECS Memory Utilization"
          region = var.aws_region
          metrics = [
            ["AWS/ECS", "MemoryUtilization", "ClusterName", "${var.project_name}-${var.environment}"]
          ]
          period = 300
          stat   = "Average"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 7
        width  = 12
        height = 6
        properties = {
          title  = "ALB Request Count"
          region = var.aws_region
          metrics = [
            ["AWS/ApplicationELB", "RequestCount", "LoadBalancer", var.alb_arn_suffix]
          ]
          period = 300
          stat   = "Sum"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 7
        width  = 12
        height = 6
        properties = {
          title  = "ALB Target Response Time"
          region = var.aws_region
          metrics = [
            ["AWS/ApplicationELB", "TargetResponseTime", "LoadBalancer", var.alb_arn_suffix]
          ]
          period = 300
          stat   = "Average"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 13
        width  = 12
        height = 6
        properties = {
          title  = "RDS CPU Utilization"
          region = var.aws_region
          metrics = [
            ["AWS/RDS", "CPUUtilization", "DBClusterIdentifier", "${var.project_name}-${var.environment}"]
          ]
          period = 300
          stat   = "Average"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 13
        width  = 12
        height = 6
        properties = {
          title  = "RDS Database Connections"
          region = var.aws_region
          metrics = [
            ["AWS/RDS", "DatabaseConnections", "DBClusterIdentifier", "${var.project_name}-${var.environment}"]
          ]
          period = 300
          stat   = "Average"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 19
        width  = 12
        height = 6
        properties = {
          title  = "ElastiCache CPU Utilization"
          region = var.aws_region
          metrics = [
            ["AWS/ElastiCache", "CPUUtilization", "CacheClusterId", "${var.project_name}-${var.environment}-001"]
          ]
          period = 300
          stat   = "Average"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 19
        width  = 12
        height = 6
        properties = {
          title  = "ElastiCache Cache Hits"
          region = var.aws_region
          metrics = [
            ["AWS/ElastiCache", "CacheHits", "CacheClusterId", "${var.project_name}-${var.environment}-001"]
          ]
          period = 300
          stat   = "Average"
        }
      },
      {
        type   = "log"
        x      = 0
        y      = 25
        width  = 24
        height = 6
        properties = {
          title  = "ECS Service Logs"
          region = var.aws_region
          query  = "fields @timestamp, @message | sort @timestamp desc | limit 50"
          stack  = []
        }
      }
    ]
  })
}
