provider "sonatypeiq" {
  url      = "https://my-sonatype-iq-server.tld"
  username = "username"
  password = "password"

  # Only needed if a reverse proxy or load balancer in front of Sonatype IQ Server removes or rewrites
  # the Server HTTP header, so the provider cannot work out the version for itself.
  # version_hint = "1.205.4-01"
}
