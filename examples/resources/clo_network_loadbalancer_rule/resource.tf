# A listener rule: forward external port 80 to port 8080 on the backend.
# address_id is a backend server's internal (FIXED) address -- the target the
# load balancer forwards traffic to, not the balancer's external address.
resource "clo_network_loadbalancer_rule" "http" {
  loadbalancer_id        = clo_network_loadbalancer.lb_1.id
  address_id             = "c47676ad-9124-4a56-8982-196bfa997187"
  external_protocol_port = 80
  internal_protocol_port = 8080
}