# Define required providers
terraform {
  required_providers {
    clo = {
      version = "2.5.0"
      source  = "clo-ru/clo"
    }
  }
}

# Configure the provider
provider "clo" {
  auth_url = "https://api.clo.ru"
  token    = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9."
}

# Create an instance
resource "clo_compute_instance" "test-server" {
  # ...
}