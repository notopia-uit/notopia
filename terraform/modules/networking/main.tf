module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 6.0"

  name = "${var.project_name}-${var.environment}"
  cidr = var.vpc_cidr
  azs  = var.availability_zones

  private_subnets = var.private_subnet_cidrs
  public_subnets  = var.public_subnet_cidrs

  enable_nat_gateway   = true
  single_nat_gateway   = var.single_nat_gateway
  enable_dns_hostnames = true
  enable_dns_support   = true

  manage_default_security_group = false

  tags = {
    Name        = "${var.project_name}-${var.environment}"
    Environment = var.environment
  }

  public_subnet_tags = {
    Name        = "${var.project_name}-${var.environment}-public"
    Environment = var.environment
    Tier        = "public"
  }

  private_subnet_tags = {
    Name        = "${var.project_name}-${var.environment}-private"
    Environment = var.environment
    Tier        = "private"
  }
}
