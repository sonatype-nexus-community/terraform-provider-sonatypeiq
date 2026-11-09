# Create and manage User Token Configuration for Sonatype IQ Server
# Requires Sonatype IQ Server release 199 or later
resource "sonatypeiq_user_token_configuration" "user_token_config" {
  user_token_default_expiration_days = 90
}
