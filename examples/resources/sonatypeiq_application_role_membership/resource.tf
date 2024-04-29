data "sonatypeiq_application" "sandbox" {
  public_id = "sandbox-application"
}

data "sonatypeiq_role" "developer" {
  name = "Developer"
}

resource "sonatypeiq_user" "example_user" {
  username   = "example2"
  password   = "randomthing"
  first_name = "Example"
  last_name  = "User"
  email      = "example@user.tld"
}

# Create and manage application role memberships for Sonatype IQ Server
resource "sonatypeiq_application_role_membership" "application_role_membership" {
  application_id = data.sonatypeiq_application.sandbox.id
  role_id        = data.sonatypeiq_role.developer.id
  user_name      = sonatypeiq_user.example_user.username

  # user_name and group_name are mutually exclusive.
  # group_name   = "developers"
}
