# Create and manage Users for Sonatype IQ Server
resource "sonatypeiq_user" "example_user" {
  username   = "example2"
  password   = "randomthing"
  first_name = "Example"
  last_name  = "User"
  email      = "example@user.tld"
}