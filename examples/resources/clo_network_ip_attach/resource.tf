resource "clo_network_ip_attach" "fip_attach"{
  address_id = clo_resource_ip.fip_1.id
  entity_name = "server"
  entity_id = clo_resource_instance.serv.id
}
