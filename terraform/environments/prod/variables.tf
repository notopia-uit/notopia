variable "aws_region" {
  type    = string
  default = "ap-southeast-1"
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "availability_zones" {
  type    = list(string)
  default = ["ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c"]
}

variable "domain_name" {
  type = string
}

variable "certificate_arn" {
  type    = string
  default = ""
}

variable "ecr_repository_url" {
  type    = string
  default = "ghcr.io/notopia-uit"
}

variable "image_tags" {
  type = object({
    web           = string
    document      = string
    search_worker = string
    note          = string
    authorization = string
    api_web       = string
  })
}

variable "service_counts" {
  type = object({
    web           = number
    document      = number
    search_worker = number
    note          = number
    authorization = number
    authentik     = number
    meilisearch   = number
    api_web       = number
  })
}

variable "db_username" {
  type    = string
  default = "notopia"
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "authentik_client_id" {
  type      = string
  sensitive = true
}

variable "authentik_secret" {
  type      = string
  sensitive = true
}

variable "authentik_token" {
  type      = string
  sensitive = true
}

variable "authentik_secret_key" {
  type      = string
  sensitive = true
}

variable "admin_email" {
  type = string
}

variable "admin_password" {
  type      = string
  sensitive = true
}

variable "meilisearch_api_key" {
  type      = string
  sensitive = true
}
