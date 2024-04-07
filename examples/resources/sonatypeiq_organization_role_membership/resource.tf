data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

data "sonatypeiq_role" "developer" {
  name = "Developer"
}

resource "sonatypeiq_user" "example" {
  username   = "example"
  password   = "randomthing"
  first_name = "Example"
  last_name  = "User"
  email      = "example@user.tld"
}

# Create and manage application role memberships for Sonatype IQ Server
resource "sonatypeiq_organization_role_membership" "organization_role_membership" {
  organization_id = data.sonatypeiq_organization.sandbox.id
  role_id         = data.sonatypeiq_role.developer.id
  username        = sonatypeiq_user.example.username

  # group_name can also be used but it is mutually exclusive with the user_name attribute.
  # group_name = "developers"
}

