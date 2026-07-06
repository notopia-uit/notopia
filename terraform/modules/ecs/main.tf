resource "aws_cloudwatch_log_group" "services" {
  for_each          = toset(var.service_names)
  name              = "/ecs/${var.project_name}-${var.environment}/${each.value}"
  retention_in_days = 30

  tags = {
    Name        = "${var.project_name}-${var.environment}-${each.value}"
    Environment = var.environment
  }
}

resource "aws_ecs_cluster" "main" {
  name = "${var.project_name}-${var.environment}"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}"
    Environment = var.environment
  }
}

resource "aws_service_discovery_private_dns_namespace" "main" {
  name = "${var.environment}.notopia.local"
  vpc  = var.vpc_id

  tags = {
    Name        = "${var.project_name}-${var.environment}-sd"
    Environment = var.environment
  }
}

resource "aws_service_discovery_service" "services" {
  for_each = toset(var.service_names)

  name = each.value

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
  }
}

locals {
  common_environment = [
    { name = "OTEL_SDK_DISABLED", value = "false" },
    { name = "OTEL_EXPORTER_OTLP_ENDPOINT", value = var.prometheus_workspace_endpoint != "" ? var.prometheus_workspace_endpoint : "" },
    { name = "OTEL_EXPORTER_OTLP_PROTOCOL", value = "http/protobuf" },
    { name = "TZ", value = "UTC" },
  ]
}

# ──────────────────────────────────────────────
# Web (Next.js)
# ──────────────────────────────────────────────
resource "aws_ecs_task_definition" "web" {
  family                   = "${var.project_name}-${var.environment}-web"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 512
  memory                   = 1024
  execution_role_arn       = var.ecs_task_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn

  container_definitions = jsonencode([{
    name      = "web"
    image     = "${var.ecr_repository_url}/web:${var.image_tags.web}"
    essential = true

    portMappings = [{
      containerPort = 3000
      protocol      = "tcp"
    }]

    environment = concat(local.common_environment, [
      { name = "OTEL_SERVICE_NAME", value = "notopia-${var.environment}-web" },
      { name = "BETTER_AUTH_URL", value = "https://${var.domain_name}" },
      { name = "NEXT_PUBLIC_BETTER_AUTH_URL", value = "https://${var.domain_name}" },
      { name = "NEXT_PUBLIC_API_URL", value = "https://${var.domain_name}" },
      { name = "API_URL", value = "http://${aws_service_discovery_service.services["note"].name}.${aws_service_discovery_private_dns_namespace.main.name}:8081" },
      { name = "NEXT_PUBLIC_MEILISEARCH_HOST", value = "https://${var.domain_name}" },
      { name = "MEILISEARCH_HOST", value = "http://${aws_service_discovery_service.services["meilisearch"].name}.${aws_service_discovery_private_dns_namespace.main.name}:7700" },
      { name = "AUTHENTIK_CLIENT_ID", value = var.authentik_client_id },
      { name = "AUTHENTIK_SECRET", value = var.authentik_secret },
      { name = "AUTHENTIK_DISCOVERY_URL", value = "https://auth.${var.domain_name}/application/o/.well-known/openid-configuration" },
      { name = "AUTHENTIK_REDIRECT_URI", value = "https://${var.domain_name}/api/auth/callback/authentik" },
    ])

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["web"].name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "web"
      }
    }

    healthCheck = {
      command     = ["CMD-SHELL", "wget -qO- http://localhost:3000/ || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 60
    }
  }])

  tags = {
    Name        = "${var.project_name}-${var.environment}-web"
    Environment = var.environment
  }
}

resource "aws_ecs_service" "web" {
  name            = "web"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.web.arn
  desired_count   = var.service_counts.web
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.web_target_group_arn
    container_name   = "web"
    container_port   = 3000
  }

  service_registries {
    registry_arn = aws_service_discovery_service.services["web"].arn
  }

  depends_on = [aws_lb_listener_rule.web]
}

resource "aws_lb_listener_rule" "web" {
  listener_arn = var.https_listener_arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = var.web_target_group_arn
  }

  condition {
    path_pattern {
      values = ["/"]
    }
  }
}

# ──────────────────────────────────────────────
# Note (Go)
# ──────────────────────────────────────────────
resource "aws_ecs_task_definition" "note" {
  family                   = "${var.project_name}-${var.environment}-note"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = var.ecs_task_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn

  container_definitions = jsonencode([{
    name      = "note"
    image     = "${var.ecr_repository_url}/note:${var.image_tags.note}"
    essential = true

    portMappings = [{
      containerPort = 8081
      protocol      = "tcp"
    }]

    environment = concat(local.common_environment, [
      { name = "OTEL_SERVICE_NAME", value = "notopia-${var.environment}-note" },
      { name = "NOTOPIA_NOTE_GENERAL_APP_ENV", value = var.environment },
      { name = "NOTOPIA_NOTE_LOG_LEVEL", value = "info" },
      { name = "NOTOPIA_NOTE_SERVER_URL", value = "https://${var.domain_name}" },
      { name = "NOTOPIA_NOTE_DATABASE_URL", value = "postgresql://${var.note_database.username}:${var.note_database.password}@${var.note_database.host}:${var.note_database.port}/${var.note_database.name}" },
      { name = "NOTOPIA_NOTE_KAFKA_BROKERS", value = var.msk_bootstrap_brokers },
      { name = "NOTOPIA_NOTE_KAFKA_CONSUMER_GROUP", value = "note-${var.environment}" },
      { name = "NOTOPIA_NOTE_REDIS_ADDR", value = "${var.redis_endpoint}:${var.redis_port}" },
      { name = "NOTOPIA_NOTE_SERVICES_AUTHORIZATION_URL", value = "${aws_service_discovery_service.services["authorization"].name}.${aws_service_discovery_private_dns_namespace.main.name}:18089" },
      { name = "NOTOPIA_NOTE_SERVICES_AUTHORIZATION_LIVE_URL", value = "${aws_service_discovery_service.services["authorization"].name}.${aws_service_discovery_private_dns_namespace.main.name}:28089" },
      { name = "NOTOPIA_NOTE_MEILISEARCH_HOST", value = "http://${aws_service_discovery_service.services["meilisearch"].name}.${aws_service_discovery_private_dns_namespace.main.name}:7700" },
      { name = "NOTOPIA_NOTE_AUTHENTIK_HOST", value = "https://auth.${var.domain_name}" },
      { name = "NOTOPIA_NOTE_AUTHENTIK_TOKEN", value = var.authentik_token },
    ])

    secrets = [
      { name = "NOTOPIA_NOTE_MEILISEARCH_API_KEY", valueFrom = var.meilisearch_api_key_ssm_arn },
    ]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["note"].name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "note"
      }
    }

    healthCheck = {
      command     = ["CMD-SHELL", "wget -qO- http://localhost:28081/healthz || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 30
    }
  }])

  tags = {
    Name        = "${var.project_name}-${var.environment}-note"
    Environment = var.environment
  }
}

resource "aws_ecs_service" "note" {
  name            = "note"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.note.arn
  desired_count   = var.service_counts.note
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.note_target_group_arn
    container_name   = "note"
    container_port   = 8081
  }

  service_registries {
    registry_arn = aws_service_discovery_service.services["note"].arn
  }

  depends_on = [aws_lb_listener_rule.note]
}

resource "aws_lb_listener_rule" "note" {
  listener_arn = var.https_listener_arn
  priority     = 200

  action {
    type             = "forward"
    target_group_arn = var.note_target_group_arn
  }

  condition {
    path_pattern {
      values = ["/note*", "/note/*"]
    }
  }
}

# ──────────────────────────────────────────────
# Document (NestJS)
# ──────────────────────────────────────────────
resource "aws_ecs_task_definition" "document" {
  family                   = "${var.project_name}-${var.environment}-document"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 512
  memory                   = 1024
  execution_role_arn       = var.ecs_task_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn

  container_definitions = jsonencode([{
    name      = "document"
    image     = "${var.ecr_repository_url}/document:${var.image_tags.document}"
    essential = true

    portMappings = [{
      containerPort = 8082
      protocol      = "tcp"
    }]

    environment = concat(local.common_environment, [
      { name = "OTEL_SERVICE_NAME", value = "notopia-${var.environment}-document" },
      { name = "NOTOPIA_DOCUMENT_API_URL", value = "https://${var.domain_name}" },
      { name = "NOTOPIA_DOCUMENT_DB_HOST", value = var.document_database.host },
      { name = "NOTOPIA_DOCUMENT_DB_PORT", value = tostring(var.document_database.port) },
      { name = "NOTOPIA_DOCUMENT_DB_USER", value = var.document_database.username },
      { name = "NOTOPIA_DOCUMENT_DB_PASSWORD", value = var.document_database.password },
      { name = "NOTOPIA_DOCUMENT_SERVICES_NOTE_GRPC_URL", value = "${aws_service_discovery_service.services["note"].name}.${aws_service_discovery_private_dns_namespace.main.name}:18081" },
      { name = "NOTOPIA_DOCUMENT_SERVICES_AUTHORIZATION_GRPC_URL", value = "${aws_service_discovery_service.services["authorization"].name}.${aws_service_discovery_private_dns_namespace.main.name}:18089" },
      { name = "NOTOPIA_DOCUMENT_S3_ENDPOINT", value = "https://s3.${var.aws_region}.amazonaws.com" },
      { name = "NOTOPIA_DOCUMENT_S3_REGION", value = var.aws_region },
      { name = "NOTOPIA_DOCUMENT_S3_BUCKET_NAME", value = var.s3_bucket_name },
      { name = "NOTOPIA_DOCUMENT_KAFKA_BROKERS", value = var.msk_bootstrap_brokers },
      { name = "NOTOPIA_DOCUMENT_KAFKA_CLIENT_ID", value = "document-${var.environment}" },
      { name = "NOTOPIA_DOCUMENT_KAFKA_GROUP_ID", value = "document-${var.environment}" },
      { name = "NOTOPIA_DOCUMENT_AUTHENTICATION_JWKS_URLS", value = "https://auth.${var.domain_name}/application/o/jwks/" },
    ])

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["document"].name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "document"
      }
    }

    healthCheck = {
      command     = ["CMD-SHELL", "wget -qO- http://localhost:8082/healthz || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 30
    }
  }])

  tags = {
    Name        = "${var.project_name}-${var.environment}-document"
    Environment = var.environment
  }
}

resource "aws_ecs_service" "document" {
  name            = "document"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.document.arn
  desired_count   = var.service_counts.document
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.document_target_group_arn
    container_name   = "document"
    container_port   = 8082
  }

  service_registries {
    registry_arn = aws_service_discovery_service.services["document"].arn
  }

  depends_on = [aws_lb_listener_rule.document]
}

resource "aws_lb_listener_rule" "document" {
  listener_arn = var.https_listener_arn
  priority     = 300

  action {
    type             = "forward"
    target_group_arn = var.document_target_group_arn
  }

  condition {
    path_pattern {
      values = ["/document*", "/document/*"]
    }
  }
}

# ──────────────────────────────────────────────
# Authorization (Go)
# ──────────────────────────────────────────────
resource "aws_ecs_task_definition" "authorization" {
  family                   = "${var.project_name}-${var.environment}-authorization"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = var.ecs_task_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn

  container_definitions = jsonencode([{
    name      = "authorization"
    image     = "${var.ecr_repository_url}/authorization:${var.image_tags.authorization}"
    essential = true

    portMappings = [{
      containerPort = 18089
      protocol      = "tcp"
    }]

    environment = concat(local.common_environment, [
      { name = "OTEL_SERVICE_NAME", value = "notopia-${var.environment}-authorization" },
      { name = "NOTOPIA_AUTHORIZATION_GENERAL_APP_ENV", value = var.environment },
      { name = "NOTOPIA_AUTHORIZATION_LOG_LEVEL", value = "info" },
      { name = "NOTOPIA_AUTHORIZATION_DATABASE_URL", value = "postgresql://${var.authorization_database.username}:${var.authorization_database.password}@${var.authorization_database.host}:${var.authorization_database.port}/${var.authorization_database.name}" },
      { name = "NOTOPIA_AUTHORIZATION_KAFKA_BROKERS", value = var.msk_bootstrap_brokers },
      { name = "NOTOPIA_AUTHORIZATION_KAFKA_CONSUMER_GROUP", value = "authorization-${var.environment}" },
    ])

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["authorization"].name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "authorization"
      }
    }

    healthCheck = {
      command     = ["CMD-SHELL", "wget -qO- http://localhost:28089/healthz || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 30
    }
  }])

  tags = {
    Name        = "${var.project_name}-${var.environment}-authorization"
    Environment = var.environment
  }
}

resource "aws_ecs_service" "authorization" {
  name            = "authorization"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.authorization.arn
  desired_count   = var.service_counts.authorization
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.authorization_target_group_arn
    container_name   = "authorization"
    container_port   = 18089
  }

  service_registries {
    registry_arn = aws_service_discovery_service.services["authorization"].arn
  }

  depends_on = [aws_lb_listener_rule.authorization]
}

resource "aws_lb_listener_rule" "authorization" {
  listener_arn = var.https_listener_arn
  priority     = 400

  action {
    type             = "forward"
    target_group_arn = var.authorization_target_group_arn
  }

  condition {
    path_pattern {
      values = ["/authorization*", "/authorization/*"]
    }
  }
}

# ──────────────────────────────────────────────
# Search Worker (NestJS) - no ALB, internal only
# ──────────────────────────────────────────────
resource "aws_ecs_task_definition" "search_worker" {
  family                   = "${var.project_name}-${var.environment}-search-worker"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = var.ecs_task_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn

  container_definitions = jsonencode([{
    name      = "search-worker"
    image     = "${var.ecr_repository_url}/search-worker:${var.image_tags.search_worker}"
    essential = true

    environment = concat(local.common_environment, [
      { name = "OTEL_SERVICE_NAME", value = "notopia-${var.environment}-search-worker" },
      { name = "NOTOPIA_SEARCH_WORKER_HEALTH_CHECK_PORT", value = "28083" },
      { name = "NOTOPIA_SEARCH_WORKER_KAFKA_BROKERS", value = var.msk_bootstrap_brokers },
      { name = "NOTOPIA_SEARCH_WORKER_KAFKA_CLIENT_ID", value = "search-worker-${var.environment}" },
      { name = "NOTOPIA_SEARCH_WORKER_KAFKA_GROUP_ID", value = "search-worker-${var.environment}" },
      { name = "NOTOPIA_SEARCH_WORKER_MEILI_HOST", value = "http://${aws_service_discovery_service.services["meilisearch"].name}.${aws_service_discovery_private_dns_namespace.main.name}:7700" },
    ])

    secrets = [
      { name = "NOTOPIA_SEARCH_WORKER_MEILI_API_KEY", valueFrom = var.meilisearch_api_key_ssm_arn },
    ]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["search-worker"].name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "search-worker"
      }
    }

    healthCheck = {
      command     = ["CMD-SHELL", "wget -qO- http://localhost:28083/healthz || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 30
    }
  }])

  tags = {
    Name        = "${var.project_name}-${var.environment}-search-worker"
    Environment = var.environment
  }
}

resource "aws_ecs_service" "search_worker" {
  name            = "search-worker"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.search_worker.arn
  desired_count   = var.service_counts.search_worker
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = false
  }

  service_registries {
    registry_arn = aws_service_discovery_service.services["search-worker"].arn
  }
}

# ──────────────────────────────────────────────
# Authentik (IdP)
# ──────────────────────────────────────────────
resource "aws_ecs_task_definition" "authentik" {
  family                   = "${var.project_name}-${var.environment}-authentik"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 512
  memory                   = 1024
  execution_role_arn       = var.ecs_task_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn

  container_definitions = jsonencode([
    {
      name      = "server"
      image     = "ghcr.io/goauthentik/server:2026.2.3"
      essential = true

      portMappings = [{
        containerPort = 9000
        protocol      = "tcp"
      }]

      environment = [
        { name = "AUTHENTIK_REDIS__HOST", value = "${var.redis_endpoint}" },
        { name = "AUTHENTIK_REDIS__PORT", value = tostring(var.redis_port) },
        { name = "AUTHENTIK_POSTGRESQL__HOST", value = var.authentik_database.host },
        { name = "AUTHENTIK_POSTGRESQL__PORT", value = tostring(var.authentik_database.port) },
        { name = "AUTHENTIK_POSTGRESQL__NAME", value = var.authentik_database.name },
        { name = "AUTHENTIK_POSTGRESQL__USER", value = var.authentik_database.username },
        { name = "AUTHENTIK_POSTGRESQL__PASSWORD", value = var.authentik_database.password },
        { name = "AUTHENTIK_SECRET_KEY", value = var.authentik_secret_key },
        { name = "AUTHENTIK_LISTEN__HTTPS", value = "0" },
        { name = "AUTHENTIK_LISTEN__PORT", value = "9000" },
        { name = "AUTHENTIK_BOOTSTRAP_EMAIL", value = var.admin_email },
        { name = "AUTHENTIK_BOOTSTRAP_PASSWORD", value = var.admin_password },
        { name = "AUTHENTIK_BOOTSTRAP_ALLOW_ME", value = "true" },
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.services["authentik"].name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "authentik-server"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "curl -sf http://localhost:9000/-/health/ready || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }
    },
    {
      name      = "worker"
      image     = "ghcr.io/goauthentik/server:2026.2.3"
      essential = false
      command   = ["ak", "worker"]

      environment = [
        { name = "AUTHENTIK_REDIS__HOST", value = "${var.redis_endpoint}" },
        { name = "AUTHENTIK_REDIS__PORT", value = tostring(var.redis_port) },
        { name = "AUTHENTIK_POSTGRESQL__HOST", value = var.authentik_database.host },
        { name = "AUTHENTIK_POSTGRESQL__PORT", value = tostring(var.authentik_database.port) },
        { name = "AUTHENTIK_POSTGRESQL__NAME", value = var.authentik_database.name },
        { name = "AUTHENTIK_POSTGRESQL__USER", value = var.authentik_database.username },
        { name = "AUTHENTIK_POSTGRESQL__PASSWORD", value = var.authentik_database.password },
        { name = "AUTHENTIK_SECRET_KEY", value = var.authentik_secret_key },
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.services["authentik"].name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "authentik-worker"
        }
      }
    }
  ])

  tags = {
    Name        = "${var.project_name}-${var.environment}-authentik"
    Environment = var.environment
  }
}

resource "aws_ecs_service" "authentik" {
  name            = "authentik"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.authentik.arn
  desired_count   = var.service_counts.authentik
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.authentik_target_group_arn
    container_name   = "server"
    container_port   = 9000
  }

  service_registries {
    registry_arn = aws_service_discovery_service.services["authentik"].arn
  }

  depends_on = [aws_lb_listener_rule.authentik]
}

resource "aws_lb_listener_rule" "authentik" {
  listener_arn = var.https_listener_arn
  priority     = 500

  action {
    type             = "forward"
    target_group_arn = var.authentik_target_group_arn
  }

  condition {
    path_pattern {
      values = ["/-/authentik*", "/-/authentik/*"]
    }
  }
}

# ──────────────────────────────────────────────
# Meilisearch
# ──────────────────────────────────────────────
resource "aws_ecs_task_definition" "meilisearch" {
  family                   = "${var.project_name}-${var.environment}-meilisearch"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 512
  memory                   = 1024
  execution_role_arn       = var.ecs_task_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn

  volume {
    name = "meilisearch-data"
  }

  container_definitions = jsonencode([{
    name      = "meilisearch"
    image     = "getmeili/meilisearch:v1.41"
    essential = true

    portMappings = [{
      containerPort = 7700
      protocol      = "tcp"
    }]

    environment = [
      { name = "MEILI_NO_ANALYTICS", value = "true" },
      { name = "MEILI_ENV", value = "production" },
    ]

    secrets = [
      { name = "MEILI_MASTER_KEY", valueFrom = var.meilisearch_api_key_ssm_arn },
    ]

    mountPoints = [{
      sourceVolume  = "meilisearch-data"
      containerPath = "/meili_data"
      readOnly      = false
    }]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["meilisearch"].name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "meilisearch"
      }
    }

    healthCheck = {
      command     = ["CMD-SHELL", "curl -sf http://localhost:7700/health || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 30
    }
  }])

  tags = {
    Name        = "${var.project_name}-${var.environment}-meilisearch"
    Environment = var.environment
  }
}

resource "aws_ecs_service" "meilisearch" {
  name            = "meilisearch"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.meilisearch.arn
  desired_count   = var.service_counts.meilisearch
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.meilisearch_target_group_arn
    container_name   = "meilisearch"
    container_port   = 7700
  }

  service_registries {
    registry_arn = aws_service_discovery_service.services["meilisearch"].arn
  }

  depends_on = [aws_lb_listener_rule.meilisearch]
}

resource "aws_lb_listener_rule" "meilisearch" {
  listener_arn = var.https_listener_arn
  priority     = 600

  action {
    type             = "forward"
    target_group_arn = var.meilisearch_target_group_arn
  }

  condition {
    path_pattern {
      values = ["/search*", "/search/*", "/multi-search*", "/multi-search/*", "/indexes*", "/indexes/*"]
    }
  }
}

# ──────────────────────────────────────────────
# API Web (OpenAPI docs SPA)
# ──────────────────────────────────────────────
resource "aws_ecs_task_definition" "api_web" {
  family                   = "${var.project_name}-${var.environment}-api-web"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = var.ecs_task_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn

  container_definitions = jsonencode([{
    name      = "api-web"
    image     = "${var.ecr_repository_url}/api-web:${var.image_tags.api_web}"
    essential = true

    portMappings = [{
      containerPort = 9080
      protocol      = "tcp"
    }]

    environment = local.common_environment

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.services["api-web"].name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "api-web"
      }
    }

    healthCheck = {
      command     = ["CMD-SHELL", "wget -qO- http://localhost:9080/ || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 30
    }
  }])

  tags = {
    Name        = "${var.project_name}-${var.environment}-api-web"
    Environment = var.environment
  }
}

resource "aws_ecs_service" "api_web" {
  name            = "api-web"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api_web.arn
  desired_count   = var.service_counts.api_web
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [var.ecs_security_group_id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.api_web_target_group_arn
    container_name   = "api-web"
    container_port   = 9080
  }

  service_registries {
    registry_arn = aws_service_discovery_service.services["api-web"].arn
  }

  depends_on = [aws_lb_listener_rule.api_web]
}

resource "aws_lb_listener_rule" "api_web" {
  listener_arn = var.https_listener_arn
  priority     = 700

  action {
    type             = "forward"
    target_group_arn = var.api_web_target_group_arn
  }

  condition {
    path_pattern {
      values = ["/docs", "/docs/*"]
    }
  }
}
