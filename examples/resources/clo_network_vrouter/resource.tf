# A minimal virtual router.
resource "clo_network_vrouter" "router_1" {
  project_id = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
  name       = "my-router"
}

# Attach private networks and start it stopped.
resource "clo_network_vrouter" "router_2" {
  project_id       = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
  name             = "internal-router"
  private_networks = ["3f2504e0-4f89-41d3-9a0c-0305e82c3301"]
  enabled          = false
}