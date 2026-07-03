# Manage an instance's power state opt-in. Without this resource the instance is
# left in whatever state it boots into. Destroying it stops managing power — the
# instance keeps its current state.
resource "clo_compute_instance_power" "myserv" {
  instance_id = clo_compute_instance.myserv.id
  enabled     = false # power the instance off
}