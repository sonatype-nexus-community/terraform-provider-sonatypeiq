# Create and manage the source control for source control of organizations and applications

# Set source control for an application
data "sonatypeiq_application" "application" {
  public_id = "sandbox-application"
}

resource "sonatypeiq_source_control" "application" {
  owner_type                        = "application"
  owner_id                          = data.sonatypeiq_application.application.id
  base_branch                       = "main"
  remediation_pull_requests_enabled = true
  pull_request_commenting_enabled   = true
  source_control_evaluation_enabled = false
  repository_url                    = "https://github.com/sonatype-nexus-community/terraform-provider-sonatypeiq.git"
}

# Set source control for an organization
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}

resource "sonatypeiq_source_control" "organization" {
  owner_type                        = "organization"
  owner_id                          = data.sonatypeiq_organization.sandbox.id
  remediation_pull_requests_enabled = true
  pull_request_commenting_enabled   = true
  source_control_evaluation_enabled = true
  base_branch                       = "my-cool-branch"
}
