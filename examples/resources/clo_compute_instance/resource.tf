resource "clo_compute_instance" "myserv" {
  project_id = "e9ff0f7-0b8c-4ec5-a0a4-e30ce0db287"
  name = "my_server"
  flavor_ram = 4
  flavor_vcpus = 2
  image_id = "2d6270-c4b6-4d2c-b238-8fa58f35634d"
  block_device {
    bootable = true
    storage_type = "volume"
    size = 40
  }
  addresses {
    version = 4
    with_floating = true
    external = true
    ddos_protection = false
  }
}