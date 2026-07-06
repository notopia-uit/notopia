data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_acm_certificate" "main" {
  count    = var.certificate_arn != "" ? 0 : 1
  domain   = "*.${var.domain_name}"
  statuses = ["ISSUED"]
}
