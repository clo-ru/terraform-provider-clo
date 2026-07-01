# A TCP load balancer bound to an existing address.
resource "clo_network_loadbalancer" "lb_1" {
  project_id          = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
  name                = "web-lb"
  algorithm           = "ROUND_ROBIN"
  session_persistence = false

  # Bind an existing address; omit the block to allocate one automatically.
  address {
    id = "b0298e0c-057c-4298-8505-5e57f5a60f8c"
  }

  # Health check. delay (interval) >= 80s, timeout >= 15s and <= delay, max_retries 1-10.
  healthmonitor {
    type        = "TCP"
    delay       = 80
    timeout     = 15
    max_retries = 3
  }
}

# An HTTP load balancer that allocates a new DDoS-protected address.
resource "clo_network_loadbalancer" "lb_2" {
  project_id = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
  name       = "api-lb"
  algorithm  = "LEAST_CONNECTIONS"

  address {
    ddos_protection = true
  }

  healthmonitor {
    type           = "HTTP"
    delay          = 80
    timeout        = 15
    max_retries    = 3
    http_method    = "GET"
    url_path       = "/health"
    expected_codes = "200"
  }
}