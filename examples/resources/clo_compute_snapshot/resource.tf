# A point-in-time snapshot of a server. Snapshots are stored as private images
# and auto-expire at their deleted_in timestamp.
resource "clo_compute_snapshot" "nightly" {
  server_id = clo_compute_instance.test-server.id
  name      = "nightly"
}