---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "clo_network_ips Data Source - terraform-provider-clo"
subcategory: ""
description: |-
  Fetches the list of the IP-addresses
---

# clo_network_ips (Data Source)

Fetches the list of the IP-addresses

## Example Usage

```terraform
data "clo_network_ips" "all_addresses" {
  project_id = "1e8ff0f7-0b8c-4ec5-a0a4-e30cea0db287"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project_id` (String) ID of the project that owns addresses

### Read-Only

- `id` (String) The ID of this resource.
- `result` (List of Object) The object that holds the results (see [below for nested schema](#nestedatt--result))

<a id="nestedatt--result"></a>
### Nested Schema for `result`

Read-Only:

- `address` (String)
- `attached_to` (List of Object) (see [below for nested schema](#nestedobjatt--results--attached_to))
- `created_in` (String)
- `ddos_protection` (Boolean)
- `id` (String)
- `is_primary` (Boolean)
- `ptr` (String)
- `status` (String)
- `type` (String)

<a id="nestedobjatt--results--attached_to"></a>
### Nested Schema for `result.attached_to`

Read-Only:

- `entity` (String)
- `id` (String)


