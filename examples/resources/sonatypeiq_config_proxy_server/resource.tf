# Create and manage Proxy Server Configuration for Sonatype IQ Server
resource "sonatypeiq_config_proxy_server" "proxy_config" {
  hostname = "smtp.my-domain.tld"
  port     = 465 # Default is 465 if not specified
  username = ""
  password = ""
  exclude_hosts = [
    "exluded-host-01.tld",
    "exluded-host-02.tld",
  ]
}