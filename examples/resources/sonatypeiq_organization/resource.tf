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