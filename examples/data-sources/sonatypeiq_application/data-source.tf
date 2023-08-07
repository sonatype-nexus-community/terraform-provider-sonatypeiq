# Get Application by Internal ID
data "sonatypeiq_application" "app" {
  id = "370bf138ffa0429791b7c269cd8edbb9"
}

# Get Application by Public ID
data "sonatypeiq_application" "app" {
  public_id = "sandbox-application"
}