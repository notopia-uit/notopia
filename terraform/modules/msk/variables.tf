variable "project_name" {
  type = string
}

variable "environment" {
  type = string
}

variable "private_subnet_ids" {
  type = list(string)
}

variable "security_group_id" {
  type = string
}

variable "broker_instance_type" {
  type    = string
  default = "kafka.m5.large"
}

variable "broker_count" {
  type    = number
  default = 3
}

variable "ebs_volume_size" {
  type    = number
  default = 100
}

variable "kms_key_arn" {
  type    = string
  default = ""
}
