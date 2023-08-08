# Create and manage System Configuration for Sonatype IQ Server
resource "sonatypeiq_system_config" "iq_config" {
  base_url       = "https://my-public-iq-server-url"
  force_base_url = false
}