# Create and manage System Configuration for Sonatype IQ Server
resource "sonatypeiq_system_config" "iq_config" {
  base_url       = "https://my-public-iq-server-url"
  force_base_url = false

  # Every property below is optional and only managed when set. A property you remove from this
  # configuration is deleted, so Sonatype IQ Server reverts to its default.
  #
  # Properties are only available from the Sonatype IQ Server release that introduced them (see the
  # table above). Setting one on an older server fails at plan time.
  session_timeout                             = 45 # minutes
  csrf_protection                             = true
  policy_monitoring_hour                      = 2 # 0-23
  user_agent_suffix                           = "Approved"
  frame_ancestors_allowlist                   = ["*.example.com"]
  waived_component_upgrade_monitoring_enabled = true      # release 162+
  api_access_allow_list                       = ["admin"] # release 175+
  success_metrics_stage_id                    = "build"   # release 182+
  integrations_supported_version_count        = 5         # release 197+

  # webhook_secret_passphrase is sensitive and is never read back from Sonatype IQ Server
  # webhook_secret_passphrase = var.webhook_secret
}
