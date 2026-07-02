# Fetch a presigned URL to download a backup. The URL is ephemeral and is
# refreshed on every read.
data "clo_dbaas_backup_download" "app" {
  backup_id = clo_dbaas_backup.cluster_backup.id
}