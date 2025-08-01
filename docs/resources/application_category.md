---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sonatypeiq_application_category Resource - sonatypeiq"
subcategory: ""
description: |-
  Use this resource to manage Application Categories/Tags which can then be applied to Applications.
---

# sonatypeiq_application_category (Resource)

Use this resource to manage Application Categories/Tags which can then be applied to Applications.

## Example Usage

```terraform
data "sonatypeiq_organization" "root" {
  id = "ROOT_ORGANIZATION_ID"
}

data "sonatypeiq_role" "developer" {
  name = "Developer"
}

resource "sonatypeiq_application_category" "category1" {
  name            = "Name here"
  description     = "Description here"
  organization_id = data.sonatypeiq_organization.root
  color           = "dark-blue"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `color` (String) Color of the Application Category
- `description` (String) Description of the Application Category
- `name` (String) Name of the Application Category
- `organization_id` (String) Internal ID of the Organization to which this Application Category belongs. Use `ROOT_ORGANIZATION_ID` for the Root Organization.

### Read-Only

- `id` (String) Internal ID of the Application Category
- `last_updated` (String)
