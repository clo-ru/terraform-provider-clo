data "clo_dbaas_cluster_config" "cluster_config" {
  cluster_id = "3f2504e0-4f89-41d3-9a0c-0305e82c3301"
}

# Read a single tuned parameter from the live configuration.
output "max_connections" {
  value = data.clo_dbaas_cluster_config.cluster_config.current["max_connections"]
}