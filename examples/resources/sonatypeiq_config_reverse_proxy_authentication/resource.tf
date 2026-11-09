# Create and manage Reverse Proxy Authentication Configuration for Sonatype IQ Server
#
# WARNING: Enabling reverse proxy authentication changes how users sign in to IQ Server.
# When enabled, users authenticate via a reverse proxy that provides the username in an
# HTTP header. This may affect the account Terraform itself connects as if using the
# same authentication mechanism. Test thoroughly in a non-production environment before
# enabling in production.
#
# Note: For safety, this example shows the configuration as disabled (enabled = false).
# You can prepare the configuration without activating it.
resource "sonatypeiq_config_reverse_proxy_authentication" "reverse_proxy_auth" {
  enabled                  = false                            # Set to true to enable reverse proxy authentication
  username_header          = "X-Forwarded-User"               # The header containing the username (defaults to REMOTE_USER)
  csrf_protection_disabled = false                            # WARNING: Setting to true disables CSRF protection - only use for legacy client compatibility
  logout_url               = "https://sso.example.com/logout" # Optional: URL to redirect to on logout
}
