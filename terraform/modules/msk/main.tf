module "msk" {
  source  = "terraform-aws-modules/msk-kafka-cluster/aws"
  version = "~> 3.0"

  create = true
  name   = "${var.project_name}-${var.environment}"

  kafka_version          = "3.7.1"
  number_of_broker_nodes = var.broker_count

  broker_node_instance_type   = var.broker_instance_type
  broker_node_client_subnets  = var.private_subnet_ids
  broker_node_security_groups = [var.security_group_id]

  broker_node_storage_info = {
    ebs_storage_info = {
      volume_size = var.ebs_volume_size
    }
  }

  encryption_at_rest_kms_key_arn      = var.kms_key_arn != "" ? var.kms_key_arn : null
  encryption_in_transit_client_broker = "TLS"
  encryption_in_transit_in_cluster    = true

  create_configuration = true
  configuration_name   = "${var.project_name}-${var.environment}"
  configuration_server_properties = {
    "auto.create.topics.enable" = "true"
    "delete.topic.enable"       = "true"
    "compression.type"          = "producer"
    "log.retention.hours"       = "168"
  }

  cloudwatch_logs_enabled                = true
  create_cloudwatch_log_group            = true
  cloudwatch_log_group_name              = "/msk/${var.project_name}-${var.environment}"
  cloudwatch_log_group_retention_in_days = 30

  tags = {
    Name        = "${var.project_name}-${var.environment}"
    Environment = var.environment
  }
}
