output "bootstrap_brokers_tls" {
  value = module.msk.bootstrap_brokers_tls
}

output "zookeeper_connect_string" {
  value = module.msk.zookeeper_connect_string
}

output "cluster_arn" {
  value = module.msk.arn
}
