# A database inside an existing managed-database cluster. Each database has its
# own admin user; change admin_password to rotate it on the running database.
resource "clo_dbaas_database" "app" {
  cluster_id     = clo_dbaas_cluster.cluster_1.id
  name           = "app"
  admin_username = "app_admin"
  admin_password = "change-me-please"
}