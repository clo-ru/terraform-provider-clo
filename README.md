Terraform CLO Provider
============================

The CLO provider lets you manage [clo.ru](https://clo.ru) cloud resources —
compute instances, disks, networking (including load balancers), managed
databases, and S3 storage — with Terraform.

Documentation: [registry.terraform.io](https://registry.terraform.io/providers/clo-ru/clo/latest/docs)

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) 1.25 (to build the provider plugin)

Using the Provider
-------------------

```hcl
terraform {
  required_providers {
    clo = {
      source = "clo-ru/clo"
    }
  }
}

provider "clo" {
  auth_url = "https://api.clo.ru"
  # token may also be supplied via the CLO_API_AUTH_TOKEN environment variable
  token = "<your-api-token>"
}
```

Supported Resources
-------------------

- **Compute**: `clo_compute_instance`, `clo_compute_instance_power`, `clo_compute_keypair`, `clo_compute_snapshot`, `clo_compute_snapshot_restore`
- **Disks**: `clo_disks_volume`, `clo_disks_volume_attach`
- **Network**: `clo_network_ip`, `clo_network_ip_attach`, `clo_network_vrouter`, `clo_network_loadbalancer`, `clo_network_loadbalancer_rule`
- **Database**: `clo_dbaas_cluster`, `clo_dbaas_database`, `clo_dbaas_backup`
- **Storage**: `clo_storage_s3_user`, `clo_storage_s3_user_keys`

Data Sources
------------

- **Project**: `clo_projects`, `clo_project_image`, `clo_project_images`, `clo_project_recipe`, `clo_project_recipes`
- **Compute**: `clo_compute_instance`, `clo_compute_instances`, `clo_compute_keypair`, `clo_compute_keypairs`, `clo_compute_snapshots`
- **Disks**: `clo_disks_volume`, `clo_disks_volumes`
- **Network**: `clo_network_ip`, `clo_network_ips`, `clo_network_vrouters`, `clo_network_loadbalancers`, `clo_network_loadbalancer_rules`
- **Database**: `clo_dbaas_clusters`, `clo_dbaas_cluster_config`, `clo_dbaas_databases`, `clo_dbaas_nodes`, `clo_dbaas_datastores`, `clo_dbaas_backups`, `clo_dbaas_backup_download`
- **Storage**: `clo_storage_s3_user`, `clo_storage_s3_users`, `clo_storage_s3_user_keys`

Building The Provider
---------------------

Clone the repository:

```sh
git clone git@github.com:clo-ru/terraform-provider-clo.git
```

Enter the provider directory and build:

```sh
cd terraform-provider-clo
make build
```

Run unit tests with `make test`. Acceptance tests exercise the live API, are
gated behind `TF_ACC`, and require `CLO_API_AUTH_URL`, `CLO_API_AUTH_TOKEN`, and
`CLO_API_PROJECT_ID`:

```sh
make testacc
```