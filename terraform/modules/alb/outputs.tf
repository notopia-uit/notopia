output "alb_dns_name" {
  value = module.alb.dns_name
}

output "alb_zone_id" {
  value = module.alb.zone_id
}

output "web_target_group_arn" {
  value = module.alb.target_groups["web"].arn
}

output "note_target_group_arn" {
  value = module.alb.target_groups["note"].arn
}

output "document_target_group_arn" {
  value = module.alb.target_groups["document"].arn
}

output "authorization_target_group_arn" {
  value = module.alb.target_groups["authorization"].arn
}

output "authentik_target_group_arn" {
  value = module.alb.target_groups["authentik"].arn
}

output "meilisearch_target_group_arn" {
  value = module.alb.target_groups["meilisearch"].arn
}

output "api_web_target_group_arn" {
  value = module.alb.target_groups["api_web"].arn
}

output "https_listener_arn" {
  value = module.alb.listeners["https"].arn
}

output "alb_arn" {
  value = module.alb.arn
}

output "alb_arn_suffix" {
  value = module.alb.arn_suffix
}
