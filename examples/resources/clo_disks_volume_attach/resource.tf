resource "clo_disks_volume_attach" "v_att" {
  volume_id = clo_resource_volume.volume.id
  instance_id = clo_resource_instance.serv.id
}
