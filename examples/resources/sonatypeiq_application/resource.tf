# Create and manage an "Example Application" under the "Sandbox Organization"
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

resource "sonatypeiq_application" "example" {
  name            = "Example Application"
  public_id       = "example_application"
  organization_id = data.sonatypeiq_organization.sandbox.id
}

output "example_app" {
  value = sonatypeiq_application.example
}