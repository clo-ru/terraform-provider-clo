resource "clo_network_ip_attach" "fip_attach"{
  address_id = clo_network_ip.fip_1.id
  entity_name = "server"
  entity_id = clo_compute_instance.serv.id
}
