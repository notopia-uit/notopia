output "prometheus_workspace_endpoint" {
  value = aws_prometheus_workspace.main.prometheus_endpoint
}

output "prometheus_workspace_arn" {
  value = aws_prometheus_workspace.main.arn
}

output "grafana_workspace_endpoint" {
  value = aws_grafana_workspace.main.endpoint
}

output "grafana_workspace_id" {
  value = aws_grafana_workspace.main.id
}

output "xray_role_arn" {
  value = module.xray_role.arn
}

output "prometheus_remote_write_role_arn" {
  value = module.prometheus_remote_write_role.arn
}
