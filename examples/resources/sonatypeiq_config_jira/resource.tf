# Create and manage Jira integration configuration for Sonatype IQ Server
resource "sonatypeiq_config_jira" "jira_config" {
  url      = "https://jira.my-domain.tld"
  username = "iq-integration"
  password = var.jira_password

  # Optional: custom fields for Jira integration
  # custom_fields = jsonencode({
  #   field1 = {
  #     key1 = "value1"
  #   }
  # })
}
