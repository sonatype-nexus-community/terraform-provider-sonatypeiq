---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sonatypeiq_organization Resource - sonatypeiq"
subcategory: ""
description: |-
  Use this resource to manage Organizations
---

# sonatypeiq_organization (Resource)

Use this resource to manage Organizations

## Example Usage

```terraform
# Create and manage an Organization "Sub Organization" under the "Sandbox Organization"
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

resource "sonatypeiq_organization" "sub_org" {
  name                   = "Sub Organization"
  parent_organization_id = data.sonatypeiq_organization.sandbox.id
}

output "sub_org_id" {
  value = sonatypeiq_organization.sub_org.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `name` (String) Name of the Organization
- `parent_organization_id` (String) Internal ID of the Parent Organization if this Organization has a Parent Organization

### Read-Only

- `id` (String) Internal ID of the Organization
- `last_updated` (String)
