# Manage OIDC Configuration
# Note: OIDC configuration affects how all users sign in to Sonatype IQ Server.
# Use with caution in shared environments.
resource "sonatypeiq_config_oidc" "sso" {
  client_id             = "iq-server"
  client_secret         = var.oidc_client_secret
  idp_issuer            = "https://idp.example.com"
  idp_authorization_url = "https://idp.example.com/authorize"
  idp_token_url         = "https://idp.example.com/token"

  # OAuth2 / JWT configuration
  oauth2_idp_jwks_url      = "https://idp.example.com/jwks"
  oauth2_idp_jws_algorithm = "RS256"
  oauth2_username_claim    = "preferred_username"
  oauth2_email_claim       = "email"
  oauth2_first_name_claim  = "given_name"
  oauth2_last_name_claim   = "family_name"
  oauth2_groups_claim      = "groups"
}
