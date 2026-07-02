# Look up an available datastore (database engine + version) for the project.
data "clo_dbaas_datastores" "all" {
  project_id = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
}

# A minimal managed-database cluster. An address is allocated automatically.
resource "clo_dbaas_cluster" "cluster_1" {
  project_id   = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
  name         = "app-db"
  datastore_id = data.clo_dbaas_datastores.all.result[0].id
  storage_size = 20

  flavor {
    vcpus = 2
    ram   = 4
  }
}

# A cluster bound to an existing address, created stopped.
resource "clo_dbaas_cluster" "cluster_2" {
  project_id   = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
  name         = "reporting-db"
  datastore_id = data.clo_dbaas_datastores.all.result[0].id
  storage_size = 50
  enabled      = false

  flavor {
    vcpus = 4
    ram   = 8
  }

  # Bind an existing address; omit the block to allocate one automatically.
  address {
    id = "b0298e0c-057c-4298-8505-5e57f5a60f8c"
  }
}