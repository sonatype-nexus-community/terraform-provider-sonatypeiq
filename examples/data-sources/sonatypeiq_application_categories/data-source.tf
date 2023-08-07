# Get Application Categories for the Root Organization
data "sonatypeiq_application_categories" "categories" {
  organization_id = "ROOT_ORGANIZATION_ID"
}