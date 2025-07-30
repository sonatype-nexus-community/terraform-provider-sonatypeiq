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
