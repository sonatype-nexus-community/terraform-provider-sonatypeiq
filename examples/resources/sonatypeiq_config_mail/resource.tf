# Create and manage Mail Configuration for Sonatype IQ Server
resource "sonatypeiq_config_mail" "mail_config" {
  hostname          = "smtp.my-domain.tld"
  port              = 465 # Default is 465 if not specified
  username          = ""
  password          = ""
  ssl_enabled       = false # Default is true if not specified
  start_tls_enabled = true  # Default is true if not specified
  system_email      = "no-reply@my-domain.tld"
}