module "alb" {
  source = "terraform-aws-modules/alb/aws"

  name               = "${var.project_name}-${var.environment}"
  internal           = false
  load_balancer_type = "application"
  vpc_id             = var.vpc_id
  subnets            = var.public_subnet_ids

  enable_deletion_protection = var.environment == "prod"

  security_group_ingress_rules = {
    all_http = {
      from_port   = 80
      to_port     = 80
      ip_protocol = "tcp"
      cidr_ipv4   = "0.0.0.0/0"
    }
    all_https = {
      from_port   = 443
      to_port     = 443
      ip_protocol = "tcp"
      cidr_ipv4   = "0.0.0.0/0"
    }
  }

  security_group_egress_rules = {
    all = {
      ip_protocol = "-1"
      cidr_ipv4   = "0.0.0.0/0"
    }
  }

  listeners = {
    http = {
      port     = 80
      protocol = "HTTP"

      default_action = {
        type = "redirect"
        redirect = {
          port        = "443"
          protocol    = "HTTPS"
          status_code = "HTTP_301"
        }
      }
    }

    https = {
      port            = 443
      protocol        = "HTTPS"
      ssl_policy      = "ELBSecurityPolicy-TLS13-1-2-2021-06"
      certificate_arn = var.certificate_arn

      default_action = {
        type             = "forward"
        target_group_key = "web"
      }

      rules = {
        note = {
          priority = 200
          conditions = [{
            path_pattern = {
              values = ["/note*"]
            }
          }]
          actions = [{
            type             = "forward"
            target_group_key = "note"
          }]
        }

        document = {
          priority = 300
          conditions = [{
            path_pattern = {
              values = ["/document*"]
            }
          }]
          actions = [{
            type             = "forward"
            target_group_key = "document"
          }]
        }

        authorization = {
          priority = 400
          conditions = [{
            path_pattern = {
              values = ["/authorization*"]
            }
          }]
          actions = [{
            type             = "forward"
            target_group_key = "authorization"
          }]
        }

        authentik = {
          priority = 500
          conditions = [{
            path_pattern = {
              values = ["/-/authentik*"]
            }
          }]
          actions = [{
            type             = "forward"
            target_group_key = "authentik"
          }]
        }

        meilisearch = {
          priority = 600
          conditions = [{
            path_pattern = {
              values = ["/search*", "/indexes*"]
            }
          }]
          actions = [{
            type             = "forward"
            target_group_key = "meilisearch"
          }]
        }

        api_web = {
          priority = 700
          conditions = [{
            path_pattern = {
              values = ["/docs"]
            }
          }]
          actions = [{
            type             = "forward"
            target_group_key = "api_web"
          }]
        }
      }
    }
  }

  target_groups = {
    web = {
      name        = "${var.project_name}-${var.environment}-web"
      port        = 3000
      protocol    = "HTTP"
      target_type = "ip"

      health_check = {
        path                = "/"
        healthy_threshold   = 2
        unhealthy_threshold = 5
        timeout             = 5
        interval            = 30
        matcher             = "200-499"
      }

      tags = {
        Name        = "${var.project_name}-${var.environment}-web"
        Environment = var.environment
      }
    }

    note = {
      name        = "${var.project_name}-${var.environment}-note"
      port        = 8081
      protocol    = "HTTP"
      target_type = "ip"

      health_check = {
        path                = "/healthz"
        healthy_threshold   = 2
        unhealthy_threshold = 5
        timeout             = 5
        interval            = 30
        matcher             = "200"
      }

      tags = {
        Name        = "${var.project_name}-${var.environment}-note"
        Environment = var.environment
      }
    }

    document = {
      name        = "${var.project_name}-${var.environment}-doc"
      port        = 8082
      protocol    = "HTTP"
      target_type = "ip"

      health_check = {
        path                = "/healthz"
        healthy_threshold   = 2
        unhealthy_threshold = 5
        timeout             = 5
        interval            = 30
        matcher             = "200"
      }

      tags = {
        Name        = "${var.project_name}-${var.environment}-document"
        Environment = var.environment
      }
    }

    authorization = {
      name        = "${var.project_name}-${var.environment}-authz"
      port        = 18089
      protocol    = "HTTP"
      target_type = "ip"

      health_check = {
        path                = "/healthz"
        healthy_threshold   = 2
        unhealthy_threshold = 5
        timeout             = 5
        interval            = 30
        matcher             = "200"
      }

      tags = {
        Name        = "${var.project_name}-${var.environment}-authorization"
        Environment = var.environment
      }
    }

    authentik = {
      name        = "${var.project_name}-${var.environment}-auth"
      port        = 9000
      protocol    = "HTTP"
      target_type = "ip"

      health_check = {
        path                = "/-/health/ready"
        healthy_threshold   = 2
        unhealthy_threshold = 5
        timeout             = 5
        interval            = 30
        matcher             = "200"
      }

      tags = {
        Name        = "${var.project_name}-${var.environment}-authentik"
        Environment = var.environment
      }
    }

    meilisearch = {
      name        = "${var.project_name}-${var.environment}-meili"
      port        = 7700
      protocol    = "HTTP"
      target_type = "ip"

      health_check = {
        path                = "/health"
        healthy_threshold   = 2
        unhealthy_threshold = 5
        timeout             = 5
        interval            = 30
        matcher             = "200"
      }

      tags = {
        Name        = "${var.project_name}-${var.environment}-meilisearch"
        Environment = var.environment
      }
    }

    api_web = {
      name        = "${var.project_name}-${var.environment}-api-web"
      port        = 9080
      protocol    = "HTTP"
      target_type = "ip"

      health_check = {
        path                = "/"
        healthy_threshold   = 2
        unhealthy_threshold = 5
        timeout             = 5
        interval            = 30
        matcher             = "200-499"
      }

      tags = {
        Name        = "${var.project_name}-${var.environment}-api-web"
        Environment = var.environment
      }
    }
  }

  tags = {
    Name        = "${var.project_name}-${var.environment}"
    Environment = var.environment
  }
}
