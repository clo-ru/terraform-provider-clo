resource "clo_storage_s3_user" "s3_user" {
  project_id = "e9ff0f-0b8c-4ec5-a0a4-e30cea0db287"
  canonical_name = "server"
  max_buckets = 2
  user_quota_max_size = 30
}