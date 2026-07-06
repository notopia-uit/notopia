output "alb_dns_name" {
  value = aws_lb.main.dns_name
}

output "alb_zone_id" {
  value = aws_lb.main.zone_id
}

output "web_target_group_arn" {
  value = aws_lb_target_group.web.arn
}

output "note_target_group_arn" {
  value = aws_lb_target_group.note.arn
}

output "document_target_group_arn" {
  value = aws_lb_target_group.document.arn
}

output "authz_target_group_arn" {
  value = aws_lb_target_group.authz.arn
}

output "authentik_target_group_arn" {
  value = aws_lb_target_group.authentik.arn
}

output "meilisearch_target_group_arn" {
  value = aws_lb_target_group.meilisearch.arn
}

output "api_web_target_group_arn" {
  value = aws_lb_target_group.api_web.arn
}

output "https_listener_arn" {
  value = aws_lb_listener.https.arn
}

output "alb_arn" {
  value = aws_lb.main.arn
}

output "alb_arn_suffix" {
  value = aws_lb.main.arn_suffix
}
