# Get your Sonatype IQ Server SAML Metadata
data "sonatypeiq_config_saml" "saml" {}

output "saml_metadata" {
  value = data.sonatypeiq_config_saml.saml.saml_metadata
}