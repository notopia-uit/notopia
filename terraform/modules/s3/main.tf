resource "aws_s3_bucket" "document" {
  bucket = "${var.project_name}-${var.environment}-document"

  tags = {
    Name        = "${var.project_name}-${var.environment}-document"
    Environment = var.environment
  }
}

resource "aws_s3_bucket_versioning" "document" {
  bucket = aws_s3_bucket.document.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "document" {
  bucket = aws_s3_bucket.document.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "document" {
  bucket = aws_s3_bucket.document.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "document" {
  bucket = aws_s3_bucket.document.id

  rule {
    id     = "abort-incomplete-multipart"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}
