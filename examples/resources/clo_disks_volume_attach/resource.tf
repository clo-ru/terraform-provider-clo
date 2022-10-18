resource "clo_disks_volume_attach" "v_att" {
  volume_id = clo_disks_volume.volume.id
  instance_id = clo_compute_instance.serv.id
}
