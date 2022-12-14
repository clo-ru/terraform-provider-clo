---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "clo_disks_volume Resource - terraform-provider-clo"
subcategory: ""
description: |-
  Create a new volume in the project
---

# clo_disks_volume (Resource)

Create a new volume in the project

## Example Usage

```terraform
resource "clo_disks_volume" "volume_1" {
  project_id = "e9ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
  name = "my_volume_1"
  size = 30
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project_id` (String) ID of the project where the volume should be created
- `size` (Number) Size of the new volume in Gb

### Optional

- `name` (String) Human-readable name of the new volume
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `created_in` (String) Timestamp the volume was created
- `id` (String) ID of the new volume
- `status` (String)

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `read` (String)
- `update` (String)


