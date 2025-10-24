# Manage SAML Configuration
resource "sonatypeiq_config_saml" "saml_config" {
  email_attribute        = "email"
  entity_id              = "https://DOMAIN.TLD/api/v2/config/saml/metadata"
  first_name_attribute   = "firstName"
  groups_attribute       = "groups"
  identity_provider_name = "IDP Name"
  idp_metadata           = file("/path/to/saml-metadata.xml")
  last_name_attribute    = "lastName"
  username_attribute     = "username"
}