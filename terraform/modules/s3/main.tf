module "s3" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "~> 5.0"

  bucket = "${var.project_name}-${var.environment}-document"

  control_object_ownership = true
  object_ownership         = "BucketOwnerEnforced"

  versioning = {
    enabled = true
  }

  server_side_encryption_configuration = {
    rule = {
      apply_server_side_encryption_by_default = {
        sse_algorithm = "aws:kms"
      }
    }
  }

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true

  lifecycle_rule = [
    {
      id     = "abort-incomplete-multipart"
      status = "Enabled"

      abort_incomplete_multipart_upload = {
        days_after_initiation = 7
      }
    }
  ]

  tags = {
    Name        = "${var.project_name}-${var.environment}-document"
    Environment = var.environment
  }
}
