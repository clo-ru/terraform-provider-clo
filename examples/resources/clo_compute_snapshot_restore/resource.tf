# Provision a new server from a snapshot. Destroying this resource deletes the
# server it created (along with its volumes and addresses).
resource "clo_compute_snapshot_restore" "restored" {
  snapshot_id = clo_compute_snapshot.nightly.id
  name        = "restored-from-nightly"
}