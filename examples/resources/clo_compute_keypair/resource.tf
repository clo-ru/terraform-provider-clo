# Import an existing public key.
resource "clo_compute_keypair" "imported" {
  project_id = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
  name       = "my-laptop"
  public_key = "ssh-rsa AAAAB3Nza...== user@host"
}

# Or let the API generate one. The private key is returned only once and stored
# in state, so mark it sensitive when you consume it.
resource "clo_compute_keypair" "generated" {
  project_id = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
  name       = "generated-key"
}